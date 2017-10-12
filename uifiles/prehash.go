package uifiles

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/gocaveman/caveman/webutil"
)

// PrehashEntry is one record in a PrehashData
type PrehashEntry struct {
	Name    string
	ModTime time.Time
}

// PrehashData is a list of entries that describe required files and their timestamps; to be used in prehash computation.
type PrehashData struct {
	Entries []PrehashEntry
}

// Prehash computes the prehash.  The prehash is computer before combining and minification is done, and it's computed
// solely on names and timestamps.  It's purpose is to be used in a cache entry that maps this easy/fast to get metadata
// to an existing computed hash (the full hash requires performing comining and minification, which generally takes much
// longer than computing the prehash).
func (pd *PrehashData) Prehash() string {

	// algo is simple: just put through the md5 each name and then, if not zero,
	// the mod time as a returned by UnixNano and converted to a string ("%d")

	h := md5.New()

	for _, e := range pd.Entries {
		fmt.Fprint(h, e.Name)
		if !e.ModTime.IsZero() {
			fmt.Fprint(h, strconv.FormatInt(e.ModTime.UnixNano(), 10))
		}
	}

	b := h.Sum(nil)

	return hex.EncodeToString(b)
}

func (pd PrehashData) Equal(pd2 PrehashData) bool {

	if len(pd.Entries) != len(pd2.Entries) {
		return false
	}

	for i := 0; i < len(pd.Entries); i++ {
		if pd.Entries[i].Name != pd2.Entries[i].Name {
			return false
		}
		if pd.Entries[i].ModTime.UnixNano() != pd2.Entries[i].ModTime.UnixNano() {
			return false
		}
	}

	return true

}

func (pd PrehashData) EqualIgnoreModTime(pd2 PrehashData) bool {

	if len(pd.Entries) != len(pd2.Entries) {
		return false
	}

	for i := 0; i < len(pd.Entries); i++ {
		if pd.Entries[i].Name != pd2.Entries[i].Name {
			return false
		}
	}

	return true

}

func (pd PrehashData) EncodeTextLines() (data []byte, err error) {
	var buf bytes.Buffer
	for _, e := range pd.Entries {
		fmt.Fprintf(&buf, "%s\t%d\n", e.Name, e.ModTime.UnixNano())
	}
	return buf.Bytes(), nil
}

func (pd *PrehashData) DecodeTextLines(data []byte) (err error) {

	lines := bytes.Split(data, []byte("\n"))
	pd.Entries = make([]PrehashEntry, 0, len(lines))
	for _, line := range lines {

		// skip blank lines (also incidentally eats encryption zero padding)
		line = bytes.Trim(line, "\x00\n")
		if len(line) == 0 {
			continue
		}

		parts := bytes.Split(line, []byte("\t"))
		if len(parts) < 2 {
			return fmt.Errorf("invalid line: %q", line)
		}
		var nsec int64
		fmt.Sscanf(string(parts[1]), "%d", &nsec)
		t := time.Unix(0, nsec)
		pd.Entries = append(pd.Entries, PrehashEntry{
			Name:    string(parts[0]),
			ModTime: t,
		})
	}

	return nil
}

// EncodeToken returns a string which is intended to be put in a URL so the PrehashData can be recreated from it.
// It is not human readable but is URL-safe and can be used as-is in a URL param or path.
// The key is used to encrypt the token (and allows/requires decryption during decoding.)
// To be clear: The intention behind encrypting is not to conceal the data contained within (although
// in the author's opinion obfuscating it is advantageous), but to make it difficult to generate a
// valid token that did not originate from this application (because there are various possible
// attacks that might open up).
func (pd PrehashData) EncodeToken(key []byte) (token string, err error) {

	// convert to plain text
	plainTextBytes, err := pd.EncodeTextLines()
	if err != nil {
		return "", err
	}

	// calculate a checksum
	sum1 := sha256.Sum256(plainTextBytes)
	plainTextHash := sum1[:]

	// use the gzipped version if it's shorter, otherwise it stays as-is
	plainTextBytes = gzipBytesIfShorter(plainTextBytes)

	// sha256 to get from whatever is provided to a 256-bit key
	sum2 := sha256.Sum256(key)
	aesKey := sum2[:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	// Use the plain text hash as the IV - this makes the output deterministic.  Yes if we cared about
	// security this would be bad but we don't.  We are not trying to hide the contents with cryptographic
	// security, we are trying to obfuscate it and be able to valid that it is correct.  AES happens to be
	// a reasonably fast and reliable way to do so.  It is more important that this operation is deterministic
	// with the same key than it is that it's resistent to attack.
	iv := make([]byte, block.BlockSize())
	copy(iv, plainTextHash[:block.BlockSize()])
	cfb := cipher.NewCFBEncrypter(block, iv)

	encBytes := make([]byte, len(plainTextBytes))
	cfb.XORKeyStream(encBytes, plainTextBytes)

	return base64.RawURLEncoding.EncodeToString(plainTextHash) + "," +
		base64.RawURLEncoding.EncodeToString(encBytes), nil
}

// DecodeToken is the reverse operation of EncodeToken.  The data is decrypted with the specified
// key and will error if the key does not match.
func (pd *PrehashData) DecodeToken(key []byte, token string) (reterr error) {

	// Just in case, stranger shit has happened...
	// (I've seen some inputs that cause decryption to panic, maybe I imagined it or it was a case that doesn't
	// apply here, but regardless I think it's better to have this than not.)
	defer func() {
		if r := recover(); r != nil {
			reterr = fmt.Errorf("DecodeToken caught panic: %v", r)
		}
	}()

	tokenParts := strings.Split(token, ",")
	if len(tokenParts) < 2 {
		return fmt.Errorf("does not appear to be valid token")
	}

	plainTextHash, err := base64.RawURLEncoding.DecodeString(tokenParts[0])
	if err != nil {
		return err
	}

	cipherText, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return err
	}

	// sha256 to get from whatever is provided to a 256-bit key
	sum1 := sha256.Sum256(key)
	aesKey := sum1[:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	copy(iv, plainTextHash[:block.BlockSize()])
	cfb := cipher.NewCFBDecrypter(block, iv)
	plainTextBytes := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainTextBytes, cipherText)

	plainTextBytes = gunzipBytesIfNeeded(plainTextBytes)

	sum2 := sha256.Sum256(plainTextBytes)
	plainTextHash2 := sum2[:]

	if bytes.Compare(plainTextHash, plainTextHash2) != 0 {
		return ErrInvalidToken
	}

	return pd.DecodeTextLines(plainTextBytes)

}

// EncodeToken returns a string which is intended to be put in a URL so the PrehashData can be recreated from it.
// It is not human readable but is URL-safe and can be used as-is in a URL param or path.
// The key is used to encrypt the token (and allows/requires decryption during decoding.)
// To be clear: The intention behind encrypting is not to conceal the data contained within (although
// in the author's opinion obfuscating it is advantageous), but to make it difficult to generate a
// valid token that did not originate from this application (because there are various possible
// attacks that might open up).
func EncodeToken(v interface{}, key []byte) (token string, err error) {

	plainTextBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// // convert to plain text
	// plainTextBytes, err := pd.EncodeTextLines()
	// if err != nil {
	// 	return "", err
	// }

	// calculate a checksum
	sum1 := sha256.Sum256(plainTextBytes)
	plainTextHash := sum1[:]

	// use the gzipped version if it's shorter, otherwise it stays as-is
	plainTextBytes = gzipBytesIfShorter(plainTextBytes)

	// sha256 to get from whatever is provided to a 256-bit key
	sum2 := sha256.Sum256(key)
	aesKey := sum2[:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	// Use the plain text hash as the IV - this makes the output deterministic.  Yes if we cared about
	// security this would be bad but we don't.  We are not trying to hide the contents with cryptographic
	// security, we are trying to obfuscate it and be able to valid that it is correct.  AES happens to be
	// a reasonably fast and reliable way to do so.  It is more important that this operation is deterministic
	// with the same key than it is that it's resistent to attack.
	iv := make([]byte, block.BlockSize())
	copy(iv, plainTextHash[:block.BlockSize()])
	cfb := cipher.NewCFBEncrypter(block, iv)

	encBytes := make([]byte, len(plainTextBytes))
	cfb.XORKeyStream(encBytes, plainTextBytes)

	return base64.RawURLEncoding.EncodeToString(plainTextHash) + "," +
		base64.RawURLEncoding.EncodeToString(encBytes), nil
}

// DecodeToken is the reverse operation of EncodeToken.  The data is decrypted with the specified
// key and will error if the key does not match.
func DecodeToken(v interface{}, key []byte, token string) (reterr error) {

	// Just in case, stranger shit has happened...
	// (I've seen some inputs that cause decryption to panic, maybe I imagined it or it was a case that doesn't
	// apply here, but regardless I think it's better to have this than not.)
	defer func() {
		if r := recover(); r != nil {
			reterr = fmt.Errorf("DecodeToken caught panic: %v", r)
		}
	}()

	tokenParts := strings.Split(token, ",")
	if len(tokenParts) < 2 {
		return fmt.Errorf("does not appear to be valid token")
	}

	plainTextHash, err := base64.RawURLEncoding.DecodeString(tokenParts[0])
	if err != nil {
		return err
	}

	cipherText, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return err
	}

	// sha256 to get from whatever is provided to a 256-bit key
	sum1 := sha256.Sum256(key)
	aesKey := sum1[:]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	copy(iv, plainTextHash[:block.BlockSize()])
	cfb := cipher.NewCFBDecrypter(block, iv)
	plainTextBytes := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainTextBytes, cipherText)

	plainTextBytes = gunzipBytesIfNeeded(plainTextBytes)

	sum2 := sha256.Sum256(plainTextBytes)
	plainTextHash2 := sum2[:]

	if bytes.Compare(plainTextHash, plainTextHash2) != 0 {
		return ErrInvalidToken
	}

	// return pd.DecodeTextLines(plainTextBytes)

	return json.Unmarshal(plainTextBytes, v)

}

var ErrInvalidToken = errors.New("invalid token")

func gzipBytesIfShorter(b []byte) []byte {

	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		log.Printf("gzipBytesIfShorter NewWriterLevel returned: %v", err)
		return b
	}
	io.Copy(w, bytes.NewReader(b))
	w.Close()

	newb := buf.Bytes()

	// use it if it's shorter
	if len(newb) < len(b) {
		b = newb
	}

	return b
}

func gunzipBytesIfNeeded(b []byte) []byte {

	if len(b) < 2 {
		return b
	}

	// look for gzip magic number, bail if not there
	if !(b[0] == 0x1f && b[1] == 0x8b) {
		return b
	}

	r, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		log.Printf("gunzipBytesIfNeeded error creating gzip reader: %v", err)
		return b
	}

	retb, err := ioutil.ReadAll(r)
	if err != nil {
		log.Printf("gunzipBytesIfNeeded error reading gzip: %v", err)
		return b
	}

	return retb
}
