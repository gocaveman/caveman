package regionctrl

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/migrate/migratedbr"
	"github.com/gocaveman/caveman/regions"
	"github.com/gocaveman/caveman/regions/regionsdb"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	must(err)
	return b
}
func mustUnmarshal(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	must(err)
}
func mustReadAll(in io.Reader) []byte {
	b, err := ioutil.ReadAll(in)
	must(err)
	if rc, ok := in.(io.ReadCloser); ok {
		rc.Close()
	}
	return b
}

func TestRegionController(t *testing.T) {

	assert := assert.New(t)

	driver, dsn := "sqlite3", "file:TestDBStore?mode=memory&cache=shared"

	dbStore := &regionsdb.DBStore{
		DBDriver: driver,
		DBDSN:    dsn,
	}
	assert.NoError(dbStore.AfterWire())

	ver, err := migratedbr.New(driver, dsn)
	assert.NoError(err)
	runner := migrate.NewRunner(driver, dsn, ver, regionsdb.DefaultRegionsMigrations)
	assert.NoError(runner.RunAllUpToLatest())

	// assert.NoError(dbStore.WriteDefinition(regions.Definition{
	// 	DefinitionID: "def1",
	// 	RegionName:   "leftnav+",
	// 	TemplateName: "/example1.gohtml",
	// }))

	store := &regionsdb.DBStore{DBDriver: driver, DBDSN: dsn}
	must(store.AfterWire())
	h := NewRegionController(store)

	// create one
	wrec := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/api/region", bytes.NewReader(mustMarshal(regions.Definition{
		DefinitionID: "def1",
		RegionName:   "leftnav+",
		TemplateName: "/example1.gohtml",
	})))
	r.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(wrec, r)
	assert.Equal(200, wrec.Result().StatusCode)

	// overwrite it
	wrec = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/api/region", bytes.NewReader(mustMarshal(regions.Definition{
		DefinitionID: "def1",
		RegionName:   "leftnav+",
		TemplateName: "/example1a.gohtml",
	})))
	r.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(wrec, r)
	assert.Equal(200, wrec.Result().StatusCode)

	// add another one
	wrec = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/api/region", bytes.NewReader(mustMarshal(regions.Definition{
		DefinitionID: "def2",
		RegionName:   "leftnav+",
		TemplateName: "/example2.gohtml",
	})))
	r.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(wrec, r)
	assert.Equal(200, wrec.Result().StatusCode)

	// get list and check it
	wrec = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/region", nil)
	h.ServeHTTP(wrec, r)
	assert.Equal(200, wrec.Result().StatusCode)
	var defList regions.DefinitionList
	mustUnmarshal(mustReadAll(wrec.Result().Body), &defList)
	assert.Len(defList, 2)
	assert.Equal("def1", defList[0].DefinitionID)
	assert.Equal("def2", defList[1].DefinitionID)

	// delete a record
	wrec = httptest.NewRecorder()
	r, _ = http.NewRequest("DELETE", "/api/region/def1", nil)
	h.ServeHTTP(wrec, r)
	assert.Equal(200, wrec.Result().StatusCode)

	// make sure it doesn't show any more
	wrec = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/api/region", nil)
	h.ServeHTTP(wrec, r)
	assert.Equal(200, wrec.Result().StatusCode)
	defList = nil
	mustUnmarshal(mustReadAll(wrec.Result().Body), &defList)
	assert.Len(defList, 1)
	assert.Equal("def2", defList[0].DefinitionID)

}
