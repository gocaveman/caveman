// Index and iteration for list of pages available on a site plus metadata,
// used by various other things that need to get a list of all available pages
// and to provide the various pieces of metadata for each page.
package pages

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocaveman/caveman/webutil"
)

var ErrNotFound = webutil.ErrNotFound

type Index interface {

	// get page meta - for one path

	// need to be able to iterate over everything in a pageable way, one that supports layering
	// multiple, so we can have an implementation which combines as many as needed

	PageMetaByPath(path string) (PageMeta, error)
	PageListByPrefix(prefix string, limit int, startAfterToken string) (result []string, token string, err error)
}

type CombinedIndex struct {
	IndexMap  map[string]Index
	IndexKeys []string
}

func NewCombinedIndex() *CombinedIndex {
	return &CombinedIndex{
		IndexMap: make(map[string]Index),
	}
}

func (ci *CombinedIndex) AddIndex(tokenPrefix string, index Index) error {
	if tokenPrefix == "" {
		return fmt.Errorf("tokenPrefix cannot be an empty string")
	}
	if ci.IndexMap[tokenPrefix] != nil {
		return fmt.Errorf("tokenPrefix %q is already used", tokenPrefix)
	}
	ci.IndexMap[tokenPrefix] = index
	ci.IndexKeys = append(ci.IndexKeys, tokenPrefix)
	return nil
}

func (ci *CombinedIndex) PageMetaByPath(path string) (PageMeta, error) {

	for _, idx := range ci.Indexes {
		pm, err := idx.PageMetaByPath(path)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		return pm, nil
	}

	return nil, ErrNotFound
}

func (ci *CombinedIndex) PageListByPrefix(prefix string, limit int, startAfterToken string) (result []string, token string, err error) {

	for {

		// parseToken(startAfterToken)

		// if len(result) < limit {
		// }

	}

}

func parseToken(token string) (key, item string) {

	if token == "" {
		return
	}

	parts := strings.SplitN(token, ":", 2)

	if len(parts) < 2 {
		return parts[0], item
	}

	key = url.QueryUnescape(parts[0])
	item = url.QueryUnescape(parts[1])

	return
}

// what happens if Meta needs to be replaced (it should have Data but stil...) - it's probably
// enough to generally just treat Meta as an interface{} everywhere else but have a default
// MetaMaker interface impl that returns this one - something like that ought to work fine.
// Put methods on this so interfaces can be used in order to limit knowledge by other components

// // fields go on here
// type MetaData struct {
// 	TitleTag string
// 	// Draft bool
// }

// // methods go on here
// type Meta struct {
// 	MetaData
// }

// func (m *Meta) TitleTag() string {
// 	return m.MetaData.TitleTag
// }

type PageMeta interface {
	GetTitleTag() string
	GetMetaDesc() string
}
