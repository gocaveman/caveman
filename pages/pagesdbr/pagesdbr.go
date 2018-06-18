package pagesdbr

import (
	"fmt"

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
	conn, err := dbr.Open(driver, dsn, nil)
	if err != nil {
		return nil, err
	}
	return &Index{
		Driver:     driver,
		DSN:        dsn,
		Connection: conn,
		TableName:  "page",
	}, nil
}

func (i *Index) PageMetaByPath(path string) (ret pages.PageMeta, reterr error) {
	// var pageMeta PageMeta
	// sess := i.Connection.NewSession(nil)
	// err := sess.Select("*").From("").Where("path = ?", path).LoadOne(&pageMeta)
	// if err == dbr.ErrNotFound {
	// 	err = dbutil.ErrNotFound
	// }
	// return &pageMeta, err
	return ret, fmt.Errorf("not implemented")
}

func (i *Index) PageListByPrefix(prefix string, limit int, startAfterToken string) (result []string, token string, err error) {
	return nil, "", fmt.Errorf("not implemented")
}
