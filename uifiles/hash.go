package uifiles

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"hash"

	"github.com/gocaveman/caveman/webutil"
)

// Hash is a thin wrapper around the hash.Hash returned by sha256.New, with some helper methods to make it easier to use for us.
type Hash struct {
	hash.Hash
}

func NewHash() *Hash {
	return &Hash{
		Hash: sha256.New(),
	}
}

func (h *Hash) Strings(s []string) {
	for _, str := range s {
		h.Hash.Write([]byte(str))
	}
}

func (h *Hash) FileEntryList(files FileEntryList) {
	h.Strings([]string(files))
}

func (h *Hash) DataSourceModTimes(dss []webutil.DataSource) (lasterr error) {
	b := make([]byte, 8)
	for _, ds := range dss {
		// record and return last error but don't let it stop the hashing, caller can choose to heed it or not
		fi, err := ds.Stat()
		if err != nil {
			lasterr = err
		}
		if fi == nil {
			continue
		}
		mt := fi.ModTime().UnixNano()
		binary.BigEndian.PutUint64(b, uint64(mt))
		h.Hash.Write(b)
	}
	return
}

func (h *Hash) ResultString() string {
	outb := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(outb)
}
