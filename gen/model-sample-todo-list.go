package gen

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

func init() {
	globalMapGenerator["model-sample-todo-list"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		noItem := fset.Bool("no-item", false, "Do not output the todo-item, only todo-list")
		targetFile, data, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		// `// end meta type`
		// TODO: text replacement on .../store.go

		targetListFile := targetFile
		targetItemFile := strings.Replace(targetFile, ".go", "-item.go", 1)
		if targetListFile == targetItemFile {
			return fmt.Errorf("cannot figure out item file, does your target file (%q) end with .go?", targetFile)
		}
		if strings.HasSuffix(targetListFile, "-list.go") {
			targetItemFile = strings.Replace(targetListFile, "-list.go", "-item.go", 1)
		}

		listData := make(map[string]interface{}, len(data)+10)
		for k, v := range data {
			listData[k] = v
		}
		itemData := make(map[string]interface{}, len(data)+10)
		for k, v := range data {
			itemData[k] = v
		}

		// modify store.go to register the type
		storeFile := filepath.Join(filepath.Base(targetFile), "store.go")
		storeMetaStr := "\n"
		storeMetaStr += `if err := s.Meta.Parse(TodoList{}); err != nil { return err }` + "\n"
		if !*noItem {
			storeMetaStr += `if err := s.Meta.Parse(TodoItem{}); err != nil { return err }` + "\n"
		}
		err = GoSrcReplace(s, storeFile, regexp.MustCompile(`// end meta type init`), func(s string) string {
			return storeMetaStr + "\n\n" + s
		})
		if err != nil {
			return err
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

		listData["CreateStatement"] = `CREATE TABLE {{.TablePrefix}}todo_list (
				todo_list_id VARCHAR(64),
				name VARCHAR(255),
				description TEXT,
				version BIGINT,
				create_time TEXT,
				update_time TEXT,
				PRIMARY KEY (todo_list_id)
			)`
		listData["DropStatement"] = `DROP TABLE {{.TablePrefix}}todo_list`

		err = OutputGoSrcTemplate(s, listData, targetListFile, `
package {{.PackageName}}

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

type TodoList struct {
	TodoListID string {{bq "db:\"todo_item_id\" json:\"todo_item_id\""}}
	Name string {{bq "db:\"category\" json:\"category\""}} 
	Description string {{bq "db:\"description\" json:\"description\""}}
	Version int64 {{bq "db:\"version\" json:\"version\" tmeta:\"version\""}}
	CreateTime tmetautil.DBTime {{bq "db:\"create_time\" json:\"create_time\""}}
	UpdateTime tmetautil.DBTime {{bq "db:\"update_time\" json:\"update_time\""}}
}

func (o *TodoList) CreateTimeTouch() { o.CreateTime = tmetautil.NewDBTime() }
func (o *TodoList) UpdateTimeTouch() { o.UpdateTime = tmetautil.NewDBTime() }

`, false)
		if err != nil {
			return err
		}

		if *noItem {
			itemData["CreateStatement"] = `CREATE TABLE {{.TablePrefix}}todo_item (
				todo_item_id VARCHAR(64),
				todo_list_id VARCHAR(64),
				category VARCHAR(255),
				title VARCHAR(255),
				description TEXT,
				PRIMARY KEY (todo_item_id)
			)`
			itemData["DropStatement"] = `DROP TABLE {{.TablePrefix}}todo_item`

			err = OutputGoSrcTemplate(s, itemData, targetItemFile, `
package {{.PackageName}}

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

// FIXME: timestamp columns (get the types and setup right both in
// the Go struct and the DDL)

type TodoItem struct {
	TodoItemID string {{bq "db:\"todo_item_id\" json:\"todo_item_id\""}}
	Category string {{bq "db:\"category\" json:\"category\""}} 
	Title string {{bq "db:\"title\" json:\"title\" valid:\"minlen=1\""}}
	Description string {{bq "db:\"description\" json:\"description\""}}

	// FIXME: create and update timestamps and touch methods from tmetadbr
	// FIXME: version column
	// FIXME: probably a "todo_list" and "todo_list_item" or similar would be smart,
	// in order to show basic relation operation
}
`, false)
		}
		if err != nil {
			return err
		}

		return nil
	})
}
