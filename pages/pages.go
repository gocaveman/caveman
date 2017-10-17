package pages

type Index interface {

	// get page meta - for one path

	// need to be able to iterate over everything in a pageable way, one that supports layering
	// multiple, so we can have an implementation which combines as many as needed
}

// what happens if Meta needs to be replaced (it should have Data but stil...)
type Meta struct {
}
