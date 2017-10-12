// The filesystem package provides an interface for a filesystem using only stdlib types.
// The filesystem.FileSystem interface was inspired by github.com/spf13/afero but addresses
// the issue that implementations of afero must depend on the afero library (due to the fact that
// some methods return afero.File - we instead return anonymous interfaces that just list
// the required methods for the return type).  Well-designed interfaces should
// reduce coupling rather than add it.  Thus this interface can be implemented by other packages
// with no regard as to whether they are intended for use with Caveman or not and yet be
// interoperable.  Some of unnecessary/redundant methods (Stat, Truncate, WriteString) have been removed in order to simplify.
// Also some methods (Sync, Name, ReadAt, WriteAt) did not seem relevant enough to enforce at the interface level,
// however the underlying implementation may still provide it, in which case it and it can be accessed
// with a type assertion.
// The subdir "aferofs" provides an implementation that adapts an afero.Fs to
// implement filesystem.Filesystem.
package filesystem

import (
	"os"
	"time"
)

// FileSystem should be implemented to provide a working read/write filesystem.
// These methods should the same as described in the os package.
type FileSystem interface {
	Chmod(name string, mode os.FileMode) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error

	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldname, newname string) error
	Stat(name string) (os.FileInfo, error)

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
		Readdir(count int) ([]os.FileInfo, error)
		Readdirnames(n int) ([]string, error)
	}, error)
	Open(name string) (interface {
		Close() error
		Seek(offset int64, whence int) (int64, error)
		Read(p []byte) (n int, err error)
		Readdir(count int) ([]os.FileInfo, error)
		Readdirnames(n int) ([]string, error)
	}, error)
}
