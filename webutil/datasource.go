package webutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DataSource
type DataSource interface {
	// OpenData opens a readable stream of data for the file.
	OpenData() (ReadSeekCloser, error)
	// Stat provides metadata about the underlying resource, particularly ModTime()
	Stat() (os.FileInfo, error)
}

// ReadSeekCloser is exactly what you think it is - it combines io.Reader, io.Seeker and io.Closer
type ReadSeekCloser interface {
	io.Closer
	io.Reader
	io.Seeker
}

// type FileDataSource struct {
// 	fs afero.Fs // hm, this is a problem...  webutil really shouldn't require afero; maybe this is where we internalize the Fs interface...
// 	p  string
// }

// func (fds *FileDataSource) OpenData() (ReadSeekCloser, error) {
// 	f, err := fds.fs.Open(fds.p)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return f, nil
// }

// func NewFileDataSource(fs afero.Fs, path string) DataSource {
// 	return &FileDataSource{fs: fs, p: path}
// }

type BytesDataSource struct {
	FileInfo os.FileInfo
	b        []byte
}

func (bds *BytesDataSource) String() string {
	var s string
	if len(bds.b) > 64 {
		s = string(bds.b[:64]) + "..."
	} else {
		s = string(bds.b)
	}
	// TODO: add FileInfo.Name()?
	return fmt.Sprintf("&BytesDataSource{b=%q}", s)
}

func (bds *BytesDataSource) Stat() (os.FileInfo, error) {
	return bds.FileInfo, nil
	// return &fileInfo{
	// 	name:    bds.name,
	// 	size:    len(bds.b),
	// 	mode:    os.FileMode(0644),
	// 	modTime: bds.modTime,
	// }, nil
}

type byteData struct {
	*bytes.Reader
}

func (d *byteData) Close() error { return nil }

func (fds *BytesDataSource) OpenData() (ReadSeekCloser, error) {
	return &byteData{Reader: bytes.NewReader(fds.b)}, nil
}

func NewBytesDataSource(b []byte, name string, modTime time.Time) *BytesDataSource {
	return &BytesDataSource{
		b: b,
		FileInfo: &fileInfo{
			name:    name,
			size:    int64(len(b)),
			mode:    os.FileMode(0644),
			modTime: modTime,
		},
	}
}

type HTTPFSDataSource struct {
	fs http.FileSystem
	p  string
}

func (fds *HTTPFSDataSource) OpenData() (ReadSeekCloser, error) {
	f, err := fds.fs.Open(fds.p)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fds *HTTPFSDataSource) Stat() (os.FileInfo, error) {
	f, err := fds.fs.Open(fds.p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Stat()
}

func NewHTTPFSDataSource(fs http.FileSystem, path string) *HTTPFSDataSource {
	return &HTTPFSDataSource{fs: fs, p: path}
}

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.isDir }
func (fi *fileInfo) Sys() interface{}   { return fi.sys }
