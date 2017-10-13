package uifiles

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gocaveman/caveman/filesystem/aferofs"
	"github.com/gocaveman/caveman/webutil"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type testUIResolver struct{}

var resourceTime time.Time

func init() {
	resourceTime, _ = time.Parse(time.RFC3339, "2017-01-02T15:04:05Z")
}

func (tr *testUIResolver) Lookup(name string) (webutil.DataSource, error) {
	if name == "js:example.com/test1" {
		return webutil.NewBytesDataSource([]byte("/* test1.js */\nvar test1;\n"), "js:example.com/test1", resourceTime), nil
	}
	if name == "js:example.com/test2" {
		return webutil.NewBytesDataSource([]byte("/* test2.js */\nvar test2;\n"), "js:example.com/test2", resourceTime), nil
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

	assert := assert.New(t)

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

	outputStore := NewFileSystemOutputStore(aferofs.New(afero.NewBasePathFs(afero.NewOsFs(), cacheDir)), 0)

	fm := NewFileMangler(uiResolver, http.Dir(localFSDir), outputStore)
	fm.TokenKey = []byte("test")

	wrec := httptest.NewRecorder()
	var w http.ResponseWriter = wrec

	// create a FileSet attached to the context by running it through ServeHTTPChain
	r := httptest.NewRequest("GET", "/some/page.html", nil)
	w, r = fm.ServeHTTPChain(w, r)
	fs := r.Context().Value("uifiles.FileSet").(*FileSet)

	// try resolved file
	err = fs.Require("js:example.com/test1")
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

	// check for minification
	if bytes.Contains(b, []byte(`/*`)) {
		t.Logf("resulting JS file does not appear to be minified (it has a comment)")
		t.Fail()
	}

	// remake the FileMangler (as if server bounced)
	fm = NewFileMangler(uiResolver, http.Dir(localFSDir), outputStore)
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

	// recreate FileMangler, delete the disk file and request again (should regenerate from token)
	fm = NewFileMangler(uiResolver, http.Dir(localFSDir), outputStore)
	fm.TokenKey = []byte("test")
	os.RemoveAll(cacheDir)
	os.Mkdir(cacheDir, 0755)

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

	// touch one of the disk files (should break the prehash and regenerate but otherwise work correctly)
	cacheDirF, err := os.Open(cacheDir)
	assert.Nil(err)
	names, err := cacheDirF.Readdirnames(-1)
	assert.Nil(err)
	for _, name := range names {
		os.Chtimes(filepath.Join(cacheDir, filepath.Base(name)), time.Now(), time.Now())
	}
	cacheDirF.Close()

	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", setp, nil)
	fm.ServeHTTPChain(w, r)

	// check again to make sure it looks right
	b, _ = httputil.DumpResponse(wrec.Result(), true)
	if (!bytes.Contains(b, []byte(`test1`))) || (!bytes.Contains(b, []byte(`testA`))) {
		t.Fatalf("response did not contain expected strings, instead got: %s", b)
	}

	// using the same set, delete the disk cache file, change one of the contents of the files and request the URL again - should invoke WrongContentHandler
	fm = NewFileMangler(uiResolver, http.Dir(localFSDir), outputStore)
	fm.TokenKey = []byte("test")
	gotWrongContent := false
	fm.WrongContentHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotWrongContent = true
	})
	os.RemoveAll(cacheDir)
	os.Mkdir(cacheDir, 0755)
	assert.Nil(ioutil.WriteFile(filepath.Join(localFSDir, "testA.js"), []byte("/* testA2.js */\nvar testA2;\n"), 0644))

	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", setp, nil)
	fm.ServeHTTPChain(w, r)

	b, _ = httputil.DumpResponse(wrec.Result(), true)
	if (!bytes.Contains(b, []byte(`test1`))) || (!bytes.Contains(b, []byte(`testA2`))) {
		t.Fatalf("response did not contain expected strings, instead got: %s", b)
	}

	assert.True(gotWrongContent)

	// test the missing file case by removing testA.js, the cache file, and reseting fm and requesting setp again, should invoke WrongContentHandler
	fm = NewFileMangler(uiResolver, http.Dir(localFSDir), outputStore)
	fm.TokenKey = []byte("test")
	gotWrongContent = false
	fm.WrongContentHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotWrongContent = true
	})
	os.RemoveAll(cacheDir)
	os.Mkdir(cacheDir, 0755)
	os.Remove(filepath.Join(localFSDir, "testA.js"))

	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", setp, nil)
	fm.ServeHTTPChain(w, r)

	b, _ = httputil.DumpResponse(wrec.Result(), true)
	if !bytes.Contains(b, []byte(`test1`)) {
		t.Fatalf("response did not contain expected strings, instead got: %s", b)
	}

	assert.True(gotWrongContent)

	// verify that the OutputStore GC is being called by setting the timing and ensuring the disk file disappears
	// NOTE: timestamps on disk are often second-granular so we have to measure things in full seconds
	fm = NewFileMangler(uiResolver, http.Dir(localFSDir), outputStore)
	fm.TokenKey = []byte("test")
	fm.OutputStoreGCDelay = time.Millisecond * 50
	outputStore.GCFilesOlderThan = time.Second * 1

	// revert testA.js back to it's original content
	assert.Nil(ioutil.WriteFile(filepath.Join(localFSDir, "testA.js"), []byte("/* testA.js */\nvar testA;\n"), 0644))

	// regenerate set
	r = httptest.NewRequest("GET", "/some/page.html", nil)
	w, r = fm.ServeHTTPChain(w, r)
	fs = r.Context().Value("uifiles.FileSet").(*FileSet)
	assert.Nil(fs.Require("js:example.com/test1"))
	assert.Nil(fs.Require("js:/testA.js"))
	setp, err = fs.BuildSetPath("js")
	assert.Nil(err)

	fname := path.Base(strings.SplitN(setp, "?", 2)[0])

	// request it a few times
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * 1)
		wrec = httptest.NewRecorder()
		w = wrec
		r = httptest.NewRequest("GET", setp, nil)
		fm.ServeHTTPChain(w, r)
		// make sure it's still there
		st, err := os.Stat(filepath.Join(cacheDir, fname))
		assert.Nil(err, "setp=%q fname=%q st=%+v i=%v", setp, fname, st, i)
		// log.Printf("in gc check loop fname=%q st=%+v", fname, st)
	}
	// now request something else a few times and let the time expire
	for i := 0; i < 2; i++ {
		time.Sleep(time.Second * 1)
		wrec = httptest.NewRecorder()
		w = wrec
		// anything under /fm-assets/ should still be causing the GC to fire
		r = httptest.NewRequest("GET", "/fm-assets/404.whatever", nil)
		fm.ServeHTTPChain(w, r)
	}
	_, err = os.Stat(filepath.Join(cacheDir, fname))
	assert.NotNil(err)
	assert.True(os.IsNotExist(err))

	// try FilePaths - make sure output is right and try requesting file(s)
	filePaths, err := fs.FilePaths("js")
	assert.Nil(err)

	// request each path and make sure the file looks right
	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", filePaths[0], nil)
	fm.ServeHTTPChain(w, r)
	b, _ = httputil.DumpResponse(wrec.Result(), true)
	// check for relevant text and JS comment (verifying it did not get minified)
	if !bytes.Contains(b, []byte(`test1`)) || !bytes.Contains(b, []byte(`/*`)) {
		t.Fatalf("response did not contain expected strings, instead got: %s", b)
	}
	// check cache header
	assert.Contains(w.Header().Get("cache-control"), "max-age", "path=%q", filePaths[0])

	// change the version number and try again - make sure the cache control header isn't telling us to keep the file
	wrongVerFname := strings.Replace(filePaths[0], "?ver=", "?ver=_wrong_", 1)
	wrec = httptest.NewRecorder()
	w = wrec
	r = httptest.NewRequest("GET", wrongVerFname, nil)
	fm.ServeHTTPChain(w, r)
	assert.Contains(w.Header().Get("cache-control"), "no-store")

}
