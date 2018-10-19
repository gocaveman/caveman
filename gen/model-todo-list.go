package gen

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

func init() {
	globalMapGenerator["model-todo-list"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		withItem := fset.Bool("with-item", false, "Output the todo-item in addition to todo-list")
		noTypes := fset.Bool("no-types", false, "Do not attempt to add types to store.go")
		targetFile, data, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		targetListFile := targetFile
		targetItemFile := strings.Replace(targetFile, ".go", "-item.go", 1)
		if targetListFile == targetItemFile {
			return fmt.Errorf("cannot figure out item file, does your target file (%q) end with .go?", targetFile)
		}
		if strings.HasSuffix(targetListFile, "-list.go") {
			targetItemFile = strings.Replace(targetListFile, "-list.go", "-item.go", 1)
		}

		data["WithItem"] = *withItem

		listData := make(map[string]interface{}, len(data)+10)
		for k, v := range data {
			listData[k] = v
		}
		itemData := make(map[string]interface{}, len(data)+10)
		for k, v := range data {
			itemData[k] = v
		}

		if !*noTypes {
			// TODO: name of store file should be able to be specified, and possibly some more intelligent guessing could be done;
			// actually, a function that scans a package and figures out what the store file name and struct name is would be really useful
			// for several cases here.
			// modify store.go to register the type
			storeFile := filepath.Join(filepath.Dir(targetFile), "store.go")
			// parse if not already registered - TODO: a function should be added to tmeta for this
			storeMetaStr := "\n"
			storeMetaStr += `if s.Meta.For(TodoList{}) == nil { if err := s.Meta.Parse(TodoList{}); err != nil { return err } }` + "\n"
			if *withItem {
				storeMetaStr += `if s.Meta.For(TodoItem{}) == nil { if err := s.Meta.Parse(TodoItem{}); err != nil { return err } }` + "\n"
			}
			err = GoSrcReplace(s, storeFile, regexp.MustCompile(`// end meta type init`), func(s string) string {
				return storeMetaStr + "\n\n" + s
			})
			if err != nil {
				return err
			}
		}

		// FIXME:
		// from 5.5.3 we can use utf8mb4 as the charset for the table (and thus all the fields)
		// also use utf8_bin for collation specifically on id column(s)
		// do the version with mysql conditional comment /*!50503 whatever */
		// mysql-specific stuff should check {{eq .DriverName "mysql"}} or whatever
		// this should work: /*!50000 CHARACTER SET latin1 COLLATE latin1_general_cs */
		// and this:
		// create_time DATETIME/*!50604 (6) */,
		// update_time DATETIME/*!50604 (6) */,
		// but the basic issue is that the SQL is just different for different databases -
		// e.g. create_time needs to be TEXT for SQLite - we need our DDL generator...
		// and maybe the strategy here is not to try to have Go code that wraps every possible
		// SQL thing, but instead one that handles the common differences easily...
		// So forget about a mysql package with mysql-specific Go calls, instead what it does
		// is it allows you to say "make me a table, with a date column named blah" and
		// it knows that the ID has to be case-insensitive in MySQL, and the date has to be
		// TEXT in sqlite, etc., etc.  And if you need other weird shit, you can always write
		// it by hand.  This would actually be the most useful thing...

		// listData["CreateStatement"] = `CREATE TABLE {{.TablePrefix}}todo_list (
		// 		todo_list_id VARCHAR(64),
		// 		name VARCHAR(255),
		// 		description TEXT,
		// 		version BIGINT,
		// 		create_time TEXT,
		// 		update_time TEXT,
		// 		PRIMARY KEY (todo_list_id)
		// 	)`
		// listData["DropStatement"] = `DROP TABLE {{.TablePrefix}}todo_list`

		err = OutputGoSrcTemplate(s, listData, targetListFile, `
package {{.PackageName}}

import (
	"github.com/bradleypeabody/gouuidv6"
	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/ddl"
)

func TodoListMigrations() (ml migrate.MigrationList) {

	fl := ddl.FormatterList{ddl.NewSQLite3Formatter(true), ddl.NewMySQLFormatter(true)}

	b := ddl.New()
	b.SetCategory("0100_{{.PackageName}}")

	b.SetVersion("0100_Acreate_todo_item")
	b.Up().
		CreateTable("todo_list").
		Column("todo_list_id", ddl.VarCharPK).PrimaryKey().
		Column("name", ddl.VarChar).
		Column("description", ddl.Text).
		Column("version", ddl.BigInt).
		Column("create_time", ddl.DateTime).
		Column("update_time", ddl.DateTime).
		Down().
		DropTable("todo_list")
	b.Up().
		CreateIndex("todo_list_name", "todo_list").Columns("name").
		Down().
		DropIndex("todo_list_name", "todo_list")
	b.MustMigrations(fl...).AppendTo(&ml)

	// more migrations can go here, using the pattern of SetVersion(), statement(s), MustMigrations()...AppendTo()

	return
}

{{/*
func init() {
	if DefaultNameTypes["TodoItem"] == nil {
		DefaultNameTypes["TodoItem"] = reflect.TypeOf(TodoItem{})
		DefaultTableInfoMap.AddTable(TodoItem{}).SetKeys(false, []string{"todo_item_id"})
	}

	// register in migrateregistry and with autowire for all 3 databases
	reg := func(m *migrate.SQLTmplMigration) {
		// TODO: see if we can compact these back down to one-liners
 		var rm migrate.Migration
 		rm = m.NewWithDriverName("sqlite3")
 		DefaultStoreMigrations = append(DefaultStoreMigrations, rm)
 		autowire.Populate(migrateregistry.MustRegister(rm))
 		rm = m.NewWithDriverName("mysql")
 		DefaultStoreMigrations = append(DefaultStoreMigrations, rm)
 		autowire.Populate(migrateregistry.MustRegister(rm))
 		rm = m.NewWithDriverName("postgres")
 		DefaultStoreMigrations = append(DefaultStoreMigrations, rm)
 		autowire.Populate(migrateregistry.MustRegister(rm))
	}

	reg(&migrate.SQLTmplMigration{
		// DriverNameValue set by reg
		CategoryValue: "{{.PackageName}}",
		VersionValue: "0001_todo_item", // must be unique and indicates sequence
		UpSQL: []string{ {{bq .CreateStatement}} },
		DownSQL: []string{ {{bq .DropStatement}} },
	})

	// Add more migration steps (add/remove columns, etc.) here with
	// more calls like: reg(&SQLTmplMigration{...})

}
*/}}

type TodoList struct {
	TodoListID string {{bq "tmeta:\"pk\" db:\"todo_item_id\" json:\"todo_item_id\""}}
	Name string {{bq "db:\"category\" json:\"category\" valid:\"minlen=1\""}} 
	Description string {{bq "db:\"description\" json:\"description\""}}
	Version int64 {{bq "tmeta:\"version\" db:\"version\" json:\"version\""}}
	CreateTime tmetautil.DBTime {{bq "db:\"create_time\" json:\"create_time\""}}
	UpdateTime tmetautil.DBTime {{bq "db:\"update_time\" json:\"update_time\""}}

	{{if .WithItem}}
	// relation(s)
	TodoItemList []TodoItem {{bq "tmeta:\"has_many\" db:\"-\" json:\"todo_item_list\""}}
	{{end}}
}

func (o *TodoList) CreateTimeTouch() { o.CreateTime = tmetautil.NewDBTime() }
func (o *TodoList) UpdateTimeTouch() { o.UpdateTime = tmetautil.NewDBTime() }
func (o *TodoList) IDAssign() { if o.TodoListID == "" { o.TodoListID = gouuidv6.NewB64().String() } }

`, false)
		if err != nil {
			return err
		}

		if *withItem {
			// itemData["CreateStatement"] = `CREATE TABLE {{.TablePrefix}}todo_item (
			// 	todo_item_id VARCHAR(64),
			// 	todo_list_id VARCHAR(64),
			// 	category VARCHAR(255),
			// 	title VARCHAR(255),
			// 	description TEXT,
			// 	PRIMARY KEY (todo_item_id)
			// )`
			// itemData["DropStatement"] = `DROP TABLE {{.TablePrefix}}todo_item`

			err = OutputGoSrcTemplate(s, itemData, targetItemFile, `
package {{.PackageName}}

import (
	"github.com/bradleypeabody/gouuidv6"
	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/ddl"
)

func TodoItemMigrations() (ml migrate.MigrationList) {

	fl := ddl.FormatterList{ddl.NewSQLite3Formatter(true), ddl.NewMySQLFormatter(true)}

	b := ddl.New()
	b.SetCategory("0100_{{.PackageName}}")

	b.SetVersion("0100_Bcreate_todo_item")
	b.Up().
		CreateTable("todo_item").
		Column("todo_item_id", ddl.VarCharPK).PrimaryKey().
		Column("todo_list_id", ddl.VarCharFK).ForiegnKey("todo_list", "todo_list_id").
		Column("line", ddl.Text).
		Column("sequence", ddl.Double).
		Column("version", ddl.BigInt).
		Column("create_time", ddl.DateTime).
		Column("update_time", ddl.DateTime).
		Down().
		DropTable("todo_item")
	b.MustMigrations(fl...).AppendTo(ml)

	// more migrations can go here, using the pattern of SetVersion(), statement(s), MustMigrations()...AppendTo()

	return
}

{{/*
func init() {
	if DefaultNameTypes["TodoItem"] == nil {
		DefaultNameTypes["TodoItem"] = reflect.TypeOf(TodoItem{})
		DefaultTableInfoMap.AddTable(TodoItem{}).SetKeys(false, []string{"todo_item_id"})
	}

	// register in migrateregistry and with autowire for all 3 databases
	reg := func(m *migrate.SQLTmplMigration) {
		// TODO: see if we can compact these back down to one-liners
 		var rm migrate.Migration
 		rm = m.NewWithDriverName("sqlite3")
 		DefaultStoreMigrations = append(DefaultStoreMigrations, rm)
 		autowire.Populate(migrateregistry.MustRegister(rm))
 		rm = m.NewWithDriverName("mysql")
 		DefaultStoreMigrations = append(DefaultStoreMigrations, rm)
 		autowire.Populate(migrateregistry.MustRegister(rm))
 		rm = m.NewWithDriverName("postgres")
 		DefaultStoreMigrations = append(DefaultStoreMigrations, rm)
 		autowire.Populate(migrateregistry.MustRegister(rm))
	}

	reg(&migrate.SQLTmplMigration{
		// DriverNameValue set by reg
		CategoryValue: "{{.PackageName}}",
		VersionValue: "0001_todo_item", // must be unique and indicates sequence
		UpSQL: []string{ {{bq .CreateStatement}} },
		DownSQL: []string{ {{bq .DropStatement}} },
	})

	// Add more migration steps (add/remove columns, etc.) here with
	// more calls like: reg(&SQLTmplMigration{...})

}
*/}}

type TodoItem struct {
	TodoItemID string {{bq "tmeta:\"pk\" db:\"todo_item_id\" json:\"todo_item_id\""}}
	TodoListID string {{bq "db:\"todo_list_id\" json:\"todo_list_id\""}}
	Line string {{bq "db:\"line\" json:\"line\" valid:\"minlen=1\""}} 
	Sequence float64 {{bq "db:\"sequence\" json:\"sequence\""}}

	Version int64 {{bq "tmeta:\"version\" db:\"version\" json:\"version\""}}
	CreateTime tmetautil.DBTime {{bq "db:\"create_time\" json:\"create_time\""}}
	UpdateTime tmetautil.DBTime {{bq "db:\"update_time\" json:\"update_time\""}}

	// relation(s)
	TodoList *TodoList {{bq "tmeta:\"belongs_to\" db:\"-\" json:\"todo_list\""}}
}

func (o *TodoItem) CreateTimeTouch() { o.CreateTime = tmetautil.NewDBTime() }
func (o *TodoItem) UpdateTimeTouch() { o.UpdateTime = tmetautil.NewDBTime() }
func (o *TodoItem) IDAssign() { if o.TodoItemID == "" { o.TodoItemID = gouuidv6.NewB64().String() } }

`, false)
		}
		if err != nil {
			return err
		}

		return nil
	})
}
