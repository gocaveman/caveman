package pages

// var ErrNotFound = webutil.ErrNotFound

// exposes "views" and "includes" http.FileSystem from a PageStore
// type PageStoreHFSAdapter struct{}

type PageInfo struct {
	Path string
	Meta StringDataMap
}

// defined as tmpl.PageMeta and pageinfo.PageMeta - tmplfiles can have the yaml impl
// and pageinfo can have the map impl
type StringDataMap interface {
	SetEntry(key string, val interface{}) // SetEntry("key", nil) will delete "key"
	Entry(key string) interface{}
	Keys() []string
}

// TODO: do we need a StringDataMap which is just a simple map - or do we just use
// the YAML one for everything...

// YAMLMap implements StringDataMap with YAML marshaling.
// Key sequence and comments are preserved.
type YAMLMap struct{}

// pageinfo.Store
type PageInfoStore interface {
	FindPageInfo(path string) (*PageInfo, error)
	CreatePageInfo(pi *PageInfo) error
	UpdatePageInfo(pi *PageInfo) error
	DeletePageInfo(path string) error
	// TODO: method to iterate over arbitrarily large set
}

// tmpl.Store
type TemplateStore interface {
	FindTemplate(category, fileName string) (*TemplateItem, error)
	CreateTemplate(vi *TemplateItem) error

	// hm, how does this work for updates - i guess we just do a key merge for
	// the meta - update all the values and delete missing ones... (should be a until func
	// for that)
	UpdateTemplate(vi *TemplateItem) error

	DeleteTemplate(category, fileName string) error
	// TODO: method to iterate over arbitrarily large set, also filter by category
}

type TemplateItem struct {
	// FIXME: Hidden bool ? not sure if that makes sense, consider it
	Category     string
	FileName     string
	Body         []byte
	BodyMimeType string // mime type of input, use "text/gotmpl+html" for Go's html/template, "text/markdown" for markdown
	Meta         StringDataMap
}

// So do we have an adapter that adapts tmplfiles.Store to pageinfo.Store, probably...
// (takes into account underscore paths as hidden)

// type Page struct {

// 	// TODO: should there be an ID indicating which underlying store it came from? maybe not needed...

// 	Category string // "view", "include"
// 	Path     string // (Category,Path) must be unique

// 	Type     string  // FIXME: maybe mime type instead???  indicates the format of Contents
// 	Contents *string // string probably simpler than: DataSource webutil.DataSource
// 	// FIXME: do we need a way to distinguish between empty Contents and Contents not provided?
// 	// maybe this should be a pointer...

// 	Meta map[string]interface{} // meta data for page, FIXME: probably make this it's own type
// }

// type PageStore interface {
// 	FindPage(category, path string) (*Page, error)
// 	FindPageMeta(category, path string) (*Page, error) // returns without loading Contents
// 	DeletePage(category, path string) error
// 	CreatePage(p *Page) error
// 	UpdatePage(p *Page) error
// 	// TODO: method to iterate over arbitrarily large set of pages (probably returning
// 	// without Contents populated - since it's normally used for a list - also filter
// 	// by category, or not), think about stackability
// 	// TODO: also some sort of directory-walk function - although I guess we can defer
// 	// some of this to see what is most useful to the controller
// }

// TODO: figure out packages for this stuff

// FIXME: i18n... ugh...
// ideas:
// - package the text up into the page as templates (could be interesting, kinda clunky though...)
// - page could indicate with a list of locales in the meta what it's avail in - simple
//   convention that the editor is aware of and some template output - that actually seems
//   quite simple, page index code could also be aware (sitemap.xml), canonical tags also
// - what if the editor integrated with the i18n stuff so you click a section of text and
//   say "Translate" and it does the API calls to write whatever i18n strings and drops
//   in the template calls to pull the string... yeah! this is cool...

// StackedPageStore - PageStore that wraps other (multiple) PageStores,
// allows us to create a registry where anyone who wants to can provide
// a PageStore from their package; how does it know which one to send the
// writes to?  Also need a specific error type to avoid

// DBPageStore - PageStore on top of dbr/dbrobj

// FSPageStore - PageStore on top views and includes filesystems

// PageCtrl - REST API on top of a PageStore, used by editor and whatever else

// PageLoadMetaCtrl - uses FindPageMeta to load the metadata for a page being
// requested, before it is renderered (attached to context) - so that
// can be accessed in the template

// TODO: what about static assets - how does that fit into page editing? - probably static
// assets get a similar thing - a store, impl for folder and impl for cdn, stacker
// so both are avail, adapter to fs, ctrl for REST API.  editor just uses REST API.
// Might be smart to have an Asset be aware of it's remote URL, if applicable, so CDN
// resources can be uploaded through the REST API and then appear at whatever URL, in
// some cases on-site and some cases on s3 or whatever.  Inside a page there would be
// some means to get an Asset and then determine it's URL.  Think about localization
// and how this ties in - do we have multiple versions of an image?  blarg...
// Let's at least make the package folder for this one and copy in some notes...
// just need to decode on a name and "assets" is too generic and "static" is a lot
// like "staticgen"...

// ---------------

// FIXME: maybe this is wrong - what about just doing map[string]interface{} ?
// Yeah, why should the page store be responsible for defining how the page
// properties are implemented and what is supported... Makes way more sense to
// just have it a map and let anyone set what they want and then just document
// conventions for things.  There's no way disparate components are going to
// be able to interact if they can't add their own properties and are hindered
// but what the PageStore wants to do.  It's similar to why context.Context is
// arbitrary keys and values.
// type DefaultPageMeta struct {
// 	// data and implement BasicPageMeta,
// 	// can be embedded by things that want to add more fields
// }

// type BasicPageMeta interface {
// 	// meta tags
// 	// whatever other normal stuff
// }
