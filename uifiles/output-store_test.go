package uifiles

import (
	"testing"
	"time"

	"github.com/gocaveman/caveman/filesystem/aferofs"
	"github.com/gocaveman/caveman/webutil"
	"github.com/spf13/afero"
)

func TestFileSystemOutputStore(t *testing.T) {

	fs := aferofs.New(afero.NewMemMapFs())
	outputStore := NewFileSystemOutputStore(fs, time.Millisecond*100)

	err := outputStore.WriteFile("example1.txt", []byte("example1"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := outputStore.ReadFile("example1.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "example1" {
		t.Fatalf("wrong content: %q", string(b))
	}

	for i := 0; i < 5; i++ {
		_, err := outputStore.ReadFile("example1.txt")
		// should not fail
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 50)
		err = outputStore.GCFiles()
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Millisecond * 110)
	err = outputStore.GCFiles()
	if err != nil {
		t.Fatal(err)
	}

	_, err = outputStore.ReadFile("example1.txt")
	if err != webutil.ErrNotFound {
		t.Fatalf("expected webutil.ErrNotFound but instead got: %v", err)
	}

}
