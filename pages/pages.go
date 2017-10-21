// Index and iteration for list of pages available on a site plus metadata,
// used by various other things that need to get a list of all available pages
// and to provide the various pieces of metadata for each page.
package pages

type Index interface {

	// get page meta - for one path

	// need to be able to iterate over everything in a pageable way, one that supports layering
	// multiple, so we can have an implementation which combines as many as needed
}

// what happens if Meta needs to be replaced (it should have Data but stil...) - it's probably
// enough to generally just treat Meta as an interface{} everywhere else but have a default
// MetaMaker interface impl that returns this one - something like that ought to work fine.
// Put methods on this so interfaces can be used in order to limit knowledge by other components

// fields go on here
type MetaData struct {
	TitleTag string
	// Draft bool
}

// methods go on here
type Meta struct {
	MetaData
}

func (m *Meta) TitleTag() string {
	return m.MetaData.TitleTag
}
