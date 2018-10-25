package weberrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeberrors(t *testing.T) {

	assert := assert.New(t)

	// make sure regular errors act as expected
	baseErr := fmt.Errorf("base error")
	assert.Error(baseErr)
	assert.Equal(0, ErrorCode(baseErr))
	assert.Equal("", ErrorMessage(baseErr))
	assert.Nil(ErrorData(baseErr))

	// now one with weberrors.Detail
	err := New(baseErr, 501, "some error", "some data")
	assert.Error(err)
	assert.Equal(501, ErrorCode(err))
	assert.Equal("some error", ErrorMessage(err))
	assert.Equal("some data", ErrorData(err))

	// now wrap that with ErrLoc and make sure it still works
	errLoc := ErrLoc(err)
	assert.Error(errLoc)
	assert.Equal(501, ErrorCode(errLoc))
	assert.Equal("some error", ErrorMessage(errLoc))
	assert.Equal("some data", ErrorData(errLoc))

	errPrefix := ErrPrefix("test error: ", errLoc)
	assert.Error(errPrefix)
	assert.Equal(501, ErrorCode(errLoc))
	assert.Contains(errPrefix.Error(), "test error: ")

	// check RootCause
	assert.Equal(baseErr, RootCause(errPrefix))
	assert.Equal(baseErr, RootCause(errLoc))
	assert.Equal(baseErr, RootCause(err))
	assert.Equal(baseErr, RootCause(baseErr))
	assert.Nil(RootCause(nil))
}
