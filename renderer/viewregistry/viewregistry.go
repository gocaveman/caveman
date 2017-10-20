package viewregistry

import (
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

var reg webutil.SequenceList

// Register adds a new FileSystem to the view registry.  Duplicates (seq, name or value) are not detected or prevented.
func Register(seq float64, name string, viewFS http.FileSystem) {
	reg = append(reg, webutil.SequenceListItem{Sequence: seq, Name: name, Value: viewFS})
}

// Contents returns the current contents of the registry as a SequenceList.
func Contents() webutil.SequenceList {
	if OnlyReadableFromMain {
		webutil.MainOnly(1)
	}
	return reg.SortedCopy()
}

// MakeFS sorts the SequenceList you provide and returns a FileSystem which is
// the combination of the Values in this list.   Will panic if any value is
// not a http.FileSystem.
func MakeFS(contents webutil.SequenceList) http.FileSystem {
	l := contents.SortedCopy()

	fslist := make([]http.FileSystem, 0, len(l))
	for _, item := range l {
		fslist = append(fslist, item.Value.(http.FileSystem))
	}

	return fsutil.NewHTTPStackedFileSystem(fslist...)
}
