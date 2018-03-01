// An example theme that shows how to implement one.
package demotheme

import (
	"net/http"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/renderer/includeregistry"
	"github.com/gocaveman/caveman/renderer/viewregistry"
	"github.com/gocaveman/caveman/webutil/staticregistry"

	_ "github.com/gocaveman-libs/bootstrap4css"
)

//go:generate go run assets_generate.go

// TODO: get binary packaging fully figured out

func init() {

	includesFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return assets.Open("includes/" + name)
	})

	viewsFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return assets.Open("views/" + name)
	})

	staticFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return assets.Open("static/" + name)
	})

	includeregistry.MustRegister(100, "demotheme", includesFS)
	viewregistry.MustRegister(100, "demotheme", viewsFS)
	staticregistry.MustRegister(100, "demotheme", staticFS)

}
