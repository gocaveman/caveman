package file2sqlite

import (
	"testing"

	"github.com/gocaveman/caveman/dbutil"
	"github.com/jinzhu/gorm"

	_ "github.com/mattn/go-sqlite3"
)

// need a pattern of building Page (or whatever) structs out of pages - Meta should probably
// go in as a JSON field (to be figured out - marshaling is a bit difficult, since Meta should
// also be able to be replaced if needed), Page struct
// should of course have a default but also dev should be able to completely replace struct
// with no issues - if they want more fields or to change stuff, default should implement
// category and tags but this is just a basic ref impl - user could embed this struct and
// add more fields or as above just replace entirely - Go template code would rely on field
// names so if no Category field then those bits that reference it would fail, but otherwise
// things should work correctly; this is a special case but is really just a complex example
// of real life usage of file2sqlite

type Page struct {
	ID    int
	Title string
	Path  string
	Tags  dbutil.StringValueList `gorm:"type:text"`
}

func TestF2S(t *testing.T) {

	connstr := `file:file2sqlite?mode=memory&cache=shared`

	// rawdb, err := sql.Open("sqlite3", connstr)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer rawdb.Close()

	// _, err = rawdb.Exec(`CREATE TABLE foo (id INTEGER NOT NULL PRIMARY KEY, name TEXT)`)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	db, err := gorm.Open("sqlite3", connstr)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if errs := db.AutoMigrate(&Page{}).GetErrors(); len(errs) > 0 {
		t.Fatal(errs)
	}

	if errs := db.Create(&Page{ID: 0, Title: "example 123", Path: "/example123", Tags: []string{"tag1", "tag2"}}).GetErrors(); len(errs) > 0 {
		t.Fatal(errs)
	}

	p := Page{}
	db.First(&p, 1)

	t.Logf("p = %#v", p)

}
