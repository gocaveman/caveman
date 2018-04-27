package httpapi

import (
	"fmt"
	"reflect"
	"strings"
)

var errNotFillable = fmt.Errorf("field not fillable")

func Fill(dst, src interface{}) (reterr error) {

	// catch any of the reflect panic stuff and return it as an error
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				reterr = rerr
			}
			reterr = fmt.Errorf("caught panic: %v", r)
		}
	}()

	srcObj := reflect.ValueOf(src)
	for srcObj.Kind() == reflect.Ptr {
		srcObj = srcObj.Elem()
	}
	srcType := srcObj.Type()

	if srcType.Kind() == reflect.Map {

		keyVals := srcObj.MapKeys()
		for _, kv := range keyVals {
			kstr := fmt.Sprintf("%v", kv.Interface())

			vv := srcObj.MapIndex(kv).Interface()

			err := fillField(dst, kstr, vv)
			// just skip unfillable fields
			if err == errNotFillable {
				continue
			}
			if err != nil {
				return err
			}

		}

		return nil

	} else if srcType.Kind() == reflect.Struct {

		for i := 0; i < srcType.NumField(); i++ {
			f := srcType.Field(i)

			// use snake version of Go field name or httpapi tag if present
			fieldName := toSnake(f.Name)
			if fn := strings.SplitN(f.Tag.Get("httpapi"), ",", 2)[0]; fn != "" {
				fieldName = fn
			}

			fieldVal := srcObj.Field(i).Interface()

			err := fillField(dst, fieldName, fieldVal)
			// just skip unfillable fields
			if err == errNotFillable {
				continue
			}
			if err != nil {
				return err
			}

		}

		return nil

	}

	return fmt.Errorf("cannot fill from type %T", src)
}

func fillField(dst interface{}, fieldName string, val interface{}) error {

	dstObj := reflect.ValueOf(dst)
	for dstObj.Kind() == reflect.Ptr {
		dstObj = dstObj.Elem()
	}
	dstType := dstObj.Type()

	if dstType.Kind() == reflect.Map {

		// for maps we don't check fillability
		dstObj.SetMapIndex(reflect.ValueOf(fieldName), reflect.ValueOf(val))
		return nil

	} else if dstType.Kind() == reflect.Struct {

		// for struct fields we need to find a matching field with "fillable" set
		for i := 0; i < dstType.NumField(); i++ {
			f := dstType.Field(i)

			// name matches if specified as first part of tag or first part of tag is empty and toSnake
			// version of field name matches
			tagParts := strings.Split(f.Tag.Get("httpapi"), ",")
			if !(tagParts[0] == fieldName || (tagParts[0] == "" && toSnake(f.Name) == fieldName)) {
				continue
			}

			// make sure there's at least one option
			if len(tagParts) < 2 {
				continue
			}

			// check fillable
			fillable := false
			for _, p := range tagParts[1:] {
				if p == "fillable" {
					fillable = true
				}
			}
			if !fillable {
				return errNotFillable
			}

			// do the actual assignment
			dstObj.Field(i).Set(reflect.ValueOf(val))

		}

		return nil

	}

	return fmt.Errorf("cannot fill to type %T", dst)

}

// // NOTE: need to document that fieldName will be the JSON name - NOT the Go name, i.e. "customer_id" not "CustomerID"

// type readFielder interface {
// 	ReadField(fieldName string) (interface{}, error)
// }

// func readField(obj interface{}, fieldName string) (interface{}, error) {

// 	if obj == nil {
// 		return nil, ErrNotFound
// 	}

// 	if rf, ok := obj.(readFielder); ok {
// 		return rf.ReadField(fieldName)
// 	}

// 	objt := reflect.TypeOf(obj)
// 	for objt.Kind() == reflect.Ptr {
// 		objt = objt.Elem()
// 	}
// 	objv := reflect.ValueOf(obj)
// 	for objv.Kind() == reflect.Ptr {
// 		objv = objv.Elem()
// 	}

// 	if objt.Kind() == reflect.Map {
// 		retval := objv.MapIndex(reflect.ValueOf(fieldName))
// 		if !retval.IsValid() {
// 			// for map we don't return ErrNotFound just nil
// 			return nil, nil
// 		}
// 		return retval.Interface(), nil
// 	}

// 	if objt.Kind() == reflect.Struct {

// 		sfi, _ := structFieldIndex(objt, fieldName)
// 		if sfi < 0 {
// 			return nil, ErrNotFound
// 		}

// 		sfv := objv.Field(sfi)
// 		return sfv.Interface(), nil
// 	}

// 	return nil, fmt.Errorf("invalid type %T", obj)
// }

// type writeFielder interface {
// 	WriteField(fieldName string, v interface{}) error
// }

// func writeField(obj interface{}, fieldName string, v interface{}) error {

// 	if obj == nil {
// 		return ErrNotFound
// 	}

// 	if wf, ok := obj.(writeFielder); ok {
// 		return wf.WriteField(fieldName, v)
// 	}

// 	objt := reflect.TypeOf(obj)
// 	for objt.Kind() == reflect.Ptr {
// 		objt = objt.Elem()
// 	}
// 	objv := reflect.ValueOf(obj)
// 	for objv.Kind() == reflect.Ptr {
// 		objv = objv.Elem()
// 	}

// 	// map fields are always fillable
// 	if objt.Kind() == reflect.Map {
// 		objv.SetMapIndex(reflect.ValueOf(fieldName), reflect.ValueOf(v))
// 		return nil
// 	}

// 	if objt.Kind() == reflect.Struct {

// 		sfi, fillable := structFieldIndex(objt, fieldName)
// 		if sfi < 0 {
// 			return ErrNotFound
// 		}

// 		sfv := objv.Field(sfi)
// 		sfv.Set(reflect.ValueOf(v))

// 		return nil
// 	}

// 	return fmt.Errorf("invalid type %T", obj)
// }

// func structFieldIndex(structType reflect.Type, fieldName string) (int, bool) {

// 	for structType.Kind() == reflect.Ptr {
// 		structType = structType.Elem()
// 	}

// 	if fieldName == "" {
// 		return -1, false
// 	}

// 	convFieldName := toSnake(fieldName)

// 	for i := 0; i < structType.NumField(); i++ {

// 		sf := structType.Field(i)

// 		parts := strings.Split(sf.Tag.Get("httpapi"), ",")

// 		if parts[0] == convFieldName {

// 			if len(parts) == 1 {
// 				return i, false
// 			}

// 			for _, attr := range parts[1:] {
// 				if attr == "fillable" {
// 					return i, true
// 				}
// 			}

// 			return i, false

// 		}

// 		// TODO: do we really want to be looking at these other options - not sure if this
// 		// makes sense...

// 		// check json struct tag
// 		jsonName := strings.SplitN(sf.Tag.Get("json"), ",", 2)[0]
// 		if jsonName == fieldName || jsonName == convFieldName {
// 			return i, false
// 		}

// 		// check db struct tag
// 		dbName := strings.SplitN(sf.Tag.Get("db"), ",", 2)[0]
// 		if dbName == fieldName || dbName == convFieldName {
// 			return i, false
// 		}

// 		// check exact match to field name
// 		if sf.Name == fieldName || sf.Name == convFieldName {
// 			return i, false
// 		}

// 	}

// 	return -1, false

// }

// // Courtesy of: https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
// // toSnake converts the given string to snake case following the Golang format:
// // acronyms are converted to lower-case and preceded by an underscore.
// func toSnake(in string) string {
// 	runes := []rune(in)
// 	length := len(runes)

// 	var out []rune
// 	for i := 0; i < length; i++ {
// 		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
// 			out = append(out, '_')
// 		}
// 		out = append(out, unicode.ToLower(runes[i]))
// 	}

// 	return string(out)
// }
