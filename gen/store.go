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

// Default{{.StoreName}} is the default instance that is registered with autowire.
var Default{{.StoreName}} *{{.StoreName}} = New{{.StoreName}}("sqlite3", nil)

// Default{{.StoreName}}Migrations is all of our migrations for this store.
var Default{{.StoreName}}Migrations migrate.MigrationList

func init() {
	autowire.ProvideAndPopulate("", Default{{.StoreName}})
}

func New{{.StoreName}}(driverName string, db *sql.DB) *{{.StoreName}} {
	return &{{.StoreName}}{
		DriverName: driverName,
		DB: db,
		Meta: tmeta.NewMeta(),
	}
}

// {{.StoreName}} provides database storage methods.
type {{.StoreName}} struct {
	DriverName string {{bq "autowire:\"driver_name\""}}
	*sql.DB {{bq "autowire:\"\""}}
	*tmeta.Meta {{bq "autowire:\"\""}}

	dbrc *dbr.Connection
	dbrmu sync.RWMutex
}

func (s *{{.StoreName}}) dbrConn() *dbr.Connection {
	s.dbrmu.RLock()
	dbrc := s.dbrc
	s.dbrmu.RUnlock()

	if dbrc == nil {
		return dbrc
	}

	s.dbrmu.Lock()
	defer s.dbrmu.Unlock()

	dbrc = &dbr.Connection{
		DB: s.DB,
    	EventReceiver: &dbr.NullEventReceiver{},
	}
	switch s.DriverName {
	case "sqlite3":
		dbrc.Dialect = dialect.SQLite3
	case "mysql":
		dbrc.Dialect = dialect.MySQL
	case "postgres":
		dbrc.Dialect = dialect.PostgreSQL
	default:
		panic("unknown driver: "+s.DriverName)
	}

	s.dbrc = dbrc

	return dbrc
}

func (s *{{.StoreName}}) AfterWire() error {

	// TODO: figure out Meta overrides (so 3p packages can override the struct used for a particular table)
	if s.Meta == nil {
		s.Meta = tmeta.New()
	}

	// begin meta type init

	// end meta type init

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
