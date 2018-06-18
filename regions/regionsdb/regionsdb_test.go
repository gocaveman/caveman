package regionsdb

import (
	"log"
	"testing"

	"github.com/gocaveman/caveman/filesystem/aferofs"
	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/migrate/migratedbr"
	"github.com/gocaveman/caveman/regions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

func TestDBStore(t *testing.T) {

	assert := assert.New(t)

	fs := afero.NewMemMapFs()
	fs.MkdirAll("/tmp", 0755)

	driver, dsn := "sqlite3", "file:TestDBStore?mode=memory&cache=shared"

	dbStore := &DBStore{
		DBDriver:       driver,
		DBDSN:          dsn,
		FileSystem:     aferofs.New(fs),
		FilePathPrefix: "/tmp",
	}
	assert.NoError(dbStore.AfterWire())

	ver, err := migratedbr.New(driver, dsn)
	assert.NoError(err)
	runner := migrate.NewRunner(driver, dsn, ver, DefaultRegionsMigrations)
	assert.NoError(runner.RunAllUpToLatest())

	assert.NoError(dbStore.WriteDefinition(regions.Definition{
		DefinitionID: "def1",
		RegionName:   "leftnav+",
		TemplateName: "/example1.gohtml",
	}))

	b, err := afero.ReadFile(fs, "/tmp/regions.yaml")
	assert.NoError(err)

	log.Printf("regions: %s", b)

}
