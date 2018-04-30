package gen

import (
	"github.com/spf13/pflag"
)

func init() {
	globalMapGenerator["model-sample-todo-item"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		targetFile, data, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		// FIXME:
		// from 5.5.3 we can use utf8mb4 as the charset for the table (and thus all the fields)
		// also use utf8_bin for collation specifically on id column(s)
		// do the version with mysql conditional comment /*!50503 whatever */
		// mysql-specific stuff should check {{eq .DriverName "mysql"}} or whatever
		data["CreateStatement"] = `
			CREATE TABLE {{.TablePrefix}}todo_item (
				todo_item_id VARCHAR(32),
				category VARCHAR(255),
				title VARCHAR(255),
				description TEXT,
				PRIMARY KEY (todo_item_id)
			);
		`
		data["DropStatement"] = `
			DROP TABLE {{.TablePrefix}}todo_item;
		`

		return OutputGoSrcTemplate(s, data, targetFile, `
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
}
`, false)

	})
}
