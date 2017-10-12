package uifiles

import (
	"io/ioutil"
	"os"

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

// RWFSOutputStore implements OutputStore on top of a webutil.RWFS
type RWFSOutputStore struct {
	RWFS webutil.RWFS
}

func NewRWFSOutputStore(rwfs webutil.RWFS) *RWFSOutputStore {
	return &RWFSOutputStore{RWFS: rwfs}
}

func (s *RWFSOutputStore) WriteFile(fname string, content []byte) error {
	f, err := s.RWFS.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(content)
	return err
}

func (s *RWFSOutputStore) ReadFile(fname string) ([]byte, error) {
	f, err := s.RWFS.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (s *RWFSOutputStore) GCFiles() error {
	panic("not implemented yet")
	return nil
}
