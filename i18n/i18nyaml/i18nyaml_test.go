package i18nyaml

import (
	"bytes"
	"os"
	"testing"

	"github.com/gocaveman/caveman/filesystem/aferofs"
	"github.com/gocaveman/caveman/i18n"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {

	assert := assert.New(t)

	yamlText := `
test1: Test Number 1
test2: Test Number 2
test3: Test Number 3
`

	tr, err := Load(bytes.NewReader([]byte(yamlText)), "default", "en")
	assert.Nil(err)

	v, err := tr.Translate("default", "test2", "en")
	assert.Nil(err)

	assert.Equal("Test Number 2", v)

}

func TestLoadDir(t *testing.T) {

	assert := assert.New(t)

	afs := afero.NewMemMapFs()
	fs := aferofs.New(afs)
	fs.Mkdir("/i18n", 0755)

	f, err := fs.OpenFile("/i18n/default.en.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	assert.Nil(err)
	defer f.Close()
	f.Write([]byte(`
test1: Test Number 1
test2: Test Number 2
test3: Test Number 3
`))

	f, err = fs.OpenFile("/i18n/default.es.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	assert.Nil(err)
	defer f.Close()
	f.Write([]byte(`
test1: Prueba número 1
test2: Prueba número 2
test3: Prueba número 3
`))

	ns, err := LoadDir(afero.NewHttpFs(afs), "/i18n")
	assert.Nil(err)

	tr := i18n.NewNamedSequenceTranslator(ns, true)
	v, err := tr.Translate("default", "test2", "es")
	assert.Nil(err)

	assert.Equal("Prueba número 2", v)

}
