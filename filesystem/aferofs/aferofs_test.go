package aferofs

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

func TestAferofs(t *testing.T) {

	fs := New(afero.NewMemMapFs())
	f, err := fs.Create("/example1.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("testing123\n"))
	f.Close()

	f2, err := fs.Open("/example1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	b, err := ioutil.ReadAll(f2)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "testing123\n" {
		t.Fatalf("unexpected response: %q", string(b))
	}

}
