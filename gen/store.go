package gen

import (
	"flag"
)

func init() {
	globalMapGenerator["store"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := flag.NewFlagSet("gen", flag.ContinueOnError)
		storeName := fset.String("name", "Store", "The name of the store struct to define")
		// tableMarshal := fset.Bool("table-marshal", false, "Adds functionality for marshaling tables to/from a flat files.")
		targetFile, data, err := ParseFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		data["StoreName"] = *storeName

		// FIXME: we should probably add a test file and make it define the environment
		// in which this store (and the other store-crud stuif) is tested.  For example,
		// some people will want to run the tests against their MySQL database, others
		// might be fine with the default sqlite stuff, but either way you definitely don't
		// want to have to set it up separately for each model object, make more sense
		// at the store level.

		return OutputGoSrcTemplate(s, data, targetFile, `
package {{.PackageName}}

import (
	"github.com/gocaveman/caveman/migrate"
)

// TODO: figure out if DefaultTableInfoMap should have "{{.StoreName}}" in it,
// thus making it so we can put multiple stores in the same package with out conflict...

// DefaultTableInfoMap is the default database config info for the various tables.
var DefaultTableInfoMap = make(dbrobj.TableInfoMap)

// DefaultNameTypes is a map that tells us which specific struct a CamelCase name corresponds to
var DefaultNameTypes = make(map[string]reflect.Type)

// Default{{.StoreName}} is the default instance that is registered with autowire.
var Default{{.StoreName}} *{{.StoreName}} = New{{.StoreName}}(nil)

// Default{{.StoreName}}Migrations is all of our migrations for this store.
var Default{{.StoreName}}Migrations migrate.MigrationList

func init() {
	autowire.ProvideAndPopulate("", Default{{.StoreName}})
}

func New{{.StoreName}}(c *dbrobj.Connector) *{{.StoreName}} {
	return &{{.StoreName}}{
		Connector: c,
		NameTypes: make(map[string]reflect.Type),
	}
}

// {{.StoreName}} provides database storage methods.
type {{.StoreName}} struct {
	*dbrobj.Connector {{bq "autowire:\"\""}}

	NameTypes map[string]reflect.Type // map logical model names to actual structs
}

func (s *{{.StoreName}}) AfterWire() error {

	// bring in default table info stuff, allowing for overrides
	// TODO: can we make this more concise?

	if s.NameTypes == nil {
		s.NameTypes = make(map[string]reflect.Type)
	}
	for k, v := range DefaultNameTypes {
		if s.NameTypes[k] == nil {
			s.NameTypes[k] = v
		}
	}

	if s.TableInfoMap == nil {
		s.TableInfoMap = make(dbrobj.TableInfoMap)
	}
	for k, v := range DefaultTableInfoMap {
		if s.TableInfoMap[k] == nil {
			s.TableInfoMap[k] = v
		}
	}

	return nil
}

{{/* HOLD ON THIS FOR NOW - LET'S GET A REAL USE CASE AND SEE HOW IMPORTANT
     IT IS TO INCLUDE THIS IN THE TEMPLATE AS AN OPTION.

func (s *{{.StoreName}}) loadFlatFile() error {
	// If enabled...
}

// TODO: we also need the cmdline flag and option(s) to put in the calls 
// to loadFlatFile and storeFlatFile in the individual store-crud stuff,
// at the appropriate points.

// TODO: what happens if we externalize this and make it an interface, similar
// to how controllers have callbacks for things, we could make a callback mechanism
// that gets plugged into in order to implement the flat file persistence.
// FlatFileLoader, FlatFileStorer, FlatFileLoadStorer - hm, but something needs
// to know which table(s) and the connection, etc.  Seems like the coupling may be too
// tight for this approach to make sense.  Possibly better to just provide a utility
// function to do the dirty work with one method call, and thus making it easy to
// implement loadFlatFile and storeFlatFile.  Think through more...

// TODO: we also should have a mode which is read-only, storeFlatFile() doesn't get
// called (or is a nop).  Possibly having a flatFileReadOnly flag would be applicable
// in these cases.

func (s *{{.StoreName}}) storeFlatFile() error {
	// If enabled, write the specified types (from command line flags, checked for in TableInfoMap)
	// out to whatever mechanism we have.

	// what about: (both being present enables)
	// flatFileFS FileSystem
	// flatFilePath string

}

*/}}

`, false)

	})
}
