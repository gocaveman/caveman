// The filesystem package provides an interface for a filesystem using only stdlib types.
// The filesystem.FileSystem interface was inspired by github.com/spf13/afero but addresses
// the issue that implementations must depend on the afero library.  Well-designed interfaces
// reduce coupling rather than add it.  Thus this interface can be implemented by other packages
// with no regard as to whether they are intended for use with Caveman or not and yet be
// interoperable.  The subdir "aferofs" provides an implementation that adapts an afero.Fs to
// implement filesystem.Filesystem.
package filesystem

import (
	"os"
	"time"
)

// FileSystem should be implemented to provide a working read/write filesystem.
type FileSystem interface {
	Chmod(name string, mode os.FileMode) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error

	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldname, newname string) error
	Stat(name string) (os.FileInfo, error)

	Readdir(name string, count int) ([]os.FileInfo, error)
	Readdirnames(name string, n int) ([]string, error)

	Create(name string) (interface {
		Close() error
		Seek(offset int64, whence int) (int64, error)
		Read(p []byte) (n int, err error)
		Write(data []byte) (n int, err error)
	}, error)
	OpenFile(name string, flag int, perm os.FileMode) (interface {
		Close() error
		Seek(offset int64, whence int) (int64, error)
		Read(p []byte) (n int, err error)
		Write(data []byte) (n int, err error)
	}, error)
	Open(name string) (interface {
		Close() error
		Seek(offset int64, whence int) (int64, error)
		Read(p []byte) (n int, err error)
	}, error)
}

// type file struct {
// 	io.Closer
// 	io.Reader
// 	io.ReaderAt
// 	io.Seeker
// 	io.Writer
// 	io.WriterAt

// 	Name() : string
// 	Readdir(count int) : []os.FileInfo, error
// 	Readdirnames(n int) : []string, error
// 	Stat() : os.FileInfo, error
// 	Sync() : error
// 	Truncate(size int64) : error
// 	WriteString(s string) : ret int, err error
// }
