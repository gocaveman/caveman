// Lightweight dependency injection to allow for rapid but unintrusive wiring of components during startup.
//
// The basic idea that you have "providers" which are objects that are available to be
// set as values on "populators", which are structs that get populated.  For example,
// a Config struct could be an example of a provider (the Config needs to be provided to
// other things), and a ExampleHandler could be a populator (it needs to be populated).
//
//         type Config struct {
//             // some settings in here
//         }
//
//         type ExampleHandler struct {
//             Config  *Config  `autowire:""`              // no name indicates "default"
//             Prefix  string   `autowire:"api_prefix"`    // TODO: format for names?
//             Debug   bool     `autowire:"debug,optional" // the optional flag means wiring won't fail if not populated
//             Options []string `autowire:"options,slice"  // the slice flag means mulitple values can be assigned and collected here
//         }
//
//         func (h *ExampleHandler)ServeHTTP(w http.ResponseWriter, r *http.Request) {
//             // h.Config, h.Prefix and h.Debug are all available
//         }
//
//         func main() {
//             // ...
//
//             // config is set up here in the main function
//             config := &Config{/* ... */}
//             autowire.Provide("", config) // provide *Config
//             autowire.Provide("api_prefix", "/api") // example of providing a named value
//             // does the actual population, calls any AfterWire() methods, will error
//             // if any non-optional autowire fields on populators don't have values,
//             // or if any AfterWire() methods return an error
//             err := autowire.Contents().Run()
//
//             // ...
//         }
//
//         func init() {
//             // the ExampleHandler could be set up in another package, with no
//             // mention of it in main()
//             exampleHandler := &ExampleHandler{}
//             // tell autowire that our ExampleHandler needs to be populated
//             autowire.Populate(exampleHandler)
//             // ...
//         }
//
// Provide and Populate at the package level call those corresponding methods on a
// global/package level instance of Wirer.  You can also make your own Wirer and
// use it separately as needed.  ProvideAndPopulate() is shorthand for calling both
// functions/methods.
//
package autowire

import (
	"fmt"
	"reflect"
	"strings"
)

var global Wirer

// AfterWirer can be implemented by populators to be called once all population is done.
// It is called from Wirer.Run().
type AfterWirer interface {
	AfterWire() error
}

// Provide calls Wirer.Provide() on the global instance.
func Provide(name string, i interface{}) {
	global.Provide(name, i)
}

// Populate calls Wirer.Populate() on the global instance.
func Populate(i interface{}) {
	global.Populate(i)
}

// ProvideAndPopulate calls Wirer.ProvideAndPopulate() on the global instance.
func ProvideAndPopulate(name string, i interface{}) {
	global.ProvideAndPopulate(name, i)
}

// Contents returns a copy of the global Wirer instance, ready to be Run().
func Contents() *Wirer {
	var w Wirer
	// deep copy the slices
	w.providers = make([]provider, len(global.providers))
	copy(w.providers, global.providers)
	w.populators = make([]interface{}, len(global.populators))
	copy(w.populators, global.populators)
	return &w
}

// NOTE: Making provider type and fields of Wirer public -
// hiding them might have some slight benefit of encouraing not messing
// with the Wirer, but contrasting this with the usefulness of allowing
// easy debug by just dropping a log statement that dumps these things
// out... It probably makes more sense to make this stuff visible.

// FIXME: do this with methods instead - it's useful to expose the data
// and make it modifiable but probably should not expose how it stores
// this info.

// New returns a new empty Wirer instance ready for use.
func New() *Wirer {
	return &Wirer{}
}

// Wirer contains the Providers and Populators, usually there is one for
// the entire application.
type Wirer struct {
	providers  []provider
	populators []interface{}
}

type provider struct {
	name  string
	value interface{}
}

// Name returns the name, implementing Provider interface.
func (p provider) Name() string {
	return p.name
}

// Value returns the value, implementing Provider interface.
func (p provider) Value() interface{} {
	return p.value
}

// Provider represents an instance with a specific name that is provided.
type Provider interface {
	Name() string
	Value() interface{}
}

// Providers returns the list of providers.
func (w *Wirer) Providers() []Provider {
	ret := make([]Provider, 0, len(w.providers))
	for _, p := range w.providers {
		ret = append(ret, p)
	}
	return ret
}

// Populators returns the list of populators.
func (w *Wirer) Populators() []interface{} {
	ret := make([]interface{}, len(w.populators))
	copy(ret, w.populators)
	return ret
}

// Provide add an instance to the list of providers with the specified name.
// Name can and should be an empty string to indicate the default instance
// of this type.  For common types, or types which will have more than one
// instance, specify a meaningful name TODO(bgp): describe format, otherwise
// it's going to get really random without any guidelines.
func (w *Wirer) Provide(name string, i interface{}) {
	w.providers = append(w.providers, provider{name: name, value: i})
}

// Populate specifies a struct with fields to be populated.  Only fields
// with an 'autowire' struct tag are populated.
func (w *Wirer) Populate(i interface{}) {
	w.populators = append(w.populators, i)
}

// ProvideAndPopulate is the same as calling Provide(name, i); Populate(i);
func (w *Wirer) ProvideAndPopulate(name string, i interface{}) {
	w.Provide(name, i)
	w.Populate(i)
}

// Run performs the actual population ("wiring").  Errors are returned
// non-optional fields which cannot be populated or any AfterWire()
// calls which returns errors.  Struct fields which already have a
// non-empty value will not be overwritten (i.e. to assign to an autowire
// field it must begin with it's empty value - this allows you to
// easily override autowire values without causing errors or having
// this method clobber the value you set).
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
			tagO := newTagOpts(tagParts[1:])
			// optional := false
			// if len(tagParts) > 1 {
			// 	if tagParts[1] == "optional" {
			// 		optional = true
			// 	}
			// }

			// loop through each provider
			for _, vprov := range w.providers {
				if vprov.name == name {

					// case where types match
					if reflect.TypeOf(vprov.value).AssignableTo(fieldt.Type) {
						fieldv.Set(reflect.ValueOf(vprov.value))
						break
					}

					// case where it's tagged as a slice and the slice element value matches
					if tagO.HasOpt("slice") && fieldt.Type.Kind() == reflect.Slice {
						if reflect.TypeOf(vprov.value).AssignableTo(fieldt.Type.Elem()) {
							newSlice := reflect.Append(fieldv, reflect.ValueOf(vprov.value))
							fieldv.Set(newSlice)
							continue // for a slice there can be mulitple, keep checking
						}
					}
				}
			}

			if isZeroOfUnderlyingType(fieldv.Interface()) {
				if !tagO.HasOpt("optional") {
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

type tagOpts map[string]bool

func (t tagOpts) HasOpt(n string) bool {
	return t[n]
}

func newTagOpts(sslice []string) tagOpts {
	if len(sslice) == 0 {
		return nil
	}
	ret := make(tagOpts, len(sslice))
	for _, s := range sslice {
		ret[s] = true
	}
	return ret
}

// FIXME: switch to using hidden member vars but provide public accessor methods DONE
// TODO: add "slice" as an option, allowing adding a matching one to a slice
// (we're going to use this instead of registries) - what about the idea of
// just adding a slice itself - so you Populate(&hl) - see if that's viable
// TODO: review if we can do something with the code generator snippets - so
// we can say "insert this snippet in the main wiring section" or maybe it's
// just "add one of these types as provider+populator"
