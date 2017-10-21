package fsutil

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"
)

type AssetFunc func(path string) ([]byte, error)
type AssetDirFunc func(path string) ([]string, error)
type AssetInfoFunc func(path string) (os.FileInfo, error)

type HTTPBindataFs struct {
	AssetFunc
	AssetDirFunc
	AssetInfoFunc
	Prepend string // require this in front of a name request in order to satisfy it (i.e. make it appear like all files are under this dir)
}

func (fs *HTTPBindataFs) Open(name string) (http.File, error) {

	name = strings.TrimPrefix(path.Clean("/"+name), "/")
	name = strings.TrimPrefix(path.Clean("/"+strings.TrimPrefix(name, fs.Prepend)), "/")

	data, err := fs.AssetFunc(name)
	if err != nil {
		return nil, err
	}
	return &HTTPBindataFile{name: name, Reader: bytes.NewReader(data), fs: fs}, nil
}

type HTTPBindataFile struct {
	name string
	*bytes.Reader
	fs *HTTPBindataFs
}

func (f *HTTPBindataFile) Close() error { return nil }

func (f *HTTPBindataFile) Name() string { return f.name }

func (f *HTTPBindataFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, &os.SyscallError{
		Syscall: "readdirent",
		Err:     syscall.EINVAL,
	}
}

func (f *HTTPBindataFile) Stat() (os.FileInfo, error) { return f.fs.AssetInfoFunc(f.name) }

var _ http.File = &HTTPBindataFile{}
