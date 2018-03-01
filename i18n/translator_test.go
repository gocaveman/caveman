package i18n

import (
	"testing"

	"github.com/gocaveman/caveman/webutil"
	"github.com/stretchr/testify/assert"
)

func TestMapTranslator(t *testing.T) {

	assert := assert.New(t)

	mt := NewMapTranslator()
	mt.SetEntry("default", "example1", "en", "Example Number 1")
	mt.SetEntry("default", "example1", "es", "Ejemplo número 1")

	v, err := mt.Translate("default", "example1", "en")
	assert.Nil(err)
	assert.Equal("Example Number 1", v)

	v, err = mt.Translate("default", "example1", "es", "en")
	assert.Nil(err)
	assert.Equal("Ejemplo número 1", v)

	v, err = mt.Translate("default", "example1", "fi", "en")
	assert.Nil(err)
	assert.Equal("Example Number 1", v)

	_, err = mt.Translate("default", "example1", "fi")
	assert.Equal(ErrNotFound, err)

	_, err = mt.Translate("default", "example-not-exist", "en")
	assert.Equal(ErrNotFound, err)

	_, err = mt.Translate("not-exist", "example1", "en")
	assert.Equal(ErrNotFound, err)

}

func TestNamedSequenceTranslator(t *testing.T) {

	assert := assert.New(t)

	mt1 := NewMapTranslator()
	mt1.SetEntry("default", "example1", "en", "Example Number 1 (mt1)")
	mt1.SetEntry("default", "example1", "es", "Ejemplo número 1 (mt1)")

	mt2 := NewMapTranslator()
	mt2.SetEntry("default", "example1", "es", "Ejemplo número 1 (mt2)")
	mt2.SetEntry("default", "example1", "fi", "Esimerkki numero 1 (mt2)")

	var ns webutil.NamedSequence
	ns = append(ns, webutil.NamedSequenceItem{Sequence: 1, Name: "mt1", Value: mt1})
	ns = append(ns, webutil.NamedSequenceItem{Sequence: 2, Name: "mt2", Value: mt2})

	nst := NewNamedSequenceTranslator(ns, true)

	v, err := nst.Translate("default", "example1", "en")
	assert.Nil(err)
	assert.Equal("Example Number 1 (mt1)", v)

	v, err = nst.Translate("default", "example1", "es")
	assert.Nil(err)
	assert.Equal("Ejemplo número 1 (mt1)", v)

	v, err = nst.Translate("default", "example1", "fi")
	assert.Nil(err)
	assert.Equal("Esimerkki numero 1 (mt2)", v)

	_, err = nst.Translate("not-here", "not-here", "en")
	assert.Equal(ErrNotFound, err)

}
