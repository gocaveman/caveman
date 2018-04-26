package valid

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// NOTE: need to document that fieldName will be the JSON name - NOT the Go name, i.e. "customer_id" not "CustomerID"

type ReadFielder interface {
	ReadField(fieldName string) (interface{}, error)
}

// FIXME: should these really be public?  ReadField?  I can see it both ways but exposing this
// and letting people write code that depends on it may be a mistake.

func ReadField(obj interface{}, fieldName string) (interface{}, error) {

	if obj == nil {
		return nil, ErrNotFound
	}

	if rf, ok := obj.(ReadFielder); ok {
		return rf.ReadField(fieldName)
	}

	objt := reflect.TypeOf(obj)
	for objt.Kind() == reflect.Ptr {
		objt = objt.Elem()
	}
	objv := reflect.ValueOf(obj)
	for objv.Kind() == reflect.Ptr {
		objv = objv.Elem()
	}

	if objt.Kind() == reflect.Map {
		retval := objv.MapIndex(reflect.ValueOf(fieldName))
		if !retval.IsValid() {
			// for map we don't return ErrNotFound just nil
			return nil, nil
		}
		return retval.Interface(), nil
	}

	if objt.Kind() == reflect.Struct {

		sfi := StructFieldIndex(objt, fieldName)
		if sfi < 0 {
			return nil, ErrNotFound
		}

		sfv := objv.Field(sfi)
		return sfv.Interface(), nil
	}

	return nil, fmt.Errorf("invalid type %T", obj)
}

type WriteFielder interface {
	WriteField(fieldName string, v interface{}) error
}

func WriteField(obj interface{}, fieldName string, v interface{}) error {

	if obj == nil {
		return ErrNotFound
	}

	if wf, ok := obj.(WriteFielder); ok {
		return wf.WriteField(fieldName, v)
	}

	objt := reflect.TypeOf(obj)
	for objt.Kind() == reflect.Ptr {
		objt = objt.Elem()
	}
	objv := reflect.ValueOf(obj)
	for objv.Kind() == reflect.Ptr {
		objv = objv.Elem()
	}

	if objt.Kind() == reflect.Map {
		objv.SetMapIndex(reflect.ValueOf(fieldName), reflect.ValueOf(v))
		return nil
	}

	if objt.Kind() == reflect.Struct {

		sfi := StructFieldIndex(objt, fieldName)
		if sfi < 0 {
			return ErrNotFound
		}

		sfv := objv.Field(sfi)
		sfv.Set(reflect.ValueOf(v))

		return nil
	}

	return fmt.Errorf("invalid type %T", obj)
}

func StructFieldIndex(structType reflect.Type, fieldName string) int {

	for structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if fieldName == "" {
		return -1
	}

	convFieldName := ToSnake(fieldName)

	for i := 0; i < structType.NumField(); i++ {

		sf := structType.Field(i)

		// check valid struct tag name=
		vals := StructTagToValues(sf.Tag.Get("valid"))
		name := vals.Get("name")
		if name == fieldName || name == convFieldName {
			return i
		}

		// check json struct tag
		jsonName := strings.SplitN(sf.Tag.Get("json"), ",", 2)[0]
		if jsonName == fieldName || jsonName == convFieldName {
			return i
		}

		// check db struct tag
		dbName := strings.SplitN(sf.Tag.Get("db"), ",", 2)[0]
		if dbName == fieldName || dbName == convFieldName {
			return i
		}

		// check exact match to field name
		if sf.Name == fieldName || sf.Name == convFieldName {
			return i
		}

	}

	return -1

}

// Courtesy of: https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
// ToSnake converts the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func ToSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	// log.Printf("ToSnake(%q) -> %q", in, string(out))
	return string(out)
}
