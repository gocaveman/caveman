// Easily create database-specific Data Definition Language SQL.
// The point of this package is to make it easy to generate DDL in a way that is
// 1) resistant to typos and syntax errors (Go function calls are checked at
// compile time and can often be auto-completed by your editor, not so with SQL strings),
// and 2) avoids the necessity of having to write out database-specific syntax for equivalent
// functionality (Example: booleans are TINYINT, BOOLEAN or NUMERIC for MySQL, Postgres or SQLite
// respectively - and while you could customize the type definition for each one, most of the
// time you don't want to, you just conceptually want to say "give me whatever this particular
// database uses for a boolean".)
package ddl

// concept: you create a general statement using the generic stuff - e.g. create table,
// and it auto generates all kinds of stuff, things like setting the pks you can do
// generically, but then you need to set the type of a particular column for mysql,
// and you call MySQL() (or perhaps ddlmysql.On(...)) and then can perform mysql-specific
// mods that apply only to MySQL.

// TODO: we still need to support {{.TablePrefix}} in these create table statements, think through
// TODO: and how do we generally integrate with migrate - I'm thinking we delay the generation
// of the SQL until the last possible time, at which point TablePrefix and DriverName and anything
// else are available - but how do we get these values without depending on the migrate package...

// Things to support:
// CREATE TABLE
// DROP TABLE
// ALTER TABLE ADD COLUMN
// ALTER TABLE MODIFY COLUMN (change type)
// ALTER TABLE DROP COLUMN
// ALTER TABLE RENAME COLUMN
// ALTER TABLE RENAME TO (rename table)
// CREATE INDEX
// DROP INDEX

// NOTE: MSSQL seems to support a ALTER INDEX statement but whatever, supporting create index
// and drop index is fine for now; and we're looking for common functionality across DBs,
// people can always just write their own db-specific stuff without this package to address
// special cases.

/*

// hm, wrong - because adding fields to the struct will produce different output - this is bad
ddl.CreateTableFrom(structHere).PrimaryKey("some_id")

// Nope...
options := &ddl.Options{TablePrefix: "{{.TablePrefix}}",DriverName: "mysql"}
// Not when we can make Options an interface... that is implemented by SQLTmplMigration
// or maybe we need a new one, crap but then we have the dual interface problem ...
// Okay check it - we can implement migrate.Migration ourselves...  it's no big deal and
// gets rid of the dual interface problem - and we can do magical things like automatically
// generate a down "drop" statement when the "create" is done... whoa! crazy (note that this
// is harder to do with a column type change where we only know the target type not the original -
// to do this we'd need to know about the earlier create table statement and its type there)

// so they have to list out the fields, kinda annoying but with cavegen to give full working
// examples and the fact that everything is autocompleted, it's managable
ddl.CreateTable("name").ColInt("a_num").ColVarchar("something", 128).ColText("textcol").ColBlob("blobcol")

ddlmysql.Use(ddlsqlite3.Use(ddlpostgres.Use(ddl.CreateTable()...))).List()

stmt := ddl.CreateTable()...
ml = append(ml, ddlmysql.Use(stmt), ddlsqlite3.Use(stmt), ddlpostgres.Use(stmt))

// ---

ddlgen := ddlpostgres.NewGen(ddlsqlite3.NewGen(ddlmysql.NewGen(ddl.NewGen())))

var stmt *ddl.Stmt

ddl.Append(&ml, ddlgen.CreateTable("example").ColInt("id"))
ddl.Append(&ml, ddlmysql.New(ddlgen.CreateTable("example").ColInt("id")).ColBool("other"))

so maybe the object we start with is not ddlgen, but something that conceptualy is
an entire list of changes - so we can know about earlier statemtns (e.g. prior field type
to generate our 'down' for a column type change)

*/

// type Z struct{}

// func (z Z) Do() {}

// type A interface {
// 	Do()
// }

// type B interface {
// 	Do()
// }

// type AList []A
// type BList []B

// func F(b ...B) {}

// func init() {

// 	al := AList{Z{}, Z{}}
// 	// var bl BList
// 	// bl = BList([]B([]A(al)))

// 	// log.Printf("bl: %v", bl)
// 	F(al[0], al[1])

// }

// type Statement struct {
// }

// type CreateTableStmt struct {
// }

// func (s *CreateTableStmt)

// type Generator struct {
// }

// func NewGenerator() *Generator {
// 	return &Generator{}
// }

// func (g *Generator) CreateTable(name string) *Generator {
// }

type DataType interface {
	DataTypeString()
}

type DefDataType string

type CreateTableStmt interface {
	ColVarchar(name string, length int) CreateTableStmt
}
