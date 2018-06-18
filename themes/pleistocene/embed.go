package pleistocene

import (
	"net/http"
	"path"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/tmpl"
	"github.com/gocaveman/caveman/tmpl/tmplregistry"
)

//go:generate go run embed-gogen.go

func init() {

	// TODO: static stuff

	tmplregistry.MustRegister(tmplregistry.SeqTheme, "pleistocene", NewTmplStore())

}

func NewViewsFS() http.FileSystem {
	return fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return EmbeddedAssets.Open("/views" + path.Clean("/"+name))
	})
}

func NewIncludesFS() http.FileSystem {
	return fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return EmbeddedAssets.Open("/includes" + path.Clean("/"+name))
	})
}

func NewTmplStore() tmpl.Store {
	return &tmpl.HFSStore{
		FileSystems: map[string]http.FileSystem{
			tmpl.ViewsCategory:    NewViewsFS(),
			tmpl.IncludesCategory: NewIncludesFS(),
		},
	}
}
