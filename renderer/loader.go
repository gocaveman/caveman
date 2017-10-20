package renderer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gocaveman/caveman/webutil"
)

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
func NewFileExtLoader(loaderMap map[string]Loader) *FileExtLoader {
	if loaderMap == nil {
		loaderMap = make(map[string]Loader)
	}
	return &FileExtLoader{LoaderMap: loaderMap}
}

// WithExt helps in builder a FileExtLoader. Example:
// l := NewFileExtLoader().WithExt(".gohtml", gohtmlLoader).WithExt(".md", markdownLoader)
func (l *FileExtLoader) WithExt(ext string, loader Loader) *FileExtLoader {
	l.LoaderMap[ext] = loader
	return l
}

// Load calls the next loader in our LoaderMap based on the file extension of the file.
// Returns webutil.ErrNotFound if file extension is not in map.
func (l *FileExtLoader) Load(filename string) (rc io.ReadCloser, reterr error) {
	ext := path.Ext(filename)
	nextl := l.LoaderMap[ext]
	if nextl != nil {
		return nextl.Load(filename)
	}
	return nil, webutil.ErrNotFound
}

// GohtmlLoader reads Go template files from an http.FileSystem, replacing the header block (---) with a comment.
type GohtmlLoader struct {
	FileFS http.FileSystem
}

// NewGohtmlLoader returns a new instance of GohtmlLoader.
func NewGohtmlLoader(fileFS http.FileSystem) *GohtmlLoader {
	return &GohtmlLoader{FileFS: fileFS}
}

type readCloser struct {
	io.Reader
	io.Closer
}

func (l *GohtmlLoader) Load(filename string) (rc io.ReadCloser, reterr error) {

	f, err := l.FileFS.Open(filename)
	if err != nil {
		return nil, err
	}

	// ensure f gets closed on error
	defer func() {
		if reterr != nil {
			f.Close()
		}
	}()

	var buf bytes.Buffer

	br := bufio.NewReader(f)

	line, err := br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	// see if the first line has the YAML separator
	if strings.TrimRight(line, "\t\r\n ") == "---" {

		// replace separator with go template open comment
		buf.WriteString("{{/*\n")

		for {
			line, err := br.ReadString('\n')
			if err == io.EOF {
				return nil, fmt.Errorf("file starts a block with '---' but does not close it")
			} else if err != nil {
				return nil, err
			}

			// check for end of block
			if strings.TrimRight(line, "\t\r\n ") == "---" {
				buf.WriteString("\n*/}}") // write the newline before in order to ensure templates can have no whitespace in them at the beginning if needed (doctype, e.g.)
				break
			} else {
				// not end of block, just add a blank line (keeping the line count the same for easier debugging)
				buf.WriteString("\n")
			}
		}

	} else {
		// no YAML separator, take the line we just read and put in the buffer and continue
		buf.WriteString(line)
	}

	ret := &readCloser{
		Reader: io.MultiReader(br, f),
		Closer: f,
	}

	return ret, nil
}
