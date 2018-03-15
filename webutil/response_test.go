package webutil

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func TestResponse(t *testing.T) {

	hl := NewDefaultHandlerList(
		NewCtxSetHandler("webutil.TestResponse", "test_value"),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/test1" {
				fmt.Fprintf(w, "<html><body>testing! %q</body></html>", r.Context().Value("webutil.TestResponse"))
			}
		}),
		http.NotFoundHandler(),
	)

	// var hl HandlerList
	// hl = append(hl, NewContextCancelHandler())
	// hl = append(hl, NewGzipHandler())
	// hl = append(hl, NewCtxSetHandler("webutil.TestResponse", "test_value"))
	// hl = append(hl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if r.URL.Path == "/test1" {
	// 		fmt.Fprintf(w, "<html><body>testing! %q</body></html>", r.Context().Value("webutil.TestResponse"))
	// 	}
	// }))
	// hl = append(hl, http.NotFoundHandler())

	r := httptest.NewRequest("GET", "/test1", nil)
	r.Header.Set("accept-encoding", "gzip")
	tw := httptest.NewRecorder()
	var w http.ResponseWriter = tw

	w, r = hl.ServeHTTPChain(w, r)

	result := tw.Result()
	b, _ := httputil.DumpResponse(result, false)
	log.Printf("RESULT:\n%s", b)

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}
	// log.Printf("result body %v", body)

	if result.Header.Get("content-encoding") == "gzip" {
		gr, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		body, err = ioutil.ReadAll(gr)
		if err != nil && err != io.ErrUnexpectedEOF { // ignore unexpected EOF, just verify the stream below got to the end
			t.Fatal(err)
		}
	}

	if !bytes.Contains(body, []byte("test_value")) {
		t.Fatalf("result did not contain test_value")
	}

	if !bytes.Contains(body, []byte("</html>")) {
		t.Fatalf("result did not contain </html>")
	}

	log.Printf("RESULT BODY:\n%s", body)

}
