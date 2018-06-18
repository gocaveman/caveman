package migrate

import (
	"testing"

	"github.com/gocaveman/caveman/migrate/migratedbr"
	"github.com/gocraft/dbr"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

func writeFile(fs afero.Fs, p string, b []byte) {

	f, err := fs.Create(p)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(b)

}

func writeTestFiles(fs afero.Fs) {

	fs.Mkdir("/data", 0755)

	writeFile(fs, "/data/sqlite3-example1-2017120101_test1-up.sql", []byte(`
CREATE TABLE test1(
	id TEXT,
	PRIMARY KEY(id)
);
ALTER TABLE test1 ADD COLUMN name TEXT;
`))

	writeFile(fs, "/data/sqlite3-example1-2017120101_test1-down.sql", []byte(`
DROP TABLE test1;
`))

	writeFile(fs, "/data/sqlite3-example1-2017120102_test2-up.sql", []byte(`
CREATE TABLE test2(
	id TEXT,
	name TEXT,
	PRIMARY KEY(id)
);
`))

	writeFile(fs, "/data/sqlite3-example1-2017120102_test2-down.sql", []byte(`
DROP TABLE test2;
`))

	writeFile(fs, "/data/sqlite3-example2-2017120103_test3-up.sql", []byte(`
CREATE TABLE test3(
	id TEXT,
	name TEXT,
	PRIMARY KEY(id)
);
`))

	writeFile(fs, "/data/sqlite3-example2-2017120103_test3-down.sql", []byte(`
DROP TABLE test3;
`))

}

func TestLoadSQLMigrationsHFS(t *testing.T) {

	assert := assert.New(t)

	fs := afero.NewMemMapFs()
	writeTestFiles(fs)

	hfs := afero.NewHttpFs(fs)

	ml, err := LoadSQLMigrationsHFS(hfs, "/data")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(3, len(ml.WithDriverName("sqlite3")))
	assert.Equal(2, len(ml.WithDriverName("sqlite3").WithCategory("example1")))
	assert.Equal("2017120101_test1", ml.WithDriverName("sqlite3").WithCategory("example1").Versions()[0])
	assert.Equal("2017120102_test2", ml.WithDriverName("sqlite3").WithCategory("example1").Versions()[1])

	mig := ml.WithDriverName("sqlite3").WithCategory("example1")[0].(*SQLMigration)
	assert.Equal("2017120101_test1", mig.Version())
	assert.Equal(2, len(mig.UpSQL))

}

func TestRunner(t *testing.T) {

	assert := assert.New(t)

	fs := afero.NewMemMapFs()
	writeTestFiles(fs)

	hfs := afero.NewHttpFs(fs)

	ml, err := LoadSQLMigrationsHFS(hfs, "/data")
	if err != nil {
		t.Fatal(err)
	}

	dsn := `file:TestRunner?mode=memory&cache=shared`
	driverName := "sqlite3"

	versioner, err := migratedbr.New(driverName, dsn)
	if err != nil {
		t.Fatal(err)
	}

	runner := NewRunner(driverName, dsn, versioner, ml)

	result, err := runner.CheckAll(true)
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range result {
		if item.Category == "example1" {
			assert.Equal("", item.CurrentVersion)
			assert.Equal("2017120102_test2", item.LatestVersion)
		}
	}

	err = runner.RunAllUpToLatest()
	assert.NoError(err)

	conn, err := dbr.Open(driverName, dsn, nil)
	assert.NoError(err)

	sess := conn.NewSession(nil)

	_, err = sess.InsertInto("test1").Columns("id", "name").Values("k1", "Key 1").Exec()
	assert.NoError(err)
	_, err = sess.InsertInto("test2").Columns("id", "name").Values("k2", "Key 2").Exec()
	assert.NoError(err)

	var name string
	err = sess.Select("name").From("test2").Where("id = ?", "k2").LoadOne(&name)
	assert.NoError(err)

	assert.Equal("Key 2", name)

	// now step example1 down
	err = runner.RunTo("example1", "2017120101_test1")
	assert.NoError(err)

	// make sure test2 errors now
	err = sess.Select("name").From("test2").Where("id = ?", "k2").LoadOne(&name)
	assert.Error(err)

	version := ""
	err = sess.Select("version").From("migration_state").Where("category = ?", "example1").LoadOne(&version)
	assert.NoError(err)
	assert.Equal("2017120101_test1", version)

	// and upgrade it again
	err = runner.RunUpTo("example1", "2017120102_test2")
	assert.NoError(err)
	// and make sure test2 exists now
	_, err = sess.InsertInto("test2").Columns("id", "name").Values("k2", "Key 2").Exec()
	assert.NoError(err)
	err = sess.Select("name").From("test2").Where("id = ?", "k2").LoadOne(&name)
	assert.NoError(err)

}

func TestSQLTmplMigration(t *testing.T) {

	assert := assert.New(t)

	dsn := `file:TestSQLTmplMigration?mode=memory&cache=shared`
	driverName := "sqlite3"

	var ml MigrationList
	ml = append(ml, &SQLTmplMigration{
		DriverNameValue: driverName,
		CategoryValue:   "example1",
		VersionValue:    "0001",
		TablePrefix:     "prefix_",
		UpSQL: []string{`
			CREATE TABLE {{.TablePrefix}}test1(
				id TEXT,
				name TEXT,
				PRIMARY KEY(id)
			);
		`},
		DownSQL: []string{`
			DROP TABLE {{.TablePrefix}}test1;
		`},
	})

	versioner, err := migratedbr.New(driverName, dsn)
	if err != nil {
		t.Fatal(err)
	}

	runner := NewRunner(driverName, dsn, versioner, ml)

	err = runner.RunAllUpToLatest()
	assert.NoError(err)

	conn, err := dbr.Open(driverName, dsn, nil)
	assert.NoError(err)

	sess := conn.NewSession(nil)

	_, err = sess.InsertInto("prefix_test1").Columns("id", "name").Values("k1", "Key 1").Exec()
	assert.NoError(err)
}
