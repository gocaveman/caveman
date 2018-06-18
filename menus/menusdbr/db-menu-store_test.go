package menusdbr

import (
	"testing"

	"github.com/gocaveman/caveman/menus"
	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/migrate/migratedbr"
	"github.com/gocraft/dbr"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

func initDBTest(t *testing.T) *DBMenuStore {

	assert := assert.New(t)

	dsn := `file:TestDBMenuStore?mode=memory&cache=shared`
	driver := "sqlite3"

	conn, err := dbr.Open(driver, dsn, nil)
	assert.NoError(err)

	ver, err := migratedbr.New(driver, dsn)
	assert.Nil(err)

	runner := migrate.NewRunner(driver, dsn, ver, Migrations)
	err = runner.RunAllUpToLatest()
	assert.Nil(err)

	s := &DBMenuStore{
		Connection: conn,
	}

	return s
}

func TestDBMenuStore(t *testing.T) {

	assert := assert.New(t)

	s := initDBTest(t)

	err := s.CreateMenuItem(&menus.MenuItem{
		ParentMenuID: "",
		MenuID:       "example1",
		Title:        "The Title Here",
		Enabled:      true,
	})
	assert.NoError(err)

	mi, err := s.ReadMenuItem("example1")
	assert.NoError(err)
	assert.Equal("The Title Here", mi.Title)

	mi = &menus.MenuItem{
		ParentMenuID: "",
		MenuID:       "example1",
		Title:        "New Title",
		Enabled:      true,
	}
	err = s.UpdateMenuItem(mi)
	assert.NoError(err)

	err = s.DeleteMenuItem("example1")
	assert.NoError(err)

	err = s.CreateMenuItem(&menus.MenuItem{MenuID: "example1"})
	assert.NoError(err)
	err = s.CreateMenuItem(&menus.MenuItem{MenuID: "example2"})
	assert.NoError(err)
	err = s.CreateMenuItem(&menus.MenuItem{MenuID: "example3"})
	assert.NoError(err)

	ids, err := s.FindChildren("")
	assert.NoError(err)
	assert.Equal([]string{"example1", "example2", "example3"}, ids)

}
