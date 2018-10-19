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
 	"context"

 	"github.com/bradleypeabody/gouuidv6"
 	"github.com/gocaveman/caveman/valid"
 	"github.com/gocaveman/tmeta/tmetadbr"
 	"github.com/gocaveman/tmeta/tmetautil"
 	"github.com/gocraft/dbr"
)

// Create{{.ModelName}} inserts the record into the database.
func (s *{{.StoreName}}) Create{{.ModelName}}(ctx context.Context, o {{if .GenericMode}}interface{}{{else}}*{{.ModelName}}{{end}}) error {

	tx, err := s.dbrc.NewSession(s.EventReceiver).BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = valid.Obj(o, nil)
	if err != nil {
		return err
	}

	b := tmetadbr.New(tx, s.Meta)
	_, err = b.MustInsert(o).Exec()
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Update{{.ModelName}} updates this record in the database.
func (s *{{.StoreName}}) Update{{.ModelName}}(ctx context.Context, o {{if .GenericMode}}interface{}{{else}}*{{.ModelName}}{{end}}) error {

	tx, err := s.dbrc.NewSession(s.EventReceiver).BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	err = valid.Obj(o, nil)
	if err != nil {
		return err
	}

	b := tmetadbr.New(tx, s.Meta)
	err = b.ResultWithOneUpdate(b.MustUpdateByID(o).Exec())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete{{.ModelName}} deletes this record in the database.
func (s *{{.StoreName}}) Delete{{.ModelName}}(ctx context.Context, o {{if .GenericMode}}interface{}{{else}}*{{.ModelName}}{{end}}) error {

	tx, err := s.dbrc.NewSession(s.EventReceiver).BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	b := tmetadbr.New(tx, s.Meta)
	err = b.ResultWithOneUpdate(b.MustDeleteByID(o).Exec())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Fetch{{.ModelName}} get a record in the database by ID.
func (s *{{.StoreName}}) Fetch{{.ModelName}}(ctx context.Context, o {{if .GenericMode}}interface{}{{else}}*{{.ModelName}}{{end}}, id string, related ...string) error {

	tx, err := s.dbrc.NewSession(s.EventReceiver).BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	b := tmetadbr.New(tx, s.Meta)
	err = b.MustSelectByID(o, id).LoadOne(o)
	if err != nil {
		return err
	}

	ti := s.Meta.For(o)
	for _, r := range related {
		rstmt, err := b.SelectRelation(o, r)
		if err != nil {
			return err
		}
		_, err = rstmt.Load(
			ti.RelationTargetPtr(o, r))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

{{/* FIXME: security goes in controller - so move the strict check of the
fields on criteria and orderby and related etc up up there
*/}}

func (s *{{.StoreName}}) Search{{.ModelName}}Count(ctx context.Context, criteria tmetautil.Criteria, orderBy tmetautil.OrderByList, maxRows int64) (int64, error) {

	tx, err := s.dbrc.NewSession(s.EventReceiver).BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.RollbackUnlessCommitted()

	ti := s.Meta.For({{.ModelName}}{})

	stmt := tx.Select(ti.SQLPKFields()...).From(ti.SQLName())

	whereSql, args, err := criteria.SQL()
	if err != nil {
		return 0, err
	}
	if len(whereSql) > 0 {
		stmt = stmt.Where(whereSql, args...)
	}

	for _, o := range orderBy {
		stmt = stmt.OrderDir(o.Field, !o.Desc)
	}

	if maxRows >= 0 {
		stmt = stmt.Limit(uint64(maxRows))
	}

	buf := dbr.NewBuffer()
	err = stmt.Build(s.dbrc.Dialect, buf)
	if err != nil {
		return 0, err
	}
	innerSQL := buf.String()
	args = buf.Value()

	var c int64
	err = tx.SelectBySql("SELECT count(1) c FROM ("+innerSQL+") t", args...).LoadOne(&c)
	if err != nil {
		return 0, err
	}

	return c, tx.Commit()
}

// Search{{.ModelName}} builds and runs a select statement from the input you provide.
func (s *{{.StoreName}}) Search{{.ModelName}}(ctx context.Context, criteria tmetautil.Criteria, orderBy tmetautil.OrderByList, limit, offset int64, related ...string) ([]{{.ModelName}}, error) {

	tx, err := s.dbrc.NewSession(s.EventReceiver).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUnlessCommitted()

	b := tmetadbr.New(tx, s.Meta)

	ti := s.Meta.For({{.ModelName}}{})

	stmt := tx.Select(ti.SQLFields(true)...).From(ti.SQLName())

	whereSql, args, err := criteria.SQL()
	if err != nil {
		return nil, err
	}
	if len(whereSql) > 0 {
		stmt = stmt.Where(whereSql, args...)
	}

	for _, o := range orderBy {
		stmt = stmt.OrderDir(o.Field, !o.Desc)
	}

	if offset > 0 {
		stmt = stmt.Offset(uint64(offset))
	}
	if limit >= 0 {
		stmt = stmt.Limit(uint64(limit))
	}

	// empty set should return zero length slice instead of nil for proper JSON output and semantic correctness
	ret := make([]{{.ModelName}}, 0)
	_, err = stmt.Load(&ret)
	if err != nil {
		return nil, err
	}

	if len(related) > 0 {
		for i := range ret {
			for _, r := range related {
				rstmt, err := b.SelectRelation(&ret[i], r)
				if err != nil {
					return nil, err
				}
				_, err = rstmt.Load(
					ti.RelationTargetPtr(&ret[i], r))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return ret, tx.Commit()
}


{{/* NOTE: the paging/limits here are all based on the idea of not overloading the database server and putting sensible limits on how
	much data a client can ask for. However it is completely feasible for clients to page through an arbitrarily large data set
	by asking for the next N records where KEY > LAST_KEY, and clients don't need any special tools for that particularly
	- this should definitely be documented somewhere */}}


// TODO: figure out Find... methods... (think about listing page); also make sure empty
// list returns zero length slice and not nil, so JSON marshaling etc work as expected
// (it's also, while less efficient, semantically correct - the search didn't return "nothing"
// it returned zero of the specified element - and this difference also shows up in
// the resulting JSON)

// TODO: upsert? - maybe it's an option to add an example if desired.

// TODO: related/joins; also specifically look at having methods that say "set the
// list of this type of join to exactly X set as one call in one transaction", rather
// than having to delete and re-write.

`, false)
		if err != nil {
			return err
		}

		// TODO: disable tests for now until we fill this in
		if false && *tests {

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
	err = store.Create{{.ModelName}}(o)
	assert.NoError(err)
	// TODO: o.Title = "Deep Clean the Kitchen"
	err = store.Update{{.ModelName}}(o)
	assert.NoError(err)
	o2 := &{{.ModelName}}{}
	err = store.Fetch{{.ModelName}}(o2, o.{{.ModelName}}ID)
	assert.NoError(err)
	// TODO: assert.Equal("Deep Clean the Kitchen", o2.Title)
	err = store.Delete{{.ModelName}}(o2)
	assert.NoError(err)
	err = store.Fetch{{.ModelName}}(o2)
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
