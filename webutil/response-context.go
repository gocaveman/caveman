package webutil

import (
	"bufio"
	"context"
	"net"
	"net/http"
)

// TODO: a struct (WrapResponseWriter) that cleanly wraps a ResponseWriter
// (can be embedded easily by another struct to add functionality)

// TODO: a struct that embeds WrapResponseWriter to implement context cancellation
// when WriteHeader is called.  Name: CancelingResponseWriter

// TODO: we could also add one here that dumps everything to the log, lower priority
// but probably useful (it should be smart enough to ungzip what GzipResponseWriter
// has done in order to make it human readable).

// DummyResponseWriterCloser implements only the context cancellation and otherwise does nothing.

type ContextCancelHandler struct{}

func NewContextCancelHandler() ContextCancelHandler {
	return ContextCancelHandler{}
}

func (h ContextCancelHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	ctx, cancelFunc := context.WithCancel(r.Context())
	retw := &ContextCancelResponseWriter{
		ResponseWriter: w,
		cancelFunc:     cancelFunc,
	}
	return retw, r.WithContext(ctx)
}

type ContextCancelResponseWriter struct {
	http.ResponseWriter
	cancelFunc context.CancelFunc
}

func (w *ContextCancelResponseWriter) Write(p []byte) (int, error) {
	// TODO: Either as an option here or in a separate handler,
	// we should have something that can dump out the stack trace; as
	// a debug tool so we can see what handler ended up writing.
	w.cancelFunc()
	return w.ResponseWriter.Write(p)
}

func (w *ContextCancelResponseWriter) WriteHeader(c int) {
	w.cancelFunc()
	w.ResponseWriter.WriteHeader(c)
}

// func (w *ContextCancelResponseWriter) Close() (err error) {
// 	log.Printf("FIXME: Close() is probably not a good idea - using Flush() instead of Close keeps to the existing API and can serve the same purpose")
// 	w.cancelFunc()
// 	if c, ok := w.ResponseWriter.(io.Closer); ok {
// 		err = c.Close()
// 	}
// 	return err
// }

// Flush cancels the context and calls Flush on the underlying ResponseWriter.
func (w *ContextCancelResponseWriter) Flush() {
	w.cancelFunc()
	w.ResponseWriter.(http.Flusher).Flush()
}

// Hijack cancels the context and calls Hijack on the underlying ResponseWriter.
func (w *ContextCancelResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.cancelFunc()
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *ContextCancelResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *ContextCancelResponseWriter) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.(http.Pusher).Push(target, opts)
}
