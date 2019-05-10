package routerold

import (
	"bufio"
	"net"
	"net/http"
)

// WrapResponseWriter implements http.ResponseWriter and allows you to plug in your own functions
// for the various calls.
type WrapResponseWriter interface {
	Parent() http.ResponseWriter

	// http.ResponseWriter methods
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)

	// assign functions to handle ResponseWriter methods
	SetHeaderFunc(func() http.Header)
	SetWriteFunc(func([]byte) (int, error))
	SetWriteHeaderFunc(func(statusCode int))

	// override additional optional ResponseWriter methods, returns true if supported
	SetHijackFunc(func() (net.Conn, *bufio.ReadWriter, error)) bool
	SetPushFunc(func(target string, opts *http.PushOptions) error) bool
	SetFlushFunc(func()) bool
}

// NewWrapResponseWriter will create the appropriate WrapResponseWriter for HTTP/1.1 or HTTP/2
// and return it.
func NewWrapResponseWriter(w http.ResponseWriter, r *http.Request) WrapResponseWriter {

	if r.ProtoMajor == 2 {
		return &WrapHTTP2ResponseWriter{parent: w}
	}

	return &WrapHTTP1ResponseWriter{parent: w}
}

// WrapHTTP1ResponseWriter implements and wraps an http.ResponseWriter with HTTP/1.1 methods
// implemented. Hijack() is suppported, Push() is not.
type WrapHTTP1ResponseWriter struct {
	parent http.ResponseWriter

	headerFunc      func() http.Header
	writeFunc       func([]byte) (int, error)
	writeHeaderFunc func(statusCode int)

	hijackFunc func() (net.Conn, *bufio.ReadWriter, error)
	flushFunc  func()
}

func (w *WrapHTTP1ResponseWriter) Parent() http.ResponseWriter {
	return w.parent
}

func (w *WrapHTTP1ResponseWriter) Header() http.Header {
	if w.headerFunc != nil {
		return w.headerFunc()
	}
	return w.Parent().Header()
}

func (w *WrapHTTP1ResponseWriter) Write(b []byte) (int, error) {
	if w.writeFunc != nil {
		return w.writeFunc(b)
	}
	return w.Parent().Write(b)
}

func (w *WrapHTTP1ResponseWriter) WriteHeader(statusCode int) {
	if w.writeHeaderFunc != nil {
		w.writeHeaderFunc(statusCode)
		return
	}
	w.Parent().WriteHeader(statusCode)
}

func (w *WrapHTTP1ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.hijackFunc != nil {
		return w.hijackFunc()
	}
	return w.Parent().(http.Hijacker).Hijack()
}

func (w *WrapHTTP1ResponseWriter) Flush() {
	if w.flushFunc != nil {
		w.flushFunc()
		return
	}
	w.Parent().(http.Flusher).Flush()
}

func (w *WrapHTTP1ResponseWriter) SetHeaderFunc(f func() http.Header) {
	w.headerFunc = f
}
func (w *WrapHTTP1ResponseWriter) SetWriteFunc(f func([]byte) (int, error)) {
	w.writeFunc = f
}
func (w *WrapHTTP1ResponseWriter) SetWriteHeaderFunc(f func(statusCode int)) {
	w.writeHeaderFunc = f
}

func (w *WrapHTTP1ResponseWriter) SetHijackFunc(f func() (net.Conn, *bufio.ReadWriter, error)) bool {
	w.hijackFunc = f
	return true
}

func (w *WrapHTTP1ResponseWriter) SetPushFunc(f func(target string, opts *http.PushOptions) error) bool {
	return false // not supported for HTTP/1.1
}

func (w *WrapHTTP1ResponseWriter) SetFlushFunc(f func()) bool {
	w.flushFunc = f
	return true
}

// WrapHTTP2ResponseWriter implements and wraps an http.ResponseWriter with HTTP/2 methods
// implemented. Push() is suppported, Hijack() is not.
type WrapHTTP2ResponseWriter struct {
	parent http.ResponseWriter

	headerFunc      func() http.Header
	writeFunc       func([]byte) (int, error)
	writeHeaderFunc func(statusCode int)

	pushFunc  func(target string, opts *http.PushOptions) error
	flushFunc func()
}

func (w *WrapHTTP2ResponseWriter) Parent() http.ResponseWriter {
	return w.parent
}

func (w *WrapHTTP2ResponseWriter) Header() http.Header {
	if w.headerFunc != nil {
		return w.headerFunc()
	}
	return w.Parent().Header()
}

func (w *WrapHTTP2ResponseWriter) Write(b []byte) (int, error) {
	if w.writeFunc != nil {
		return w.writeFunc(b)
	}
	return w.Parent().Write(b)
}

func (w *WrapHTTP2ResponseWriter) WriteHeader(statusCode int) {
	if w.writeHeaderFunc != nil {
		w.writeHeaderFunc(statusCode)
		return
	}
	w.Parent().WriteHeader(statusCode)
}

func (w *WrapHTTP2ResponseWriter) Push(target string, opts *http.PushOptions) error {
	if w.pushFunc != nil {
		return w.pushFunc(target, opts)
	}
	return w.Parent().(http.Pusher).Push(target, opts)
}

func (w *WrapHTTP2ResponseWriter) Flush() {
	if w.flushFunc != nil {
		w.flushFunc()
		return
	}
	w.Parent().(http.Flusher).Flush()
}

func (w *WrapHTTP2ResponseWriter) SetHeaderFunc(f func() http.Header) {
	w.headerFunc = f
}
func (w *WrapHTTP2ResponseWriter) SetWriteFunc(f func([]byte) (int, error)) {
	w.writeFunc = f
}
func (w *WrapHTTP2ResponseWriter) SetWriteHeaderFunc(f func(statusCode int)) {
	w.writeHeaderFunc = f
}

func (w *WrapHTTP2ResponseWriter) SetHijackFunc(f func() (net.Conn, *bufio.ReadWriter, error)) bool {
	return false // not supported for HTTP/2
}

func (w *WrapHTTP2ResponseWriter) SetPushFunc(f func(target string, opts *http.PushOptions) error) bool {
	w.pushFunc = f
	return true
}

func (w *WrapHTTP2ResponseWriter) SetFlushFunc(f func()) bool {
	w.flushFunc = f
	return true
}
