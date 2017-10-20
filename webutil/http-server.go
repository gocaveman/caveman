package webutil

import (
	"context"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
)

// StartHTTPServer is a very simple helper that starts an http.Server.
// The WaitGroup.Add(1) and Done are called at the appropriate times
// and logging is done.  This is just to help make projects' main.go
// more concise and easier to read.
func StartHTTPServer(s *http.Server, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Printf("Listening for HTTP at %q", s.Addr)
	go func() {
		defer wg.Done()
		err := s.ListenAndServe()
		if err != nil {
			log.Printf("Error from HTTP server: %v", err)
		}
	}()
}

// NewRequestContextHandler returns a handler that sets "http.Request"
// in the context.  Intended for use in HTML templates for things like
// getting URL params.
func NewRequestContextHandler() ChainHandler {
	return ChainHandlerFunc(func(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {
		return w, r.WithContext(context.WithValue(r.Context(), "http.Request", r))
	})
}

// NewStaticFileHandler returns a handler similar to the stdlib FileServer.
// The primary difference is that instead of returning an error if the file
// is not found it will instead not serve anything (so the request will
// fall through to the next handler in the chain).  It also sets the cache-control
// header (only if not already set) to tell the browser to not cache if no
// query string or to cache for one week if query string.  Provides an easy way
// to control which files are cached by the browser with the default to not cache
// them.  This handler will ignore directories and take no action if one is found.
func NewStaticFileHandler(fs http.FileSystem) http.Handler {

	fsh := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean("/" + r.URL.Path)
		f, err := fs.Open(p)
		if err == nil {
			defer f.Close()
			st, err := f.Stat()
			if err != nil {
				HTTPError(w, r, err, "Error inspecting file.", 500)
				return
			}
			if st.IsDir() {
				return
			}

			cc := w.Header().Get("cache-control")
			if cc == "" {
				if r.URL.RawQuery == "" {
					w.Header().Set("cache-control", "no-cache")
				} else {
					w.Header().Set("cache-control", "max-age=604800")
				}
			}

			fsh.ServeHTTP(w, r)
		} else {
			// differentiate here between a file that doesn't exist an something else wrong with the filesystem
			if !os.IsNotExist(err) {
				HTTPError(w, r, err, "Error serving file.", 500)
				return
			}
		}
	})

}
