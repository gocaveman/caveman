package renderer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFileNamer(t *testing.T) {

	assert := assert.New(t)

	fn := NewDefaultFileNamer(".gohtml", ".html")
	assert.Equal([]string{"/index.gohtml", "/index.html"}, fn.FileNames("/"))
	assert.Equal([]string(nil), fn.FileNames("/index"))
	assert.Equal([]string(nil), fn.FileNames("/index.html"))
	assert.Equal([]string{"/example.gohtml", "/example.html"}, fn.FileNames("/example"))
	assert.Equal([]string(nil), fn.FileNames("/example.html"))
	assert.Equal([]string{"/example/index.gohtml", "/example/index.html"}, fn.FileNames("/example/"))
	assert.Equal([]string(nil), fn.FileNames("/../example/"))
	assert.Equal([]string(nil), fn.FileNames("/example/.html"))
	assert.Equal([]string(nil), fn.FileNames("/example/."))

}
