package valid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SampleStruct struct {

	// check valid struct tag name=
	FieldValidName string `valid:"something,name=field_valid_name_here,whatever"`

	// check json struct tag
	FieldJSONName string `json:"field_json_name_here,omitempty"`

	// check db struct tag
	FieldDBName string `json:"field_db_name_here"`

	// check exact match to field name
	FieldExactName string
}

func TestReadFields(t *testing.T) {

	assert := assert.New(t)

	// struct

	sample := &SampleStruct{
		FieldValidName: "value - valid",
		FieldJSONName:  "value - json",
		FieldDBName:    "value - db",
		FieldExactName: "value - exact",
	}

	var v interface{}
	var err error

	v, err = ReadField(sample, "field_valid_name_here")
	assert.NoError(err)
	assert.Equal("value - valid", v)

	v, err = ReadField(sample, "FieldValidNameHere")
	assert.NoError(err)
	assert.Equal("value - valid", v)

	v, err = ReadField(sample, "field_json_name_here")
	assert.NoError(err)
	assert.Equal("value - json", v)

	v, err = ReadField(sample, "field_db_name_here")
	assert.NoError(err)
	assert.Equal("value - db", v)

	v, err = ReadField(sample, "FieldExactName")
	assert.NoError(err)
	assert.Equal("value - exact", v)

	// // should also be available as snake_case
	// v, err = ReadField(sample, "field_exact_name")
	// assert.NoError(err)
	// assert.Equal("value - exact", v)

	v, err = ReadField(sample, "DoesNotExist")
	assert.Error(err)

	// map
	sampleMap := map[string]interface{}{
		"key1": "value1",
	}

	v, err = ReadField(sampleMap, "key1")
	assert.Equal("value1", v)
	assert.NoError(err)
	v, err = ReadField(sampleMap, "does_not_exist")
	assert.Nil(v)
	assert.NoError(err)

}

func TestWriteFields(t *testing.T) {

	assert := assert.New(t)

	sample := &SampleStruct{}

	assert.NoError(WriteField(sample, "field_valid_name_here", "v1"))
	assert.Equal("v1", sample.FieldValidName)

	assert.NoError(WriteField(sample, "FieldValidNameHere", "v2"))
	assert.Equal("v2", sample.FieldValidName)

	assert.NoError(WriteField(sample, "field_json_name_here", "v3"))
	assert.Equal("v3", sample.FieldJSONName)

	assert.NoError(WriteField(sample, "field_db_name_here", "v4"))
	assert.Equal("v4", sample.FieldDBName)

	assert.NoError(WriteField(sample, "FieldExactName", "v5"))
	assert.Equal("v5", sample.FieldExactName)

	assert.Error(WriteField(sample, "DoesNotExist", "whatever"))

	// map
	sampleMap := map[string]interface{}{
		"key1": "value1",
	}
	assert.NoError(WriteField(sampleMap, "key1", "value1a"))
	assert.Equal("value1a", sampleMap["key1"])
	assert.NoError(WriteField(sampleMap, "key2", "value2"))
	assert.Equal("value2", sampleMap["key2"])

}
