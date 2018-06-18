package gen

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {

	assert := assert.New(t)

	// make a temp dir and set up a demo project
	tmpDir, err := ioutil.TempDir("", "TestGen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	// log.Printf("tmpDir: %s", tmpDir)

	// switch to tmpDir and when done, switch back
	saveDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(saveDir)
	os.Chdir(tmpDir)

	os.Mkdir(filepath.Join(tmpDir, "src"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "src/demoproj"), 0755)

	s := &Settings{
		WorkDir: tmpDir,
		GOPATH:  tmpDir,
	}

	assert.NoError(globalMapGenerator.Generate(s, "model-sample-todo-item", "src/demoproj/model-todo.go"))
	assert.NoError(globalMapGenerator.Generate(s, "store", "-name", "DemoStore", "src/demoproj/store.go"))
	bdata, err := ioutil.ReadFile(filepath.Join(s.GOPATH, "src/demoproj/store.go"))
	assert.NoError(err)
	assert.Contains(string(bdata), "DemoStore")
	assert.NoError(globalMapGenerator.Generate(s, "store", "src/demoproj/store.go"))

	packageName, err := DetectDirPackage(filepath.Join(tmpDir, "src/demoproj"))
	assert.NoError(err)
	assert.Equal("demoproj", packageName)

}
