package menusdbr

import (
	"github.com/gocaveman/caveman/autowire"
	"github.com/gocaveman/caveman/menus"
	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/migrate/migrateregistry"
	"github.com/gocaveman/caveman/webutil"
	"github.com/gocraft/dbr"
)

var Migrations migrate.MigrationList

func init() {

	// register in migrateregistry and with autowire for all 3 databases
	reg := func(m *migrate.SQLTmplMigration) {
		var rm migrate.Migration
		rm = m.NewWithDriverName("sqlite3")
		Migrations = append(Migrations, rm)
		autowire.Populate(migrateregistry.MustRegister(rm))
		rm = m.NewWithDriverName("mysql")
		Migrations = append(Migrations, rm)
		autowire.Populate(migrateregistry.MustRegister(rm))
		rm = m.NewWithDriverName("postgres")
		Migrations = append(Migrations, rm)
		autowire.Populate(migrateregistry.MustRegister(rm))
	}

	reg(&migrate.SQLTmplMigration{
		CategoryValue: "menusdbr",
		VersionValue:  "0001_menus_create", // must be unique and indicates sequence
		UpSQL: []string{`
			CREATE TABLE {{.TablePrefix}}menu (
				menu_id VARCHAR(255),
				parent_menu_id VARCHAR(255),
				sequence DOUBLE,
				title VARCHAR(255),
				meta TEXT,
				enabled INTEGER,
				registry_edited INTEGER,
				PRIMARY KEY (menu_id)
			)
		`},
		DownSQL: []string{`DROP TABLE {{.TablePrefix}}menu`},
	})

	reg(&migrate.SQLTmplMigration{
		CategoryValue: "menusdbr",
		VersionValue:  "0002_menus_indexes", // must be unique and indicates sequence
		UpSQL: []string{`
			CREATE INDEX {{.TablePrefix}}menu_parent ON {{.TablePrefix}}menu (parent_menu_id)
		`},
		DownSQL: []string{`
			DROP INDEX {{.TablePrefix}}menu_parent ON {{.TablePrefix}}menu
		`},
	})

}

// FIXME: timestamp columns (get the types and setup right both in
// the Go struct and the DDL)

type DBMenuItem struct {
	menus.MenuItem
	RegistryEdited bool `db:"registry_edited"`
}

type DBMenuStore struct {
	Connection  *dbr.Connection // `autowire:""`
	TablePrefix string          `autowire:"db_table_prefix,optional"`
}

func (s *DBMenuStore) ReadMenuItem(id string) (*menus.MenuItem, error) {
	sess := s.Connection.NewSession(nil)
	var dbmi DBMenuItem
	err := sess.Select("*").From(s.TablePrefix+"menu").Where("menu_id=?", id).LoadOne(&dbmi)
	if err != nil {
		if err == dbr.ErrNotFound {
			return nil, menus.ErrNotFound
		}
		return nil, err
	}
	return &dbmi.MenuItem, nil
}

func (s *DBMenuStore) CreateMenuItem(mi *menus.MenuItem) error {
	sess := s.Connection.NewSession(nil)
	var dbmi DBMenuItem
	dbmi.MenuItem = *mi
	// TODO: registry integration
	_, err := sess.InsertInto(s.TablePrefix+"menu").
		Columns("menu_id", "parent_menu_id", "sequence", "title", "meta", "enabled", "registry_edited").
		Values(dbmi.MenuID, dbmi.ParentMenuID, dbmi.Sequence, dbmi.Title, webutil.SimpleStringDataMap(dbmi.Meta), dbmi.Enabled, dbmi.RegistryEdited).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (s *DBMenuStore) UpdateMenuItem(mi *menus.MenuItem) error {
	sess := s.Connection.NewSession(nil)
	var dbmi DBMenuItem
	dbmi.MenuItem = *mi
	// TODO: registry integration
	res, err := sess.Update(s.TablePrefix+"menu").
		Set("menu_id", dbmi.MenuID).
		Set("parent_menu_id", dbmi.ParentMenuID).
		Set("sequence", dbmi.Sequence).
		Set("title", dbmi.Title).
		Set("meta", webutil.SimpleStringDataMap(dbmi.Meta)).
		Set("enabled", dbmi.Enabled).
		Set("registry_edited", dbmi.RegistryEdited).
		Exec()
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return menus.ErrNotFound
	}
	return nil
}

func (s *DBMenuStore) DeleteMenuItem(id string) error {
	sess := s.Connection.NewSession(nil)

	res, err := sess.DeleteFrom(s.TablePrefix+"menu").Where("menu_id=?", id).Exec()
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return menus.ErrNotFound
	}
	return nil

}

func (s *DBMenuStore) FindChildren(id string) ([]string, error) {
	sess := s.Connection.NewSession(nil)
	var ret []string
	_, err := sess.Select("menu_id").From(s.TablePrefix+"menu").Where("parent_menu_id=?", id).Load(&ret)
	return ret, err
}
