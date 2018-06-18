package tmpl

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/shurcooL/httpfs/vfsutil"
)

// HFSStore implements read-only Store on top of a set of http.FileSystem.
type HFSStore struct {
	FileExtMimeTypes map[string]string
	FileSystems      map[string]http.FileSystem // key is category, value FileSystem to expose
}

// CreateTemplate returns ErrWriteNotSupported
func (s *HFSStore) CreateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error {
	return ErrWriteNotSupported
}

// UpdateTemplate returns ErrWriteNotSupported
func (s *HFSStore) UpdateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error {
	return ErrWriteNotSupported
}

// DeleteTemplate returns ErrWriteNotSupported
func (s *HFSStore) DeleteTemplate(category, fileName string) error {
	return ErrWriteNotSupported
}

// Categories returns the list of available categories.
func (s *HFSStore) Categories() ([]string, error) {
	var ret []string
	for k := range s.FileSystems {
		ret = append(ret, k)
	}
	return ret, nil
}

// Read a template, returning a complete Tmpl with Body and Meta.
// ErrNotFound will be returned if it doesn't exist.
func (s *HFSStore) ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error) {

	// defer func() {
	// 	if err != nil {
	// 		log.Printf("ReadTemplate(%q, %q), err: %v", category, fileName, err)
	// 	}
	// }()

	fs := s.FileSystems[category]
	if fs == nil {
		return nil, "", nil, ErrNotFound
	}

	f, err := fs.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) || err == ErrNotFound {
			return nil, "", nil, ErrNotFound
		}
		return nil, "", nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, "", nil, err
	}
	if fi.IsDir() {
		return nil, "", nil, ErrNotFound // directories are invisible as far as ReadTemplate is concerned
	}

	var yamlMeta *YAMLStringDataMap
	yamlMeta, body, err = ParseYAMLHeadTemplate(f)
	if err != nil {
		return nil, "", nil, err
	}

	meta = yamlMeta.Map()

	mimeType = s.FileExtMimeTypes[path.Ext(fileName)]
	if mimeType == "" {
		mimeType = DefaultFileExtMimeTypes[path.Ext(fileName)]
	}

	err = nil

	return
}

// FindByPrefix will return a slice of file names for the specified category
// that begin with a prefix, up to an indicated limit.  Limit -1 means all.
// Prefix "/" will list all.  Prefix must be a directory name.
func (s *HFSStore) FindByPrefix(category, fileNamePrefix string, limit int) ([]string, error) {

	fs := s.FileSystems[category]
	if fs == nil {
		return nil, ErrNotFound
	}

	var ret []string

	var done = errors.New("done")

	err := vfsutil.WalkFiles(fs, fileNamePrefix, func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		ret = append(ret, path)
		if limit > 0 && len(ret) >= limit {
			return done
		}
		return nil
	})
	if err != nil && err != done {
		return nil, err
	}

	if limit > 0 && len(ret) > limit {
		ret = ret[:limit]
	}

	return ret, nil
}
