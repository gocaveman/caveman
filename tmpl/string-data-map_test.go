package tmpl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLStringDataMap(t *testing.T) {

	assert := assert.New(t)

	m, err := ReadYAMLStringDataMap(bytes.NewReader([]byte(`

# comment here

# a comment there
something: blah

# another comment
something_else:
    bleh: blee

# whatelse
a_list: ["hi","hiya","hey","howdy"]

# two entries in a row (second should have no comment)
entry1: blah1
entry2: blah2

# going to delete this one
deleteme: here

`)))
	assert.NoError(err)

	t.Logf("DATA: %#v", m)

	// delete an entry that is not there
	m.SetEntry("nonexistent entry", nil)

	// delete an entry that is there
	m.SetEntry("deleteme", nil)

	// update an existing entry
	m.SetEntry("something", "blah_updated")

	// add a new entry
	m.SetEntry("somethingnew", "newvalue")

	var buf bytes.Buffer
	err = WriteYAMLStringDataMap(&buf, m)
	assert.NoError(err)

	t.Logf("OUT: %s", buf.String())

	assert.Equal(`

# comment here

# a comment there
something: blah_updated

# another comment
something_else:
  bleh: blee

# whatelse
a_list:
- hi
- hiya
- hey
- howdy

# two entries in a row (second should have no comment)
entry1: blah1
entry2: blah2
somethingnew: newvalue

`, buf.String())

}
