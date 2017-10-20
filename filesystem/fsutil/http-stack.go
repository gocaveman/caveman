package fsutil

import (
	"log"
	"net/http"
)

// HTTPStackedFileSystem implements http.FileSystem by falling back through multiple layers
// to find where a file exists and using the first match.  FileSystems are checked in the
// order provided, i.e. index 0 takes precedence over all others, etc.
type HTTPStackedFileSystem struct {
	Stack []http.FileSystem
}

func NewHTTPStackedFileSystem(subfss ...http.FileSystem) *HTTPStackedFileSystem {
	return &HTTPStackedFileSystem{
		Stack: subfss,
	}
}

func (fs *HTTPStackedFileSystem) Open(name string) (ret http.File, err error) {
	log.Printf("need a way to debug this, so we can find out which layer files are being found in")
	for _, subfs := range fs.Stack {
		ret, err = subfs.Open(name)
		if err == nil {
			break
		}
	}
	return
}
