package httpgzip

import (
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"

	"github.com/gocaveman/caveman/router"
)

// New will return a new GZIPChainHandler ready to compress responses using gzip.
// By default, all responses matching pathPrefix (empty to mean all) will be
// compressed.  Individual responses can opt out of compression by setting the response
// "Content-Encoding" header to something other than "gzip".
func New(pathPrefix string) *GZIPChainHandler {
	return &GZIPChainHandler{
		PathPrefix: pathPrefix,
		Sequence:   router.RouteSequenceSetup,
	}
}

type GZIPChainHandler struct {
	PathPrefix string
	Sequence   float64
}

func (h *GZIPChainHandler) RoutePathPrefix() string {
	return h.PathPrefix
}

func (h *GZIPChainHandler) RouteSequence() float64 {
	return h.Sequence
}

func (h *GZIPChainHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {

	// imperfect but simple and works for all practical cases
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return w, r
	}

	wwrap := router.NewWrapResponseWriter(w, r)

	// indicate gzip response
	wwrap.Header().Set("Content-Encoding", "gzip")
	// log.Printf("wwrap.Header(): %v", w.Header())

	gw := gzip.NewWriter(w)
	gwWritten := false

	wwrap.SetWriteFunc(func(b []byte) (int, error) {
		if gwWritten && gw != nil { // if we've been gzipping, keep going
			return gw.Write(b)
		}
		// see if header is set up for it and gw hasn't been nuked
		if gw != nil && wwrap.Header().Get("Content-Encoding") == "gzip" {
			// start gzip writing
			gwWritten = true
			return gw.Write(b)
		}
		// nop, looks like they opted out, do a regular write
		return wwrap.Parent().Write(b)
	})

	wwrap.SetFlushFunc(func() {
		if gwWritten && gw != nil { // gzip flush if gw hasn't been nuked and some data was written
			gw.Flush()
		}
		wwrap.Parent().(http.Flusher).Flush()
	})

	wwrap.SetHijackFunc(func() (net.Conn, *bufio.ReadWriter, error) {
		gw = nil // disable any further gzip stuff
		return wwrap.Parent().(http.Hijacker).Hijack()
	})

	wnext, rnext = wwrap, r

	// set deferred handler so we can flush the gzip stream if any data was written
	rnext = rnext.WithContext(router.DeferChainHandler(rnext.Context(), router.ChainHandlerFunc(func(w http.ResponseWriter, r *http.Request) (wret http.ResponseWriter, rret *http.Request) {
		wret, rret = w, r
		if gwWritten && gw != nil {
			// log.Printf("flushing...")
			w.(http.Flusher).Flush() // calling Flush will trigger the gzip flush
			gw.Close()               // then closer the gzip writer
			// gw.Reset(w)              // note totally sure on this, but this would mean that further writes will start a new gzip stream...
		}
		return
	})))

	return
}
