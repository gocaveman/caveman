package webutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasPathPrefix(t *testing.T) {

	assert := assert.New(t)

	assert.True(HasPathPrefix("/example", "/example"))
	assert.True(HasPathPrefix("/example/", "/example"))
	assert.True(HasPathPrefix("/example/something", "/example"))
	assert.True(HasPathPrefix("/example/something/else", "/example"))
	assert.False(HasPathPrefix("/example1", "/example"))
	assert.False(HasPathPrefix("/other/example", "/example"))

}

func TestPathParse(t *testing.T) {

	assert := assert.New(t)

	i := 0
	s := ""

	// basic case
	assert.Nil(PathParse("/something/123/abc", "/something/%d/%s", &i, &s))
	assert.Equal(123, i)
	assert.Equal("abc", s)

	// basic failed case
	assert.NotNil(PathParse("/nothing/123/abc", "/something/%d/%s", &i, &s))

	assert.Nil(PathParse("/something/123/else/abc", "/something/%d/else/%s", &i, &s))

	// trailing slash, should still work
	assert.Nil(PathParse("/something/123/else/abc", "/something/%d/else/%s/", &i, &s))

	// path does not include all elements
	assert.NotNil(PathParse("/something/123/else/abc", "/something/%d/else/%s/what", &i, &s))

	// path has too many elements
	assert.NotNil(PathParse("/something/123/else/abc", "/something/%d/else", &i))

	// path has too many elements
	assert.NotNil(PathParse("/something/123/else/abc", "/something/%d", &i))

}
