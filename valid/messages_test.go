package valid

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessages(t *testing.T) {

	assert := assert.New(t)

	var ms Messages
	ms = append(ms, Message{
		FieldName: "example1",
		Message:   "Example Message",
	})

	b, err := json.Marshal(ms)
	if err != nil {
		t.Fatal(err)
	}

	bstr := string(b)
	assert.Contains(bstr, "example1")
	assert.Contains(bstr, "Example Message")

}
