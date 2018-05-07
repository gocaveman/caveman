package renderer

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gocaveman/caveman/webutil"
)

// TODO: this should probably be what implements the AJAX functionality of making it so we can request just a portion
// of a page by naming it's template section - should probably have a whitelist of sections allowed

// PartNameFunc is a function that can say if a template name is okay to serve as a part.
type PartNameFunc func(string) bool

// RenderHandler implements http Handler and converts requests for pages into
type RenderHandler struct {

	// maps r.URL.Path to a list of possible filenames
	FileNamer FileNamer

	// the renderer to use
	Renderer Renderer `autowire:""`

	// loads page information before rendering, if provided
	PageInfoReader PageInfoReader `autowire:""`

	// function to test against part names
	PartNameFunc PartNameFunc

	// parameter name which specifices the part to serve (if part serving is enabled)
	PartNameParam string
}

// NewHandler returns a RenderHandler with the default FileNamer and the Renderer you provide.
func NewHandler(r Renderer) *RenderHandler {
	return &RenderHandler{
		Renderer:  r,
		FileNamer: NewDefaultFileNamer(),
		PartNameFunc: PartNameFunc(func(name string) bool {
			return name == "body"
		}),
		PartNameParam: "part",
	}
}

func (h *RenderHandler) setHeaders(w http.ResponseWriter) {

	if w.Header().Get("content-type") == "" {
		w.Header().Set("content-type", "text/html")
	}

	if w.Header().Get("cache-control") == "" {
		w.Header().Set("cache-control", "no-store")
	}

}

func (h *RenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Path

	// use either PageInfoReader or FileNamer to figure out what template we're serving
	var fns []string
	if h.PageInfoReader != nil {
		tmplFileName, _, err := h.PageInfoReader.ReadPageInfo(p)
		if err == nil && tmplFileName != "" {
			fns = []string{tmplFileName}
		}
	} else {
		fns = h.FileNamer.FileNames(p)
	}

	renderPartName := ""
	if h.PartNameParam != "" {
		partName := r.URL.Query().Get(h.PartNameParam)
		if partName != "" {
			if h.PartNameFunc(partName) {
				renderPartName = partName
			} else {
				webutil.HTTPError(w, r, nil, fmt.Sprintf("invalid part name %q", partName), 400)
				return
			}
		}
	}

	// a part was requested for rendering
	if renderPartName != "" {

		// try each file name, skipping over files that don't exist
		for _, fn := range fns {
			ctx, t, err := h.Renderer.Parse(r.Context(), fn)
			// err := h.Renderer.ParseAndExecute(r.Context(), fn, w, nil)
			if err == nil {

				h.setHeaders(w)

				err := t.ExecuteTemplate(w, renderPartName, ctx)
				if err != nil {
					webutil.HTTPError(w, r, err, "error during render handler (executing part)", 500)
				}

				break
			} else if os.IsNotExist(err) || err == webutil.ErrNotFound {
				continue
			} else {
				webutil.HTTPError(w, r, err, "error during render handler (parsing for part)", 500)
				break
			}
		}

		return
	}

	// try each file name, skipping over files that don't exist
	for _, fn := range fns {

		ctx, t, err := h.Renderer.Parse(r.Context(), fn)
		// err := h.Renderer.ParseAndExecute(r.Context(), fn, w, nil)
		if err == nil {

			h.setHeaders(w)

			err := t.ExecuteTemplate(w, fn, ctx)
			if err != nil {
				webutil.HTTPError(w, r, err, "error during render handler (executing)", 500)
				break
			}

			break
		} else if os.IsNotExist(err) || err == webutil.ErrNotFound {
			continue
		} else {
			webutil.HTTPError(w, r, err, "error during render handler (parsing)", 500)
			break
		}
	}

}

// PageInfoReader is a subset of pageinfo.Store, which we use to load page meta
type PageInfoReader interface {
	ReadPageInfo(path string) (tmplFileName string, meta map[string]interface{}, err error)
}

// NotFoundHandler will return a handler that renders the specified page with a 404 status code.
// By convention you usually want to pass "/_404.gohtml" as the path.
func NotFoundHandler(rend Renderer, path404 string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		rend.ParseAndExecuteHTTP(w, r, path404)
	})
}
