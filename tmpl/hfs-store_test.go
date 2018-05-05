package tmpl

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func hfsStoreTestSetup(t *testing.T) (viewsTmpdir, includesTmpdir string, s *HFSStore) {

	assert := assert.New(t)

	var err error
	viewsTmpdir, err = ioutil.TempDir("", "TestHFSStoreViews")
	assert.NoError(err)
	includesTmpdir, err = ioutil.TempDir("", "TestHFSStoreIncludes")
	assert.NoError(err)

	assert.NoError(os.Mkdir(filepath.Join(viewsTmpdir, "test1"), 0755))

	assert.NoError(ioutil.WriteFile(filepath.Join(viewsTmpdir, "test1", "file1.gohtml"), []byte(`---
testkey1: testvalue1
---
{{template "test1" .}}
`), 0644))
	assert.NoError(ioutil.WriteFile(filepath.Join(viewsTmpdir, "test1", "file2.gohtml"), []byte(`---
testkey2: testvalue2
---
{{template "test2" .}}
`), 0644))

	assert.NoError(ioutil.WriteFile(filepath.Join(includesTmpdir, "someinclude.gohtml"), []byte(`---
testkey2: testvalue2
---
{{template "test2" .}}
`), 0644))

	s = &HFSStore{
		FileSystems: map[string]http.FileSystem{
			ViewsCategory:    http.Dir(viewsTmpdir),
			IncludesCategory: http.Dir(includesTmpdir),
		},
	}

	return

}

func TestHFSStoreFindByPrefix(t *testing.T) {

	assert := assert.New(t)

	viewsTmpdir, includesTmpdir, s := hfsStoreTestSetup(t)
	defer os.RemoveAll(viewsTmpdir)
	defer os.RemoveAll(includesTmpdir)

	cats, err := s.Categories()
	assert.NoError(err)
	assert.Len(cats, 2)

	names, err := s.FindByPrefix(ViewsCategory, "/", -1)
	assert.Equal([]string{"/test1/file1.gohtml", "/test1/file2.gohtml"}, names)
	names, err = s.FindByPrefix(ViewsCategory, "/test1", -1)
	assert.Equal([]string{"/test1/file1.gohtml", "/test1/file2.gohtml"}, names)
	names, err = s.FindByPrefix(IncludesCategory, "/", -1)
	assert.Equal([]string{"/someinclude.gohtml"}, names)

}

func TestHFSStoreReadTemplate(t *testing.T) {

	assert := assert.New(t)

	viewsTmpdir, includesTmpdir, s := hfsStoreTestSetup(t)
	defer os.RemoveAll(viewsTmpdir)
	defer os.RemoveAll(includesTmpdir)

	body, mimeType, meta, err := s.ReadTemplate(ViewsCategory, "/test1/file1.gohtml")
	assert.NoError(err)

	assert.Equal(GoTemplateHtmlMimeType, mimeType)
	assert.Equal("testvalue1", meta["testkey1"].(string))
	assert.Contains(string(body), `{{template "test1" .}}`)

	// check a file that doesn't exist
	_, _, _, err = s.ReadTemplate(ViewsCategory, "/test1/doesnotexist.gohtml")
	assert.True(os.IsNotExist(err), "err=%#v", err)

	// make sure trying to read a dir doesn't produce something funky
	_, _, _, err = s.ReadTemplate(ViewsCategory, "/test1/")
	assert.True(os.IsNotExist(err), "err=%#v", err)

}
