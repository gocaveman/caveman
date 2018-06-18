package menusdbr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBFileStore(t *testing.T) {

	assert := assert.New(t)

	s := initDBTest(t)
	assert.NotNil(s)

}
