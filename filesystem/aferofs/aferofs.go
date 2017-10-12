// Adapt Fs from github.com/spf13/afero to Filesystem
package aferofs

import (
	"os"
	"time"

	"github.com/gocaveman/caveman/filesystem"
	"github.com/spf13/afero"
)

// New returns a new instance adapted from afero.Fs.
func New(fs afero.Fs) filesystem.FileSystem {
	return &AferoFs{Fs: fs}
}

type AferoFs struct {
	Fs afero.Fs
}

func (a *AferoFs) Chmod(name string, mode os.FileMode) error {
	return a.Fs.Chmod(name, mode)
}

func (a *AferoFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return a.Fs.Chtimes(name, atime, mtime)
}

func (a *AferoFs) Mkdir(name string, perm os.FileMode) error {
	return a.Fs.Mkdir(name, perm)
}

func (a *AferoFs) MkdirAll(path string, perm os.FileMode) error {
	return a.Fs.MkdirAll(path, perm)
}

func (a *AferoFs) Remove(name string) error {
	return a.Fs.Remove(name)
}

func (a *AferoFs) RemoveAll(path string) error {
	return a.Fs.RemoveAll(path)
}

func (a *AferoFs) Rename(oldname, newname string) error {
	return a.Fs.Rename(oldname, newname)
}

func (a *AferoFs) Stat(name string) (os.FileInfo, error) {
	return a.Fs.Stat(name)
}

func (a *AferoFs) Create(name string) (interface {
	Close() error
	Seek(offset int64, whence int) (int64, error)
	Read(p []byte) (n int, err error)
	Write(data []byte) (n int, err error)
}, error) {
	return a.Fs.Create(name)
}

func (a *AferoFs) OpenFile(name string, flag int, perm os.FileMode) (interface {
	Close() error
	Seek(offset int64, whence int) (int64, error)
	Read(p []byte) (n int, err error)
	Write(data []byte) (n int, err error)
	Readdir(count int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
}, error) {
	return a.Fs.OpenFile(name, flag, perm)
}

func (a *AferoFs) Open(name string) (interface {
	Close() error
	Seek(offset int64, whence int) (int64, error)
	Read(p []byte) (n int, err error)
	Readdir(count int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
}, error) {
	return a.Fs.Open(name)
}
