package htmlmin

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path"
	"testing"

	"github.com/gocaveman/caveman/webutil"
)

func TestHTMLMin(t *testing.T) {

	var hl webutil.HandlerList

	hl = append(hl, webutil.NewContextCancelHandler())
	hl = append(hl, NewHandler())
	hl = append(hl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Ext(r.URL.Path) == ".html" {
			w.Header().Set("content-type", "text/html")
			fmt.Fprintf(w, "<!doctype html>\n\n<html>\n\n<body>\n\n<div>\n\nTesting      123</div></body></html>")
		} else if path.Ext(r.URL.Path) == ".css" {
			w.Header().Set("content-type", "text/css")
			fmt.Fprintf(w, ".test1 { display: block\n\n}")
		}
	}))

	r := httptest.NewRequest("GET", "/test1.html", nil)
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

	// write something not text/html and make sure all is well
	r = httptest.NewRequest("GET", "/test1.css", nil)
	tw = httptest.NewRecorder()
	w = tw

	w, r = hl.ServeHTTPChain(w, r)
	w.(io.Closer).Close()

	result = tw.Result()
	b, _ = httputil.DumpResponse(result, true)
	log.Printf("RESULT:\n%s", b)

	if !bytes.Contains(b, []byte(".test1 { display: block\n\n}")) {
		t.Fatalf("unexpected output: %s", b)
	}

}
