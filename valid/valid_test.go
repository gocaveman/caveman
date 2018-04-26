package valid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjMap(t *testing.T) {

	assert := assert.New(t)

	var rules Rules
	rules = append(rules, NewMinLenRule("example1", 10))

	m := map[string]interface{}{
		"example1": "Something",
	}

	err := Obj(m, rules)
	msgs, ok := err.(Messages)
	assert.True(ok)
	assert.Len(msgs, 1)
	assert.Equal("minlen", msgs[0].Code)

}

type S1 struct {
	Example1 string `valid:"name=example1,minlen=10"`
}

func TestObjStruct(t *testing.T) {

	assert := assert.New(t)

	s1 := &S1{Example1: "Something"}

	err := Obj(s1, nil)
	msgs, ok := err.(Messages)
	assert.True(ok)
	assert.Len(msgs, 1)
	assert.Equal("minlen", msgs[0].Code)

}

// func TestMapAssign(t *testing.T) {

// 	assert := assert.New(t)

// 	var rules Rules
// 	rules = append(rules, NewMinLenRule("example1", 4))

// 	s1 := &S1{}

// 	err := MapAssign(s1, map[string]interface{}{
// 		"example1": "Value1",
// 		"example0": "Value0",
// 	}, rules)

// 	assert.NoError(err)
// 	assert.Equal("Value1", s1.Example1)

// }

// func TestMapAssignStrict(t *testing.T) {

// 	assert := assert.New(t)

// 	var rules Rules
// 	rules = append(rules, NewMinLenRule("example1", 4))

// 	s1 := &S1{}

// 	err := MapAssignStrict(s1, map[string]interface{}{
// 		"example1": "Value1",
// 		"example0": "Value0",
// 	}, rules)

// 	assert.Contains(err.Error(), "not found on target")

// }

// func TestRequestAssignGET(t *testing.T) {

// 	assert := assert.New(t)

// 	var rules Rules
// 	rules = append(rules, NewMinLenRule("example1", 4))

// 	s1 := &S1{}

// 	req, err := http.NewRequest("GET", "/the-path?example1=Value1&example0=Value0", nil)
// 	assert.NoError(err)

// 	err = RequestAssign(s1, req, rules)
// 	assert.NoError(err)
// 	assert.Equal("Value1", s1.Example1)
// }

// func TestRequestAssignPOSTJSON(t *testing.T) {

// 	assert := assert.New(t)

// 	var rules Rules
// 	rules = append(rules, NewMinLenRule("example1", 4))

// 	s1 := &S1{}

// 	req, err := http.NewRequest("POST", "/the-path", bytes.NewReader([]byte(`{"example1":"Value1"}`)))
// 	req.Header.Set("content-type", "application/json")
// 	assert.NoError(err)

// 	err = RequestAssign(s1, req, rules)
// 	assert.NoError(err)
// 	assert.Equal("Value1", s1.Example1)
// }

// func TestRequestAssignPOSTValues(t *testing.T) {

// 	assert := assert.New(t)

// 	var rules Rules
// 	rules = append(rules, NewMinLenRule("example1", 4))

// 	s1 := &S1{}

// 	req, err := http.NewRequest("POST", "/the-path", bytes.NewReader([]byte(`example1=Value1`)))
// 	req.Header.Set("content-type", "application/x-www-form-urlencoded")
// 	assert.NoError(err)

// 	err = RequestAssign(s1, req, rules)
// 	assert.NoError(err)
// 	assert.Equal("Value1", s1.Example1)
// }
