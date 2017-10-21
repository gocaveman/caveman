package uiregistry

import (
	"testing"
	"time"

	"github.com/gocaveman/caveman/webutil"
)

func TestUIRegistry(t *testing.T) {

	t.Logf("TestUIRegistry")

	reg := NewUIRegistry()

	reg.Register("js:jquery", nil, webutil.NewBytesDataSource([]byte("/*jquery.js*/"), "jquery.js", time.Now()))
	reg.Register("js:bootstrap", []string{"js:jquery"}, webutil.NewBytesDataSource([]byte("/*bootstrap.js*/"), "bootstrap.js", time.Now()))
	reg.Register("js:something", []string{"js:bootstrap"}, webutil.NewBytesDataSource([]byte("/*something.js*/"), "something.js", time.Now()))

	names, err := reg.ResolveDeps("js:something")
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 3 {
		t.FailNow()
	}

	for _, n := range names {
		ds, err := reg.Lookup(n)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("ds: %v", ds)
	}

}
