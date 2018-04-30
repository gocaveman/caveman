package gen

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/pflag"
)

func init() {
	globalMapGenerator["store-crud"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		storeName := fset.String("store", "Store", "The name of the store struct to add methods to.")
		modelName := fset.String("model", "", "The model object name, if not specified default will be deduced from file name.")
		genericMode := fset.Bool("generic", false, "Generic mode outputs methods that use interface{} instead of the specific type and can have the underlying model object swapped out.")
		tests := fset.Bool("tests", true, "Create test file with test(s) for these store methods.")
		targetFile, data, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		data["StoreName"] = *storeName
		data["GenericMode"] = *genericMode

		modelNameFixed := *modelName
		if modelNameFixed == "" {
			_, fname := path.Split(targetFile)
			modelNameFixed = NameSnakeToCamel(fname, []string{"store-"}, nil)
		}
		data["ModelName"] = modelNameFixed

		err = OutputGoSrcTemplate(s, data, targetFile, `
package {{.PackageName}}

import (
 	"github.com/bradleypeabody/gouuidv6"
 	"github.com/gocaveman/caveman/valid"
)

// New{{.ModelName}} returns a new empty instance of the appropriate type.
func (s *{{.StoreName}}) New{{.ModelName}}() {{if .GenericMode}}interface{}{{else}}*{{.ModelName}}{{end}} {
	{{if .GenericMode}} return reflect.New(s.NameTypes["{{.ModelName}}"]).Addr().Interface()
	{{else}} return &{{.ModelName}}{}
	{{end}} }

// Insert{{.ModelName}} inserts the record into the database.
func (s *{{.StoreName}}) Insert{{.ModelName}}(o {{if .GenericMode}}interface{}{{else}}*{{.ModelName}}{{end}}) error {
 	// FIXME: need UUID only for non-auto-inc case, should we move gouuidv6 into caveman?  maybe gocaveman/gouuidv6
 	if o.{{.ModelName}}ID == "" {
 		o.{{.ModelName}}ID = gouuidv6.NewB64().String()
 	}
 	err := valid.Obj(o, nil)
 	if err != nil {
 		return err
 	}
 	sess := s.NewSession(nil)
 	return sess.ObjInsert(o)
}

// Update{{.ModelName}} updates this record in the database.
func (s *{{.StoreName}}) Update{{.ModelName}}(o *{{.ModelName}}) error {
	// TODO: version field code for when we have that working...
	err := valid.Obj(o, nil)
	if err != nil {
		return err
	}
	sess := s.NewSession(nil)
	return sess.ObjUpdate(o)
}

// TODO: update diff? (probably not a bad idea, see if it makes sense with how the controller is done)

// Delete{{.ModelName}} deletes this record from the database.  The key can
// either be in the instance provided, or specified separately.
func (s *{{.StoreName}}) Delete{{.ModelName}}(o *{{.ModelName}}, pk ...interface{}) error {
	sess := s.NewSession(nil)
	return sess.ObjDelete(o, pk...)
}

// Get{{.ModelName}} fetches a record from the database.  The key can
// either be in the instance provided, or specified separately.
// If the record cannot be found, dbr.ErrNotFound will be returned.
func (s *{{.StoreName}}) Get{{.ModelName}}(o *{{.ModelName}}, pk ...interface{}) error {
	sess := s.NewSession(nil)
	return sess.ObjGet(o, pk...)
}

// TODO: figure out Find... methods...

// TODO: upsert? - maybe it's an option to add an example if desired.

`, false)
		if err != nil {
			return err
		}

		if *tests {

			testsTargetFile := strings.Replace(targetFile, ".go", "_test.go", 1)
			if testsTargetFile == targetFile {
				return fmt.Errorf("unable to determine test file name for %q", targetFile)
			}

			err = OutputGoSrcTemplate(s, data, testsTargetFile, `
package {{.PackageName}}

func Test{{.ModelName}}CRUD(t *testing.T) {

	assert := assert.New(t)

	dbDriver := "sqlite3"
	dbDsn := "file:Test{{.ModelName}}CRUD?mode=memory&cache=shared"

	ml := DefaultStoreMigrations.WithDriverName(dbDriver).Sorted()
	versioner, err := migratedbr.New(dbDriver, dbDsn)
	assert.NoError(err)
	runner := migrate.NewRunner(dbDriver, dbDsn, versioner, ml)
	err = runner.RunAllUpToLatest()
	assert.NoError(err)

	conn, err := dbr.Open(dbDriver, dbDsn, nil)
	assert.NoError(err)
	store := NewStore(dbrobj.NewConfig().NewConnector(conn, nil))

	err = store.AfterWire()
	assert.NoError(err)

	o := &{{.ModelName}}{
		// TODO: populate with valid data
		// Category:    "testcat",
		// Title:       "Clean the Kitchen",
		// Description: "It's gross",
	}
	err = store.Insert{{.ModelName}}(o)
	assert.NoError(err)
	// TODO: o.Title = "Deep Clean the Kitchen"
	err = store.Update{{.ModelName}}(o)
	assert.NoError(err)
	o2 := &{{.ModelName}}{}
	err = store.Get{{.ModelName}}(o2, o.{{.ModelName}}ID)
	assert.NoError(err)
	// TODO: assert.Equal("Deep Clean the Kitchen", o2.Title)
	err = store.Delete{{.ModelName}}(o2)
	assert.NoError(err)
	err = store.Get{{.ModelName}}(o2)
	assert.Error(err)

}

`, false)
			if err != nil {
				return err
			}

		}

		return nil
	})
}
