package webutil

import (
	"fmt"
	"net/http"
)

// NewBasicAuthOneHandler creates a new BasicAuthHandler using the defaults.
func NewBasicAuthOneHandler(username, password string) *BasicAuthHandler {
	return &BasicAuthHandler{
		IncludePathPrefixes: []string{"/"},
		ExcludePathPrefixes: []string{"/.well-known"},
		CheckLogin:          NewCheckLoginOneFunc(username, password),
		Realm:               "restricted access",
	}
}

// BasicAuthHandler will redirect all requests which do not have r.URL.Scheme=="https".
type BasicAuthHandler struct {
	IncludePathPrefixes []string                             // include these path prefixes
	ExcludePathPrefixes []string                             // exclude these path prefixes
	CheckLogin          func(username, password string) bool // check usernamd and password for validity
	Realm               string
}

// NewCheckLoginOneFunc returns a function that checks for the validity of a username and password.
func NewCheckLoginOneFunc(username, password string) func(username, password string) bool {
	return func(checkUsername, checkPassword string) bool {
		return username == checkUsername && password == checkPassword
	}
}

func (h *BasicAuthHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {

	check := false

	for _, k := range h.IncludePathPrefixes {
		if HasPathPrefix(r.URL.Path, k) {
			check = true
			break
		}
	}

	for _, k := range h.ExcludePathPrefixes {
		if HasPathPrefix(r.URL.Path, k) {
			check = false
			break
		}
	}

	if check {
		u, p, ok := r.BasicAuth()
		if ok {
			if h.CheckLogin(u, p) {
				return w, r // login matches, pass through
			}
		}

		// check required but logic failed, send back 401
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", h.Realm))
		http.Error(w, "Unauthorized", 401)
	}

	return w, r
}
