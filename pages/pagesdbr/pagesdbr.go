package pagesdbr

import (
	"github.com/gocaveman/caveman/dbutil"
	"github.com/gocaveman/caveman/pages"
	"github.com/gocraft/dbr"
)

type PageMeta struct {
	Path     string              `db:"path" json:"path"`
	TitleTag string              `db:"title_tag" json:"title_tag"`
	MetaDesc string              `db:"meta_desc" json:"meta_desc"`
	Data     dbutil.StringObjMap `db:"data" json:"data"`
}

func (p *PageMeta) GetTitleTag() string {
	return p.TitleTag
}

func (p *PageMeta) GetMetaDesc() string {
	return p.MetaDesc
}

func (p *PageMeta) GetData() map[string]interface{} {
	return p.Data
}

type Index struct {
	Driver     string
	DSN        string
	Connection *dbr.Connection
	TableName  string
}

func NewIndex(driver, dsn string) (*Index, error) {
	conn, err := dbr.Open(driver, dsn)
	return &Index{
		Driver:     driver,
		DSN:        dsn,
		Connection: conn,
		TableName:  "page",
	}
}

func (i *Index) PageMetaByPath(path string) (pages.PageMeta, error) {
	var pageMeta PageMeta
	sess := i.Connection.NewSession(nil)
	err := sess.Select("*").From(table).Where("path = ?", path).LoadStruct(&pageMeta)
	if err == dbr.ErrNotFound {
		err = dbutil.ErrNotFound
	}
	return &pageMeta, err
}

func (i *Index) PageListByPrefix(prefix string, limit int, startAfterToken string) (result []string, token string, err error) {

	//

}
