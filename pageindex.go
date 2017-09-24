package caveman

import (
	"context"
	"fmt"
	"sync"
)

var ErrExceededLimit = fmt.Errorf("exceeded limit")

// PageIndex is a registry of pages for this site.  This is used for
// search indexing, the sitemap or other features which require
// a list of pages on the site.  It also provides a way to get metadata
// associated with these pages.  Note that this is completely
// separate from handlers which serve pages, although it is
// common for a particular feature to provide both an http.Handler
// and a way to register in the PageIndex.  PageIndex is ephemeral,
// it does not persist it's state, the application configures it at
// startup.
// type PageIndex struct {
// 	// FIXME: add context to calls?????  No, this might hurt more than help by promoting that things vary by request that really shouldn't; add it only where needed
// }

// func (pi *PageIndex) FindPages(belowPath string, limit int) ([]string, error) {
// 	return nil, nil
// }

// func (pi *PageIndex) GetPageMeta(path string) (*PageMeta, error) {
// 	return nil, nil
// }

// func (pi *PageIndex) Register(piable PageIndexable) error {
// 	return nil
// }

// type PageIndexable interface {
// 	PathPrefix() string // all paths returns will include this prefix
// 	ListPages(limit int) (paths []string, err error)
// 	GetPage(path string) (*PageMeta, http.Handler, error)
// }

// type PageMeta struct {
// 	// Path  string
// 	// FIXME: all of the default stuff goes here...
// 	Data interface{} // impl-specific struct

// 	// Hm, what about adding a list of CSS and JS file includes here... (plus any options? what about a priority number to force ordering?)
// }

// type DefaultAttrs struct {
// 	// TODO: put in all kinds of stuff here - html title, desc, twitter, fb meta, etc.
// 	// don't forget "NoIndex" and "NoSitemap" - we might need a generic "GetAttr" func which will fetch properties from a strut, a map or a strudt which embeds another strut, etc. so other components can reliably just say "get me the NoIndex" property and not worry about the variations on it
// }

type PageIndex struct {
	rwmu  sync.RWMutex
	pages map[string]*PageMeta
}

// TODO: thread-safe (locked) accessors and mutators

type PageMeta struct {
	HTMLTitle       string `yaml:"html_title"`       // <title> tag
	MetaDescription string `yaml:"meta_description"` // meta description tag
	ShortTitle      string `yaml:"short_title"`      // optional shorter version of title for listing pages, etc.

	// social media meta goes here
	// also some sort of primary image

	Locale string

	// TODO: JS includes, CSS includes (with options? and priority for sequence?)

	Data map[string]interface{}
}

func CtxWithPageMeta(ctx context.Context, pageMeta *PageMeta) context.Context {
	return context.WithValue(ctx, "page-meta", pageMeta)
}
