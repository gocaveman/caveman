package ddl

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"reflect"
)

type DDLTmplMigrationList []DDLTmplMigration

// AppendTo is a helper to append a list of migrations to any slices whose element type
// is compatible with (*DDLTmplMigration).  This makes it a single simple function call
// to add your DDLTmplMigrations to a migrate.MigrationList.
// The argument must be a pointer to a slice, any incorrect type (not a pointer to
// a slice or element type not compatible) will result in a panic.
func (l DDLTmplMigrationList) AppendTo(slicePtr interface{}) {
	vptr := reflect.ValueOf(slicePtr)
	if vptr.Kind() != reflect.Ptr {
		panic("value provided is not a pointer")
	}
	vd := vptr.Elem() // deref pointer
	if vd.Kind() != reflect.Slice {
		panic("value provided does not point to a slice")
	}

	newSliceV := vd
	for i := range l {
		newSliceV = reflect.Append(newSliceV, reflect.ValueOf(&l[i]))
	}
	vd.Set(newSliceV)
}

// DDLTmplMigration implements the migrate.Migration interface and supports templates.
// It is does the same thing as migrate.SQLTmplMigration but it's here in the ddl
// package so we don't have unnecessary dependencies.
type DDLTmplMigration struct {
	DriverNameValue string
	CategoryValue   string
	VersionValue    string
	UpSQL           []string
	DownSQL         []string

	// a common reason to use DDLTmplMigration is be able to configure the table prefix
	TablePrefix string `autowire:"db.TablePrefix,optional"`
	// other custom data needed by the template(s) can go here
	Data interface{}
}

func (m *DDLTmplMigration) DriverName() string { return m.DriverNameValue }
func (m *DDLTmplMigration) Category() string   { return m.CategoryValue }
func (m *DDLTmplMigration) Version() string    { return m.VersionValue }

func (m *DDLTmplMigration) tmplExec(dsn string, stmts []string) error {

	db, err := sql.Open(m.DriverNameValue, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	for n, s := range stmts {

		t := template.New("sql")
		t, err := t.Parse(s)
		if err != nil {
			return fmt.Errorf("DDLTmplMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) template parse on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, s)
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, m)
		if err != nil {
			return fmt.Errorf("DDLTmplMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) template execute on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, s)
		}

		newS := buf.String()

		_, err = db.Exec(newS)
		if err != nil {
			return fmt.Errorf("DDLTmplMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) Exec on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, newS)
		}
	}

	return nil
}

// ExecUp will run the up migration statements.
func (m *DDLTmplMigration) ExecUp(dsn string) error {
	return m.tmplExec(dsn, m.UpSQL)
}

// ExecUp will run the down migration statements.
func (m *DDLTmplMigration) ExecDown(dsn string) error {
	return m.tmplExec(dsn, m.DownSQL)
}
