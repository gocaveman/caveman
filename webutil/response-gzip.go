package webutil

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"
)

// TODO: make a GzipResponseWriter and the ChainHandler that creates it,
// use WrapResponseWriter as a base (gzip is a great example of how it's used).

type GzipHandler struct{}

func NewGzipHandler() GzipHandler {
	return GzipHandler{}
}

func (h GzipHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {

	// TODO: we really should properly parse the accept-encoding header, but its kind of a pain; go simple for now

	// Verify client accepts gzip
	if !strings.Contains(r.Header.Get("accept-encoding"), "gzip") {
		return w, r
	}

	w.Header().Set("Content-Encoding", "gzip")

	gzw := gzip.NewWriter(w)

	retw := &GzipResponseWriter{
		ResponseWriter: w,
		gzw:            gzw,
	}
	return retw, r
}

type GzipResponseWriter struct {
	http.ResponseWriter
	gzw     *gzip.Writer
	written bool
}

func (w *GzipResponseWriter) Write(p []byte) (int, error) {

	// TODO: Review if there are some common cases where we do not want to apply gzipping, e.g.
	// - Binary content types possibly don't need it
	// - See if it interferes at all with things like server-sent events
	// - Also websockets, will things work correctly?
	// The logic could be improved here to automatically disable in cases we can detect.

	if !w.written {

		// if no content type, we autodetect it here - because if we don't then the underlying default
		// ResponseWriter will try it on the gzipped bytes which will say that everthing is a gzip file
		if w.Header().Get("content-type") == "" {
			ct := http.DetectContentType(p)
			w.Header().Set("content-type", ct)
		}

		w.WriteHeader(http.StatusOK)

		// Send an empty write through to the underlying writer - to be sure the context cancelation works.
		// w.gzw.Write() below is not guaranteed to reach the underlying response writer in this call.
		w.ResponseWriter.Write(nil)

	}

	return w.gzw.Write(p)
}

func (w *GzipResponseWriter) WriteHeader(c int) {
	w.written = true
	w.ResponseWriter.WriteHeader(c)
}

func (w *GzipResponseWriter) Close() (err error) {
	w.gzw.Close()
	if c, ok := w.ResponseWriter.(io.Closer); ok {
		err = c.Close()
	}
	return err
}

func (w *GzipResponseWriter) Flush() {
	w.gzw.Flush()
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *GzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *GzipResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *GzipResponseWriter) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.(http.Pusher).Push(target, opts)
}
