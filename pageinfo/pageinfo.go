package pageinfo

import (
	"strings"

	"github.com/gocaveman/caveman/webutil"
)

var ErrNotFound = webutil.ErrNotFound

// Store provides access to read PageInfos.
// The idea is that this provides access to data about "browsable"
// pages on the site.  Every path provided here should be accessible
// on the site, but not every accessible path need be presented in
// pageinfo.Store.  The sitemap, search engine, static site generator
// and other features rely on this interface to obtain a list of pages
// on the site.
type Store interface {
	ReadPageInfo(path string) (tmplFileName string, meta map[string]interface{}, err error) // load one
	FindByPath(pathPrefix string, limit int) ([]string, error)                              // get a list of path names
}

// NOTE: Functions can be made for things like i18n and perms that use
// the metadata in a very generic but still consistent way.  Example:
// MetaHasLocale(meta Meta, locale string) bool; or
// MetaHasPerm(meta Meta, perm string) bool
// ("Meta" can be defined in the package as interface{Value(key)interface{}})
// This way the use of a consistent string and value format for these things
// can be enforced but the pageinfo package does not need to know about
// each individual thing that defines a new meta property.

// // StringDataMap describes a map of string keys and generic interface values.
// // This interface matches StringDataMap in the tmpl package, so the meta data
// // is interoperable.  Implementations are not thread-safe.
// type StringDataMap interface {
// 	Value(key string) interface{}
// 	// FIXME: we should probably be using a read-only interface here
// 	Keys() []string
// 	Set(key string, val interface{}) // Set("key", nil) will delete "key"
// }

// PageInfo describes an individual page on the site which can be browsed to.
type PageInfo struct {
	Path         string                 // path in URL
	TmplFileName string                 // filename of template in tmpl Store, empty means rendering is custom
	Meta         map[string]interface{} // metadata for the page
}

// PageInfoListStore implements Store with a static list of PageInfos.
// This is intended for small numbers of pages statically defined.
type PageInfoListStore []PageInfo

func (s PageInfoListStore) ReadPageInfo(path string) (tmplFileName string, meta map[string]interface{}, err error) {
	for _, pi := range s {
		if pi.Path == path {
			return pi.TmplFileName, pi.Meta, nil
		}
	}
	return "", nil, ErrNotFound
}

func (s PageInfoListStore) FindByPath(pathPrefix string, limit int) (ret []string, err error) {
	for _, pi := range s {
		if strings.HasPrefix(pi.Path, pathPrefix) {
			ret = append(ret, pi.Path)
		}
	}
	return ret, nil
}

// StackedStore implements Store by combining other Stores.
type StackedStore []Store
