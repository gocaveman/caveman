package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {

	assert := assert.New(t)

	p := NewDefaultProcessor()

	sout, err := p.Process("This is some text with {{.Value \"SOMETHING\"}} in it.",
		context.WithValue(context.Background(), "SOMETHING", "hot sauce"))

	assert.NoError(err)
	assert.Equal("This is some text with hot sauce in it.", sout)

}
