// An example theme that shows how to implement one.
package demotheme

import (
	"net/http"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/renderer/includeregistry"
	"github.com/gocaveman/caveman/renderer/viewregistry"
)

//go:generate go run assets_generate.go

// need to implement includeregistry
// and get binary packaging figured out

func init() {

	includesFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return assets.Open("includes/" + name)
	})

	viewsFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return assets.Open("views/" + name)
	})

	includeregistry.Register(100, "demotheme", includesFS)
	viewregistry.Register(100, "demotheme", viewsFS)

}
