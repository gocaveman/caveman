package router

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRouter(t *testing.T) {

	hs := New()

	hs.Add(&DefaultRouteHandler{
		PathPrefix: "/whatever",
		Sequence:   RouteSequenceHandler,
		ChainHandler: ChainHandlerFunc(func(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {
			wnext, rnext = w, r
			fmt.Fprintf(w, "PATH: %s", r.URL.Path)
			return
		}),
	})

}
