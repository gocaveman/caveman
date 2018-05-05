// Registry for views to automatically be made available on the site.
package viewregistry

import (
	"log"
	"net/http"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/webutil"
)

// OnlyReadableFromMain determines if `Contents()` can be called from packages other than main.
// By default we prevent this in order to enforce that object wiring is done in the main package
// and not from other components.  The intention is to prevent hidden code dependencies from
// being introduced.
// (TODO: link to doc which explains why in more detail.)
var OnlyReadableFromMain = true

var reg webutil.NamedSequence

// MustRegister adds a new FileSystem to the view registry.  Duplicates (seq, name or value) are not detected or prevented.
func MustRegister(seq float64, name string, viewFS http.FileSystem) {
	log.Printf("viewregistry is deprecated, use tmplregistry instead")
	reg = append(reg, webutil.NamedSequenceItem{Sequence: seq, Name: name, Value: viewFS})
}

// Contents returns the current contents of the registry as a NamedSequence.
func Contents() webutil.NamedSequence {
	if OnlyReadableFromMain {
		webutil.MainOnly(1)
	}
	return reg.SortedCopy()
}

// MakeFS sorts the NamedSequence you provide and returns a FileSystem which is
// the combination of the Values in this list.   Will panic if any value is
// not a http.FileSystem. If overrideFS is not nil, it will be included as the
// first FileSystem in the stack and will take priority over the rest.
// If debug is true then opened files will output debug info to the log.
func MakeFS(overrideFS http.FileSystem, contents webutil.NamedSequence, debug bool) http.FileSystem {
	l := contents.SortedCopy()

	fslist := make([]http.FileSystem, 0, len(l))
	if overrideFS != nil {

		fs := overrideFS

		if debug {
			fs = fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
				f, err := overrideFS.Open(name)
				if f != nil {
					log.Printf("View file opened from override: %s", name)
				}
				return f, err
			})
		}

		fslist = append(fslist, fs)
	}

	for i := range l {

		item := l[i]

		itemFS := item.Value.(http.FileSystem)
		fs := itemFS

		if debug {
			fs = fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
				f, err := itemFS.Open(name)
				if f != nil {
					log.Printf("View file opened (sequence=%v, name=%q): %s", item.Sequence, item.Name, name)
				}
				return f, err
			})
		}

		fslist = append(fslist, fs)
	}

	return fsutil.NewHTTPStackedFileSystem(fslist...)
}
