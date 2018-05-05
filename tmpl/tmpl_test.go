package tmpl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseYAMLHeadTemplate(t *testing.T) {

	assert := assert.New(t)

	in := []byte(`---

somekey: somevalue
somekey2: somevalue2

---
{{template "whatever" .}}
`)

	meta, body, err := ParseYAMLHeadTemplate(bytes.NewReader(in))
	assert.NoError(err)
	assert.Equal("somevalue", meta.Data("somekey"))
	assert.Len(meta.Keys(), 2)
	assert.Contains(string(body), `{{template "whatever" .}}`+"\n")

}
