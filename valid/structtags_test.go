package valid

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type T1 struct {
	F1 string `valid:"name=field1name,notnil,minlen=5,maxlen=10"`
}

type T2 struct {
	F2 string `valid:"name=f2,regexp=^test2$"`
}

type T3 struct {
	F3 string `valid:"minlen=5"` // no name in the valid:"..." part of struct tag
}

type T4 struct {
	F4 string `valid:"email"`
}

type T5 struct {
	F5 int `valid:"minval=10,maxval=20"`
}

func TestStructTags(t *testing.T) {

	assert := assert.New(t)

	t1 := &T1{
		F1: "test",
	}

	rules, err := StructRules(reflect.TypeOf(t1), nil)
	assert.NoError(err)
	// log.Printf("rules=%#v", rules)

	err = rules.Apply(t1)
	// log.Printf("err=%#v", err)

	msgserr, _ := err.(Messages)
	assert.NotNil(msgserr)
	assert.Len(msgserr, 1)
	assert.Equal("minlen", msgserr[0].Code)

	t1.F1 = "testtesttest"
	assert.True(rules.Apply(t1).(Messages).ContainsCode("maxlen"))

	var t2 *T2
	rules, err = StructRules(reflect.TypeOf(t2), nil)
	assert.NoError(err)
	t2 = &T2{F2: "test2"}
	assert.NoError(rules.Apply(t2))
	t2 = &T2{F2: "nomatch"}
	// log.Printf("HERE: %#v", rules.Apply(t2))
	assert.True(rules.Apply(t2).(Messages).ContainsCode("regexp"))

	var t3 *T3
	rules, err = StructRules(reflect.TypeOf(t3), nil)
	assert.NoError(err)
	// log.Printf("rules: %#v", rules)
	t3 = &T3{F3: "test3"}
	assert.NoError(rules.Apply(t3))

	var t4 *T4
	rules, err = StructRules(reflect.TypeOf(t4), nil)
	assert.NoError(err)
	t4 = &T4{F4: "test4@example.com"}
	assert.NoError(rules.Apply(t4))
	t4 = &T4{F4: "test4@example"}
	assert.Error(rules.Apply(t4))

	var t5 *T5
	rules, err = StructRules(reflect.TypeOf(t5), nil)
	// log.Printf("rules: %#v", rules)
	assert.NoError(err)
	t5 = &T5{F5: 15}
	assert.NoError(rules.Apply(t5))
	t5 = &T5{F5: 5}
	assert.Error(rules.Apply(t5))
	assert.True(rules.Apply(t5).(Messages).ContainsCode("minval"))
	t5 = &T5{F5: 25}
	assert.Error(rules.Apply(t5))
	assert.True(rules.Apply(t5).(Messages).ContainsCode("maxval"))

}
