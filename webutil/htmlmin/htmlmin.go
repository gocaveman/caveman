package htmlmin

import (
	"bufio"
	"io"
	"mime"
	"net"
	"net/http"
	"regexp"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

// NewHandler returns a new HTML minifying handler with the default settings.  For simple use all that is needed is to
// add this early on in your handler list.
func NewHandler() *HTMLMinHandler {
	// we register all this other stuff because the HTML minifier might need it for other content types inline.
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	// some of these defaults are a bit much, tune it back a bit
	m.Add("text/html", &html.Minifier{
		KeepConditionalComments: true,
		KeepDefaultAttrVals:     true,
		KeepDocumentTags:        true,
		KeepEndTags:             true,
	})

	return &HTMLMinHandler{
		M: m,
	}
}

// HTMLMinHandler replaces the ResponseWriter with one that minifies HTML as it's written.
type HTMLMinHandler struct {
	M *minify.M // public in case you want to replace the minifier to change it's behavior
}

func (h *HTMLMinHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {

	retw := &HTMLMinResponseWriter{
		ResponseWriter: w,
		minw:           nil, // starts as nil, gets set on first Write() if content type is text/html
		m:              h.M,
	}
	return retw, r

}

// HTMLMinResponseWriter implements HTML minification on content type text/html written to it.
// Normally you want to use NewHandler to create an HTMLMinHandler which creates a HTMLMinResponseWriter
// for each request.
type HTMLMinResponseWriter struct {
	http.ResponseWriter
	minw io.WriteCloser
	m    *minify.M
}

// writeThrough is used in place of the minifer when minifying not enabled (wrong content type)
type writeThrough struct {
	io.Writer
}

func (wt *writeThrough) Close() error { return nil }

func (w *HTMLMinResponseWriter) Write(p []byte) (int, error) {

	if w.minw == nil {

		// see if content type is set
		ct, _, _ := mime.ParseMediaType(w.Header().Get("content-type"))

		// if not set, detect
		if ct == "" {
			ct, _, _ = mime.ParseMediaType(http.DetectContentType(p))
		}

		// if text/html, create writer
		if ct == "text/html" {
			w.minw = w.m.Writer("text/html", w.ResponseWriter)

			// Send an empty write through to the underlying writer - to be sure the context cancelation works.
			// w.minw.Write() below is not guaranteed to reach the underlying response writer in this call.
			w.ResponseWriter.Write(nil)

		} else {
			w.minw = &writeThrough{Writer: w.ResponseWriter}
		}

	} else {
		w.minw = &writeThrough{Writer: w.ResponseWriter}
	}

	return w.minw.Write(p)
}

func (w *HTMLMinResponseWriter) WriteHeader(c int) {
	w.ResponseWriter.WriteHeader(c)
}

func (w *HTMLMinResponseWriter) Close() (err error) {
	if w.minw != nil {
		w.minw.Close()
		w.minw = nil
	}
	if c, ok := w.ResponseWriter.(io.Closer); ok {
		err = c.Close()
	}
	return err
}

func (w *HTMLMinResponseWriter) Flush() {
	if w.minw != nil {
		w.minw.Close()
		w.minw = nil
	}
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *HTMLMinResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.minw != nil {
		w.minw.Close()
		w.minw = nil
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *HTMLMinResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *HTMLMinResponseWriter) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.(http.Pusher).Push(target, opts)
}
