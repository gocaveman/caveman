package valid

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRules(t *testing.T) {

	assert := assert.New(t)

	m := map[string]interface{}{
		"field1": "test",
		"field2": nil,
		"field3": "test3",
		"field4": "joey@example.com",
		"field5": 50,
	}

	assert.Error(NewMinLenRule("field1", 5).Apply(m))
	assert.NoError(NewMinLenRule("field1", 4).Apply(m))

	assert.Error(NewMaxLenRule("field1", 3).Apply(m))
	assert.NoError(NewMaxLenRule("field1", 4).Apply(m))

	assert.Error(NewNotNilRule("field2").Apply(m))

	assert.NoError(NewRegexpRule("field3", regexp.MustCompile(`^test3$`)).Apply(m))
	assert.NoError(NewRegexpRule("field3", regexp.MustCompile(`3$`)).Apply(m))
	assert.Error(NewRegexpRule("field3", regexp.MustCompile(`^blah$`)).Apply(m))

	assert.NoError(NewEmailRule("field4").Apply(m))
	assert.Error(NewEmailRule("field3").Apply(m))

	assert.NoError(NewMinValRule("field5", 40).Apply(m))
	assert.Error(NewMinValRule("field5", 60).Apply(m))

	assert.NoError(NewMaxValRule("field5", 60).Apply(m))
	assert.Error(NewMaxValRule("field5", 40).Apply(m))

}
