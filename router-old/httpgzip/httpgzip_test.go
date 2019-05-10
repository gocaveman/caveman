package httpgzip

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/gocaveman/caveman/router"
	"github.com/stretchr/testify/assert"
)

func TestGzipServer(t *testing.T) {

	assert := assert.New(t)

	// do a full end to end test with a server
	r := router.New()

	// add gzip
	r.Add(New(""))

	r.Add(&router.HTTPRouteHandler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// log.Printf("HERE: %s", w.Header().Get("Content-Encoding"))
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "v: %s", r.URL.Query().Get("v"))
		}),
	})

	r.Add(&router.HTTPRouteHandler{
		Sequence: router.RouteSequenceLast,
		Handler:  http.NotFoundHandler(),
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	u := ts.URL + "/?v=example1"
	log.Printf("URL: %s", u)

	// client := ts.Client()
	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	assert.NoError(err)
	b, _ := httputil.DumpRequestOut(req, true)
	log.Printf("REQUEST: %s", b)

	res, err := client.Do(req)
	// res, err := http.Get(u)
	assert.NoError(err)
	defer res.Body.Close()

	// var buf bytes.Buffer
	// res.Header.Write(&buf)
	// log.Printf("RESPONSE HEADER: %s", buf.String())

	b, err = httputil.DumpResponse(res, true)
	assert.NoError(err)
	log.Printf("RESPONSE (len=%d): %s", len(b), b)
	assert.Contains(string(b), "example1")

	// b, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// fmt.Printf("%s", b)

}
