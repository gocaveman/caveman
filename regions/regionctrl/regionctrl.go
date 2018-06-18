// Region controller for administrative edits to regions.  Not required for rendering.
package regionctrl

import (
	"log"
	"net/http"

	"github.com/gocaveman/caveman/httpapi"
	"github.com/gocaveman/caveman/regions"
)

func NewRegionController(store regions.Store) *RegionController {
	return &RegionController{
		Prefix: "/api/region",
		Store:  store,
	}
}

// RegionController provides an editing API on top of a regions.Store.
// (It is not used as part of normal page rendering.)
type RegionController struct {
	// FIXME: Should we break this down into a prefix and suffix?
	// It would allow us to autowire all of the default caveman stuff to a different prefix -
	// some people might find this very useful for integrating into existing sites/apps.
	Prefix string
	Store  regions.Store `autowire:""`

	// TODO: JSONRPC2, mind as well support it also, we can... although we still have to
	// debug the multiple body parsing issue... possibly each controller gets it's own endpoint??
}

func (h *RegionController) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("FIXME: RegionController needs security (auth/perms)")

	ar := httpapi.NewRequest(r)

	var def regions.Definition

	// TODO: we'll also want to load page data for the admin page
	// TODO: we should also provide the endpoint path/url here to the page, so
	// this page and anything else that needs it will track with a change to the API endpoint

	switch {

	// write
	case ar.ParseRESTObj("POST", &def, h.Prefix):
		if err := def.IsValid(); err != nil {
			ar.WriteError(w, 400, err)
			return
		}
		err := h.Store.WriteDefinition(def)
		if err != nil {
			ar.WriteError(w, 500, err)
			return
		}
		ar.WriteResult(w, 200, def)
		return

	// delete
	case ar.ParseRESTPath("DELETE", h.Prefix+"/%s", &def.DefinitionID):
		err := h.Store.DeleteDefinition(def.DefinitionID)
		if err != nil {
			ar.WriteError(w, 500, err)
			return
		}
		ar.WriteResult(w, 200, def.DefinitionID)
		return

	// list
	case ar.ParseRESTPath("GET", h.Prefix):
		defs, err := h.Store.AllDefinitions()
		if err != nil {
			ar.WriteError(w, 500, err)
			return
		}
		ar.WriteResult(w, 200, defs)
		return

	}

}
