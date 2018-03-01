package webutil

import "net/http"

// NewHTTPSRedirectHandler creates a new HTTPSRedirectHandler using the defaults.
func NewHTTPSRedirectHandler() *HTTPSRedirectHandler {
	return &HTTPSRedirectHandler{
		IgnorePathPrefixes: DefaultIgnorePathPrefixes,
		HTTPSRedirect:      DefaultHTTPS302RedirectHandler,
	}
}

// DefaultIgnorePathPrefixes are paths that we don't want to redirect to HTTPS.
var DefaultIgnorePathPrefixes = []string{
	"/.well-known", // path prefix needed for LetsEncrypt support
}

// HTTPSRedirectHandler will redirect all requests which do not have r.URL.Scheme=="https".
type HTTPSRedirectHandler struct {
	IgnorePathPrefixes []string     // do not HTTPS redirects for these path prefixes
	HTTPSRedirect      http.Handler // the handler to perform the actual redirect to HTTPS
}

// DefaultHTTPS302RedirectHandler redirects to HTTPS by doing a 302 redirect to https://HOST/PATH
// constructed from the original request, with the port number removed from the host.
var DefaultHTTPS302RedirectHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	u := *r.URL
	u.Host = u.Hostname()
	u.Scheme = "https"
	http.Redirect(w, r, u.String(), 302)
})

// DefaultHTTPS301RedirectHandler redirects to HTTPS by doing a 301 redirect to https://HOST/PATH
// constructed from the original request, with the port number removed from the host.
var DefaultHTTPS301RedirectHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	u := *r.URL
	u.Host = u.Hostname()
	u.Scheme = "https"
	http.Redirect(w, r, u.String(), 302)
})

func (h *HTTPSRedirectHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {

	if r.URL.Scheme == "https" {
		return w, r
	}

	for _, v := range h.IgnorePathPrefixes {
		if HasPathPrefix(r.URL.Path, v) {
			return w, r
		}
	}

	return w, r
}
