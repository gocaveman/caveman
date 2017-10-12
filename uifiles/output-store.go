package uifiles

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/gocaveman/caveman/filesystem"
	"github.com/gocaveman/caveman/webutil"
)

// OutputStore is the interface for something that can store the combined files we generate and can garbage collect old files.
// There is no positive way to identify that a file is no longer being used, however no activity over a reasonable time period
// (a day for example) should work well for most applications.
type OutputStore interface {

	// WriteFile saves a files to the store, overwriting it if it exists.
	WriteFile(fname string, content []byte) error

	// ReadFile returns the contents of the given file, or webutil.ErrNotFound if it's not there, or other error.
	// When this function is called, the implementation should update whatever timestamp so it avoids being GCed -
	// files may be written only once and read a great many times and this causing a GC could be bad.
	ReadFile(fname string) ([]byte, error)

	// GCFiles is called periodically to tell the store to remove old unused files.
	// The time required to call a file old enough to be removed is up to the implementation.
	GCFiles() error
}

// FileSystemOutputStore implements OutputStore on top of a filesystem.FileSystem
type FileSystemOutputStore struct {
	FileSystem       filesystem.FileSystem
	GCFilesOlderThan time.Duration
}

func NewFileSystemOutputStore(fs filesystem.FileSystem, gcFilesOlderThan time.Duration) *FileSystemOutputStore {
	if gcFilesOlderThan <= 0 {
		gcFilesOlderThan = time.Hour * 24
	}
	return &FileSystemOutputStore{FileSystem: fs, GCFilesOlderThan: gcFilesOlderThan}
}

func (s *FileSystemOutputStore) WriteFile(fname string, content []byte) error {
	f, err := s.FileSystem.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	return err
}

func (s *FileSystemOutputStore) ReadFile(fname string) ([]byte, error) {
	f, err := s.FileSystem.Open(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, webutil.ErrNotFound
		}
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	// update timestamp, to ensure GC doesn't eat it
	err = s.FileSystem.Chtimes(fname, time.Now(), time.Now())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *FileSystemOutputStore) GCFiles() error {

	rootF, err := s.FileSystem.Open("/")
	if err != nil {
		return err
	}
	defer rootF.Close()

	fis, err := rootF.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if fi.ModTime().Add(s.GCFilesOlderThan).Before(time.Now()) {
			err = s.FileSystem.Remove(fi.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
