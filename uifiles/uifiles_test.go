package uifiles

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gocaveman/caveman/webutil"
	"github.com/spf13/afero"
)

type testUIResolver struct{}

var resourceTime time.Time

func init() {
	resourceTime, _ = time.Parse(time.RFC3339, "2017-01-02T15:04:05Z")
}

func (tr *testUIResolver) Lookup(name string) (webutil.DataSource, error) {
	if name == "js:example.com/test1.js" {
		return webutil.NewBytesDataSource([]byte("/* test1.js */\nvar test1;\n"), "js:example.com/test1.js", resourceTime), nil
	}
	if name == "js:example.com/test2.js" {
		return webutil.NewBytesDataSource([]byte("/* test2.js */\nvar test2;\n"), "js:example.com/test2.js", resourceTime), nil
	}
	return nil, webutil.ErrNotFound
}

func (tr *testUIResolver) ResolveDeps(d ...string) ([]string, error) {
	return d, nil
}

func setupFileManglerTest() (uiResolver UIResolver, localFSDir string, cacheDir string, reterr error) {

	uiResolver = &testUIResolver{}

	localFSDir, reterr = ioutil.TempDir("", "fmtest-localFSDir")
	if reterr != nil {
		return
	}

	// make some local sample files
	reterr = ioutil.WriteFile(filepath.Join(localFSDir, "testA.js"), []byte("/* testA.js */\nvar testA;\n"), 0644)
	if reterr != nil {
		return
	}

	reterr = ioutil.WriteFile(filepath.Join(localFSDir, "testB.js"), []byte("/* testB.js */\nvar testB;\n"), 0644)
	if reterr != nil {
		return
	}

	cacheDir, reterr = ioutil.TempDir("", "fmtest-cacheDir")
	if reterr != nil {
		return
	}

	return
}

func TestFileMangler(t *testing.T) {

	// Things to test:
	// - require resolved/local file
	// - generate set
	// - request set file
	// - re-create FileMangler and request set file again (should get from disk)
	// - delete the disk file and request again (should regenerate from token)
	// - touch one of the disk files (should break the prehash and regenerate but otherwise work correctly)
	// - build the set, delete the disk file, change one of the contents of the files and request the URL - should invoke WrongContentHandler

	uiResolver, localFSDir, cacheDir, err := setupFileManglerTest()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(localFSDir)
	defer os.RemoveAll(cacheDir)

	outputStore := NewRWFSOutputStore(afero.NewBasePathFs(afero.NewOsFs(), cacheDir))
	log.Printf("outputStore: %#v", outputStore)

	fm := NewFileMangler(uiResolver, http.Dir(localFSDir), cacheDir)
	fm.TokenKey = []byte("test")

	wrec := httptest.NewRecorder()
	var w http.ResponseWriter = wrec

	// create a FileSet attached to the context by running it through ServeHTTPChain
	r := httptest.NewRequest("GET", "/some/page.html", nil)
	w, r = fm.ServeHTTPChain(w, r)
	fs := r.Context().Value("uifiles.FileSet").(*FileSet)

	// try resolved file
	err = fs.Require("js:example.com/test1.js")
	if err != nil {
		t.Fatal(err)
	}

	// try local file
	err = fs.Require("js:/testA.js")
	if err != nil {
		t.Fatal(err)
	}

	// generate the set (combined file)
	setp, err := fs.BuildSetPath("js")
	if err != nil {
		t.Fatal(err)
	}

	// request that combined file
	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", setp, nil)
	fm.ServeHTTPChain(w, r)

	// make sure it looks right
	b, _ := httputil.DumpResponse(wrec.Result(), true)
	if (!bytes.Contains(b, []byte(`test1`))) || (!bytes.Contains(b, []byte(`testA`))) {
		t.Fatalf("response did not contain expected strings, instead got: %s", b)
	}

	// remake the FileMangler (as if server bounced)
	fm = NewFileMangler(uiResolver, http.Dir(localFSDir), cacheDir)
	fm.TokenKey = []byte("test")

	// request combined file again
	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", setp, nil)
	fm.ServeHTTPChain(w, r)

	// check again to make sure it looks right
	b, _ = httputil.DumpResponse(wrec.Result(), true)
	if (!bytes.Contains(b, []byte(`test1`))) || (!bytes.Contains(b, []byte(`testA`))) {
		t.Fatalf("response did not contain expected strings, instead got: %s", b)
	}

}
