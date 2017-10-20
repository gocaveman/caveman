package uiregistry

import "testing"

func TestUIRegistry(t *testing.T) {

	t.Logf("TestUIRegistry")

	reg := NewUIRegistry()

	reg.Register("js:jquery", nil, NewBytesDataSource([]byte("/*jquery.js*/")))
	reg.Register("js:bootstrap", []string{"js:jquery"}, NewBytesDataSource([]byte("/*bootstrap.js*/")))
	reg.Register("js:something", []string{"js:bootstrap"}, NewBytesDataSource([]byte("/*something.js*/")))

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
