package migratedbr

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

func TestDbrVersioner(t *testing.T) {

	assert := assert.New(t)

	dsn := `file:TestDbrVersioner?mode=memory&cache=shared`

	tmpDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpDB.Close()
	tmpDB.Exec("DROP TABLE migration_state")

	mv, err := New("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}

	v, err := mv.Version("cat1")
	assert.NoError(err)
	assert.Equal("", v)

	err = mv.StartVersionChange("cat1", "")
	assert.NoError(err)

	err = mv.EndVersionChange("cat1", "2017120102")
	assert.NoError(err)

	v, err = mv.Version("cat1")
	assert.NoError(err)
	assert.Equal("2017120102", v)

	err = mv.StartVersionChange("cat1", "2017120102")
	assert.NoError(err)

	err = mv.EndVersionChange("cat1", "2017120103")
	assert.NoError(err)

	cats, err := mv.Categories()
	assert.NoError(err)
	assert.Equal([]string{"cat1"}, cats)

}
