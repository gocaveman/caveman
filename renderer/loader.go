package renderer

import "io"

// Loader is implemented by things which can take a file name and produce Go template markup.  In some cases
// conversion may be necessary, in others it make be a fairly direct load from whatever source.
// A loader implementation should minimize its knowledge of the underlying file source, usually by using
// an http.FileSystem.
type Loader interface {
	Load(filename string) (io.ReadCloser, error)
}

// FileExtLoader implements Loader by calling another loader based on file extension.
type FileExtLoader struct {
	LoaderMap map[string]Loader
}

// NewFileExtLoader creates a new FileExtLoader.  WithExt is usually called right after.
func NewFileExtLoader() *FileExtLoader {
	return &FileExtLoader{LoaderMap: make(map[string]Loader)}
}

// WithExt helps in builder a FileExtLoader. Example:
// l := NewFileExtLoader().WithExt(".gohtml", gohtmlLoader).WithExt(".md", markdownLoader)
func (l *FileExtLoader) WithExt(ext string, loader Loader) *FileExtLoader {
	l.LoaderMap[ext] = loader
	return l
}

type GohtmlLoader struct {
}
