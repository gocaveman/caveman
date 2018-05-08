package paleolithic

import (
	"net/http"
	"path"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/tmpl"
	"github.com/gocaveman/caveman/tmpl/tmplregistry"
)

//go:generate go run embed-gogen.go

func init() {

	baseFS := EmbeddedAssets
	viewsFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return baseFS.Open("/views" + path.Clean("/"+name))
	})
	includesFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return baseFS.Open("/includes" + path.Clean("/"+name))
	})

	s := &tmpl.HFSStore{
		FileSystems: map[string]http.FileSystem{
			tmpl.ViewsCategory:    viewsFS,
			tmpl.IncludesCategory: includesFS,
		},
	}

	// TODO: static stuff

	tmplregistry.MustRegister(tmplregistry.SeqTheme, "embed-gogen.go", s)

}
