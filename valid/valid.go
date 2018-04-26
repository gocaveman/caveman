// Validation rules and functionality for common cases.
//
// TODO: Note on i18n, English strings are provided as the simple, common case, but are not intended to be
// used for translation.  Validation messages also contain Code and Data which together can be used
// with the i18n package to produce properly translated messages in each language.
package valid

import (
	"fmt"
	"reflect"

	"github.com/gocaveman/caveman/webutil"
)

var ErrNotFound = webutil.ErrNotFound

// TODO: figure out how a CheckValid() (or similar) method fits into this.
// While the rules are very useful, the ability to just add custom validation
// code to a struct with a simple function should not be underestimated
// and it should be fully supported.
func Obj(obj interface{}, rules Rules) error {

	objv := reflect.ValueOf(obj)
	for objv.Kind() == reflect.Ptr {
		objv = objv.Elem()
	}

	if objv.Kind() == reflect.Map {

		if rules == nil {
			return fmt.Errorf("valid.Obj() requires rules when validating a map")
		}

	} else if objv.Kind() == reflect.Struct {

		if rules == nil {
			var err error
			rules, err = StructRules(objv.Type(), nil)
			if err != nil {
				return err
			}
		}

	} else {

		return fmt.Errorf("valid.Obj() does not understand objects of type %v", objv.Kind())

	}

	return rules.Apply(obj)
}
