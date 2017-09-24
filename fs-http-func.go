package caveman

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"time"
)

func NewHTTPFuncFS(f func(name string) (http.File, error)) http.FileSystem {
	return HTTPFuncFS(f)
}

type HTTPFuncFS func(name string) (http.File, error)

func (f HTTPFuncFS) Open(name string) (http.File, error) {
	return f(name)
}

func NewHTTPBytesFile(name string, modTime time.Time, data []byte) *HTTPBytesFile {
	return &HTTPBytesFile{
		name:    name,
		modTime: modTime,
		Reader:  bytes.NewReader(data),
	}
}

type HTTPBytesFile struct {
	*bytes.Reader
	name    string
	modTime time.Time
}

func (f *HTTPBytesFile) Close() error {
	return nil
}

func (f *HTTPBytesFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("HTTPBytesFile does not support Readdir()")
}

func (f *HTTPBytesFile) Stat() (os.FileInfo, error) {
	return NewHTTPFileInfo(f.name, int64(f.Reader.Len()), os.FileMode(0644), f.modTime, nil), nil
}

func NewHTTPFileInfo(name string, size int64, mode os.FileMode, modTime time.Time, sys interface{}) *HTTPFileInfo {
	return &HTTPFileInfo{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
		sys:     sys,
	}
}

// HTTPFileInfo implements os.FileInfo
type HTTPFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

// Name returns the base name of the file
func (i *HTTPFileInfo) Name() string { return i.name }

// Size returns the length in bytes
func (i *HTTPFileInfo) Size() int64 { return i.size }

// Mode returns file mode bits
func (i *HTTPFileInfo) Mode() os.FileMode { return i.mode }

// ModTime returns the modification time
func (i *HTTPFileInfo) ModTime() time.Time { return i.modTime }

// IsDir is an abbreviation for Mode().IsDir()
func (i *HTTPFileInfo) IsDir() bool { return i.mode.IsDir() }

// Sys is underlying data source (can return nil)
func (i *HTTPFileInfo) Sys() interface{} { return i.sys }
