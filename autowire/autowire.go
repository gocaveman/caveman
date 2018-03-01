// Lightweight dependency injection to allow for rapid but unintrusive wiring of components during startup.
package autowire

import (
	"fmt"
	"reflect"
	"strings"
)

var global Wirer

type AfterWirer interface {
	AfterWire() error
}

func Provide(name string, i interface{}) {
	global.Provide(name, i)
}

func Populate(i interface{}) {
	global.Populate(i)
}

func ProvideAndPopulate(name string, i interface{}) {
	global.Provide(name, i)
	global.Populate(i)
}

func Contents() *Wirer {
	var w Wirer
	// deep copy
	w.providers = make([]provider, len(global.providers))
	copy(w.providers, global.providers)
	w.populators = make([]interface{}, len(global.populators))
	copy(w.populators, global.populators)
	return &w
}

type Wirer struct {
	providers  []provider
	populators []interface{}
}

type provider struct {
	Name  string
	Value interface{}
}

func (w *Wirer) Provide(name string, i interface{}) {
	w.providers = append(w.providers, provider{Name: name, Value: i})
}

func (w *Wirer) Populate(i interface{}) {
	w.populators = append(w.populators, i)
}

func (w *Wirer) ProvideAndPopulate(name string, i interface{}) {
	w.Provide(name, i)
	w.Populate(i)
}

func (w *Wirer) Run() error {

	// loop over each populator
	for _, pop := range w.populators {
		vpop := reflect.ValueOf(pop)

		tpop := reflect.TypeOf(pop)

		// deref
		for vpop.Kind() == reflect.Ptr {
			vpop = vpop.Elem()
		}
		for tpop.Kind() == reflect.Ptr {
			tpop = tpop.Elem()
		}

		// loop over each field of the populator
		for i := 0; i < vpop.NumField(); i++ {

			fieldv := vpop.Field(i)

			// don't populate again if it's already there
			if !isZeroOfUnderlyingType(fieldv.Interface()) {
				continue
			}
			// if !fieldv.IsNil() {
			// 	continue
			// }

			// log.Printf("tpop.Kind(): %v", tpop.Kind())

			fieldt := tpop.Field(i)

			_, tagOk := fieldt.Tag.Lookup("autowire")
			if !tagOk {
				continue
			}

			tagParts := strings.Split(fieldt.Tag.Get("autowire"), ",")

			name := tagParts[0]
			optional := false
			if len(tagParts) > 1 {
				if tagParts[1] == "optional" {
					optional = true
				}
			}

			// loop through each provider
			for _, vprov := range w.providers {
				if vprov.Name == name {
					if reflect.TypeOf(vprov.Value).ConvertibleTo(fieldt.Type) {
						fieldv.Set(reflect.ValueOf(vprov.Value))
						break
					}
				}
			}

			if isZeroOfUnderlyingType(fieldv.Interface()) {
				if !optional {
					return fmt.Errorf("object %#v field %q is empty and not marked optional (value=%#v)", pop, fieldt.Name, fieldv.Interface())
				}
			}

		}

	}

	// for anything that implements AfterWire() call it now
	for _, pop := range w.populators {
		if aw, ok := pop.(AfterWirer); ok {
			err := aw.AfterWire()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// see https://stackoverflow.com/questions/13901819/quick-way-to-detect-empty-values-via-reflection-in-go
func isZeroOfUnderlyingType(x interface{}) bool {
	if x == nil {
		return true
	}
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
