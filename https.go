package caveman

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

// LoadKeyPairAuto looks at the parameters and loads the certificate(s) and key
// either from the path pointed to on disk or directly as PEM data.
// IsPEMKeyData and IsPEMCertData to detect key and cert data in PEM format.
// This is a simple wrapper around tls.X509KeyPair.
func LoadKeyPairAuto(cert, key string) (tls.Certificate, error) {

	var err error
	certB := []byte(cert)
	keyB := []byte(key)

	if !IsPEMCertData(certB) {
		certB, err = ioutil.ReadFile(cert)
		if err != nil {
			return tls.Certificate{}, err
		}
	}
	if !IsPEMKeyData(keyB) {
		keyB, err = ioutil.ReadFile(key)
		if err != nil {
			return tls.Certificate{}, err
		}
	}

	return tls.X509KeyPair(certB, keyB)

}

// IsPEMKeyData returns true if v contains "-----BEGIN", "PRIVATE KEY-----" and at least one newline.
func IsPEMKeyData(v []byte) bool {
	return bytes.Contains(v, []byte("-----BEGIN")) &&
		bytes.Contains(v, []byte("PRIVATE KEY-----")) &&
		bytes.Contains(v, []byte("\n"))
}

// IsPEMCertData returns true if v contains "-----BEGIN", "CERTIFICATE-----" and at least one newline.
func IsPEMCertData(v []byte) bool {
	return bytes.Contains(v, []byte("-----BEGIN")) &&
		bytes.Contains(v, []byte("CERTIFICATE-----")) &&
		bytes.Contains(v, []byte("\n"))
}

func CtxIsHTTPS(ctx context.Context) bool {
	ret, _ := ctx.Value("https").(bool)
	return ret
}

func CtxWithHTTPS(ctx context.Context) context.Context {
	return context.WithValue(ctx, "https", true)
}

func CtxWithHTTPSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(CtxWithHTTPS(r.Context())))
	})
}
