package htmlmin

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/gocaveman/caveman/webutil"
)

func TestHTMLMin(t *testing.T) {

	var hl webutil.HandlerList

	hl = append(hl, webutil.NewContextCancelHandler())
	hl = append(hl, NewHandler())
	hl = append(hl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, "<!doctype html>\n\n<html>\n\n<body>\n\n<div>\n\nTesting      123</div></body></html>")
	}))

	r := httptest.NewRequest("GET", "/test1", nil)
	tw := httptest.NewRecorder()
	var w http.ResponseWriter = tw

	w, r = hl.ServeHTTPChain(w, r)
	w.(io.Closer).Close()

	result := tw.Result()
	b, _ := httputil.DumpResponse(result, true)
	log.Printf("RESULT:\n%s", b)

	if !bytes.Contains(b, []byte(`<!doctype html><html><body><div>Testing 123</div></body></html>`)) {
		t.Fatalf("unexpected output: %s", b)
	}

}
