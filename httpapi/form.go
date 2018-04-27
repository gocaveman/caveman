package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"unicode"
)

func FormUnmarshal(values url.Values, obj interface{}) error {

	j, err := FormToJSON(values, obj)
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, obj)
	if err != nil {
		return fmt.Errorf("json.Unmarshal() error: %v; json: %s", err, j)
	}

	return nil

}

func FormToJSON(values url.Values, obj interface{}) ([]byte, error) {

	objType := reflect.TypeOf(obj)
	for objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	if objType.Kind() == reflect.Struct {
		return formStructToJSON(values, obj)
	}

	return formOtherToJSON(values, obj)

}

func formStructToJSON(origValues url.Values, obj interface{}) ([]byte, error) {

	// handle values so that "field*" works
	values := make(url.Values, len(origValues))
	for k, v := range origValues {
		k2, _ := formMultiName(k)
		values[k] = v
		values[k2] = v
	}

	objType := reflect.TypeOf(obj)
	for objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	var buf bytes.Buffer
	buf.WriteString("{")
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	// loop over the struct fields
	for i := 0; i < objType.NumField(); i++ {

		var strs []string

		f := objType.Field(i)

		outFname := ""

		if v2, ok := values[f.Name]; ok {
			strs = v2
			outFname = f.Name
		}

		jname := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if jname != "" {
			// if json name is set to '-' we explicitly skip
			if jname == "-" {
				continue
			}
			v2, ok := values[jname]
			if ok {
				strs = v2
				outFname = jname
			}
		}

		if strs != nil && len(strs) > 0 {

			// special case where value is empty and field can be nullable
			if (len(strs) == 1 && strs[0] == "") &&
				(f.Type.Kind() == reflect.Ptr ||
					f.Type.Kind() == reflect.Struct ||
					f.Type.Kind() == reflect.Interface) {
				enc.Encode(outFname)
				buf.WriteRune(':')
				buf.WriteString("null")
				buf.WriteRune(',')
				continue
			}

			// deref pointer on field type
			ftype := f.Type
			for ftype.Kind() == reflect.Ptr {
				ftype = ftype.Elem()
			}

			// log.Printf("f.Name=%q, ftype=%v", f.Name, ftype)

			// how we output the value depends a lot on the type of the field
			switch {

			case ftype.Kind() == reflect.Uint8 ||
				ftype.Kind() == reflect.Uint16 ||
				ftype.Kind() == reflect.Uint32 ||
				ftype.Kind() == reflect.Uint64 ||
				ftype.Kind() == reflect.Uint:

				var v uint64
				_, err := fmt.Sscanf(strs[0], "%d", &v)
				if err != nil {
					return nil, fmt.Errorf("unsigned integer parse failed: %v", err)
				}

				enc.Encode(outFname)
				buf.WriteRune(':')
				err = enc.Encode(v)
				if err != nil {
					return nil, err
				}
				buf.WriteRune(',')

			case ftype.Kind() == reflect.Int8 ||
				ftype.Kind() == reflect.Int16 ||
				ftype.Kind() == reflect.Int32 ||
				ftype.Kind() == reflect.Int64 ||
				ftype.Kind() == reflect.Int:

				var v int64
				_, err := fmt.Sscanf(strs[0], "%d", &v)
				if err != nil {
					return nil, fmt.Errorf("integer parse failed: %v", err)
				}

				enc.Encode(outFname)
				buf.WriteRune(':')
				err = enc.Encode(v)
				if err != nil {
					return nil, err
				}
				buf.WriteRune(',')

			case ftype.Kind() == reflect.Float32 ||
				ftype.Kind() == reflect.Float64:

				var v float64
				_, err := fmt.Sscanf(strs[0], "%f", &v)
				if err != nil {
					return nil, fmt.Errorf("float parse failed: %v", err)
				}

				enc.Encode(outFname)
				buf.WriteRune(':')
				err = enc.Encode(v)
				if err != nil {
					return nil, err
				}
				buf.WriteRune(',')

			case ftype.Kind() == reflect.String ||
				ftype.Kind() == reflect.Interface ||
				(ftype.Kind() == reflect.Slice && ftype.Elem().Kind() == reflect.Uint8):

				enc.Encode(outFname)
				buf.WriteRune(':')
				err := enc.Encode(strs[0])
				if err != nil {
					return nil, err
				}
				buf.WriteRune(',')

			case ftype.Kind() == reflect.Slice && ftype.Elem().Kind() == reflect.String:

				enc.Encode(outFname)
				buf.WriteRune(':')
				buf.WriteRune('[')

				for i := range strs {
					err := enc.Encode(strs[i])
					if err != nil {
						return nil, err
					}
					if i < len(strs)-1 {
						buf.WriteRune(',')
					}
				}

				buf.WriteRune(']')
				buf.WriteRune(',')

			default:
				// if any other type field will just get skipped

			}

		}

	}

	b := append(bytes.TrimSuffix(buf.Bytes(), []byte(`,`)), []byte(`}`)...)
	b = bytes.Replace(b, []byte("\n"), nil, -1) // remove the new lines that Encode() adds

	return b, nil

}

func formOtherToJSON(values url.Values, obj interface{}) ([]byte, error) {

	objType := reflect.TypeOf(obj)
	for objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// check for element types that require single or multi behavior
	forceSingle := objType.Elem().Kind() == reflect.String
	forceMulti := objType.Elem().Kind() == reflect.Slice

	var buf bytes.Buffer
	buf.WriteString("{")
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	for k, vals := range values {

		k, multi := formMultiName(k)

		if (forceSingle || !multi) && !forceMulti {
			enc.Encode(k)
			buf.WriteRune(':')
			enc.Encode(vals[0])
			buf.WriteRune(',')
		} else {
			enc.Encode(k)
			buf.WriteString(":[")
			for i := range vals {
				enc.Encode(vals[i])
				if i < len(vals)-1 {
					buf.WriteString(",")
				}
			}
			buf.WriteString("],")
		}
	}

	b := append(bytes.TrimSuffix(buf.Bytes(), []byte(`,`)), []byte(`}`)...)
	b = bytes.Replace(b, []byte("\n"), nil, -1) // remove the new lines that Encode() adds

	return b, nil

}

func formMultiName(k string) (kret string, isMulti bool) {

	if strings.HasSuffix(k, "*") {
		return strings.TrimSuffix(k, "*"), true
	}

	// not sure if we should support "[]" as a suffix, like PHP - having two ways to do it is
	// probably not a good thing, even if [] would be more familiar to some people
	// if strings.HasSuffix(k, "[]") {
	// 	return strings.TrimSuffix(k, "[]"), true
	// }

	return k, false
}

// Courtesy of: https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
// toSnake converts the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func toSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
