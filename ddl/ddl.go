// Easily generate common SQL DDL (data definition language) text or migrations
// using a simple builder.  This supports common DDL operations, not everything,
// and is an attempt to make the common operations easy to do (since the uncommon
// operations are usually database-specific and you can still just write them out
// by hand as SQL).  But most applications just want to create some tables and indexes
// add a column from time to time, and maybe some foriegn keys.  This package allows
// you to do that painlessly for SQLite3, MySQL and Postgres (others possible also
// by implementing the appropriate interfaces).
package ddl

// Notes on things to support:
// CREATE TABLE
// DROP TABLE
// ALTER TABLE RENAME TO (rename table)
// ALTER TABLE ADD COLUMN
// CREATE INDEX
// DROP INDEX
// ALTER TABLE MODIFY COLUMN (change type)
// ALTER TABLE DROP COLUMN
// ALTER TABLE RENAME COLUMN

// NOTE: MSSQL seems to support a ALTER INDEX statement but whatever, supporting create index
// and drop index is fine for now; and we're looking for common functionality across DBs,
// people can always just write their own db-specific stuff without this package to address
// special cases.

// Design note: using structs is wrong - because adding fields to the struct will produce different output - this is bad
// ddl.CreateTableFrom(structHere).PrimaryKey("some_id")

// New returns a new empty Builder.
func New() *Builder {
	return &Builder{}
}

// Builder is a set of up an down migration statements and is used to generate
// SQL via a formatter and optionally produce migrations compatible with migrate.Migration.
// A builder has a set of "up" statements (usually create table, create index, etc.) and
// a separate set of "down" statements (should be the reverse - drop table, drop index, etc.).
type Builder struct {
	Category string
	Version  string

	UpStmtList   StmtList
	DownStmtList StmtList

	down bool // set to true when new statments are down, default is up
}

// Reset will clear all builder fields
func (b *Builder) Reset() *Builder {
	*b = Builder{}
	return b
}

func (b *Builder) pushStmt(stmt Stmt) {
	if !b.down {
		b.UpStmtList = append(b.UpStmtList, stmt)
	} else {
		b.DownStmtList = append(b.DownStmtList, stmt)
	}
}

// MakeSQL will generate and return the up and down SQL statments using
// the formatter provided.
func (b *Builder) MakeSQL(f Formatter) (up []string, down []string, err error) {

	for _, s := range b.UpStmtList {
		sql, ferr := f.Format(s)
		if ferr != nil {
			err = ferr
			return
		}
		up = append(up, sql...)
	}

	for _, s := range b.DownStmtList {
		sql, ferr := f.Format(s)
		if ferr != nil {
			err = ferr
			return
		}
		down = append(down, sql...)
	}

	return
}

// MustMigrations is like Migrations but will panic on error.
func (b *Builder) MustMigrations(formatters ...Formatter) DDLTmplMigrationList {
	ml, err := b.Migrations(formatters...)
	if err != nil {
		panic(err)
	}
	return ml
}

// Migrations will generate and return the migrations (compatible with migrate.Migration)
// that correspond to each of the formatters provided.
// This call will also clear the Builder of everything except for the category,
// so subsequent calls can start fresh on the next migration.
func (b *Builder) Migrations(formatters ...Formatter) (ml DDLTmplMigrationList, err error) {

	for _, f := range formatters {

		var m DDLTmplMigration
		m.DriverNameValue = f.DriverName()
		m.CategoryValue = b.Category
		m.VersionValue = b.Version

		for _, s := range b.UpStmtList {
			sql, ferr := f.Format(s)
			if ferr != nil {
				err = ferr
				return
			}
			m.UpSQL = append(m.UpSQL, sql...)
		}

		for _, s := range b.DownStmtList {
			sql, ferr := f.Format(s)
			if ferr != nil {
				err = ferr
				return
			}
			m.DownSQL = append(m.DownSQL, sql...)
		}

		ml = append(ml, m)

	}

	// if successful, reset everything except category
	*b = Builder{Category: b.Category}

	return
}

// Down makes it so the next statement created is added to the down list.
func (b *Builder) Down() *Builder {
	b.down = true
	return b
}

// Up makes it so the next statement created id assed to the up list.
// This is the same as a Builder in it's initial state, i.e.
// New().Up() and just New() are functionally equivalent.
func (b *Builder) Up() *Builder {
	b.down = false
	return b
}

// SetCategory sets the category string to be included in the migration.
// Usually this is just set once when the Builder is created.  It is
// common for this to be the name of your application or package.
// This method has no effect if you are using MakeSQL() instead of Migrations().
func (b *Builder) SetCategory(category string) *Builder {
	b.Category = category
	return b
}

// SetVersion sets the version string to be included in the migration.
// After Migrations() is called, this field is cleared and needs to be set again,
// since each migration should have it's own version string.  Version strings
// must sort correctly, and the recommened form is as follows: "0001_create_tables",
// and then "0002_add_something_else" for the next migration, and so on.  The number
// serves to make it sort correctly and the rest of the text is purely descriptive.
// This method has no effect if you are using MakeSQL() instead of Migrations().
func (b *Builder) SetVersion(version string) *Builder {
	b.Version = version
	return b
}

// CreateTable will start a CREATE TABLE definition.
func (b *Builder) CreateTable(name string) *CreateTableStmt {
	stmt := &CreateTableStmt{
		Builder:   b,
		NameValue: name,
	}
	b.pushStmt(stmt)
	return stmt
}

// DropTable will start a DROP TABLE definition.
func (b *Builder) DropTable(name string) *DropTableStmt {
	stmt := &DropTableStmt{
		Builder:   b,
		NameValue: name,
	}
	b.pushStmt(stmt)
	return stmt
}

// AlterTableRename will start a ALTER TABLE RENAME TO definition.
func (b *Builder) AlterTableRename(oldName, newName string) *AlterTableRenameStmt {
	stmt := &AlterTableRenameStmt{
		Builder:      b,
		OldNameValue: oldName,
		NewNameValue: newName,
	}
	b.pushStmt(stmt)
	return stmt
}

// AlterTableAdd will start a ALTER TABLE ADD COLUMN definition.
func (b *Builder) AlterTableAdd(tableName string) *AlterTableAddStmt {
	stmt := &AlterTableAddStmt{
		Builder:   b,
		NameValue: tableName,
	}
	b.pushStmt(stmt)
	return stmt
}

// CreateIndex will start a CREATE INDEX definition.
func (b *Builder) CreateIndex(indexName, tableName string) *CreateIndexStmt {
	stmt := &CreateIndexStmt{
		Builder:        b,
		NameValue:      indexName,
		TableNameValue: tableName,
	}
	b.pushStmt(stmt)
	return stmt
}

// DropIndex will start a DROP INDEX definition.
func (b *Builder) DropIndex(indexName, tableName string) *DropIndexStmt {
	stmt := &DropIndexStmt{
		Builder:        b,
		NameValue:      indexName,
		TableNameValue: tableName,
	}
	b.pushStmt(stmt)
	return stmt
}

// Stmt marker interface used to ensure statements are of a valid type.
// (Different statements otherwise have just completely different data
// and no common methods.)
type Stmt interface {
	IsStmt()
}

type StmtList []Stmt
