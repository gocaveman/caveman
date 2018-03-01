package adminpanel

import (
	"context"
	"net/http"

	"github.com/gocaveman/caveman/renderer"
)

func NewAdminPanelUIHandler(entryList EntryList) *AdminPanelUIHandler {
	return &AdminPanelUIHandler{
		Path:      "/admin/panel",
		EntryList: entryList,
	}
}

type AdminPanelUIHandler struct {
	Path      string
	EntryList EntryList
	Renderer  renderer.Renderer // TODO: autowire
}

func (h *AdminPanelUIHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wret http.ResponseWriter, rret *http.Request) {

	if r.URL.Path == h.Path {

		h.Renderer.ParseAndExecuteHTTP(w,
			r.WithContext(context.WithValue(r.Context(), "EntryList", h.EntryList)),
			"/admin/_panel.html")

	}

	return w, r
}

const VIEWS_ADMIN_PANEL = `
{{template "admin-page.html" .}}

{{define "body"}}
this is a test
{{end}}
`

// aferofs

const INCLUDES_ADMIN_PANEL_NAV = `
<div class="">
includes admin panel here
</div>
`
