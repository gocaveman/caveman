// Administrative control panel page and tools.
package adminpanel

// need a way to register things - not sure if that's in here or a "adminregistry" subdir

// hm, gonna package in admin pages right here i guess... that should be fun - in this case
// adminpagesreg should probably be the thing to reference this and include it in the default
// handlers???? needs definite think-through.  probalby need to set up other default registries
// and see how it all compares; be sure to follow the same rule though - registries are just
// defaults but everything is actually wired in main.go
//
// interesting thought: there might only be one admin page - just the listing, and the rest is
// just items that are registered and we link to them

// another thing we probably want is the way to add a message as some sort of admin alert, e.g. for
// important events that occur that we want the user to see when they log into the admin screen,
// hm, does this also extend to email alerts?  should the api support the concept but we don't
// code it now...

// This might make sense to be a sort of folder structure where at each level it can either be
// a folder with sub items (which display as icons), or as a full entry that delegates to whatever
// registered it, but this way it's the same deal as a folder heirarchy.

type EntryItem struct {
	Name          string
	Link          string
	IconHTML      string
	RequiredPerms []string
}

func (i *EntryItem) GetName() string {
	return i.Name
}

func (i *EntryItem) GetLink() string {
	return i.Link
}

func (i *EntryItem) GetIconHTML() string {
	return i.IconHTML
}

func (i *EntryItem) GetRequiredPerms() []string {
	return i.RequiredPerms
}

type Entry interface {
	GetName() string
	GetLink() string
	GetIconHTML() string
	GetRequiredPerms() []string
}

type EntryList []Entry

// Top returns all top level entries (Entries where the Link has no higher prefixes).
func (l EntryList) Top() []Entry {
	panic("not implemented")
	return nil
}

// EntryForLink returns the entry for a specific link.
func (l EntryList) EntryForLink(link string) Entry {
	panic("not implemented")
	return nil
}

// ChildEntriesFor returns the entries which are under the specified link path (but not the Entry at this specific path).
func (l EntryList) ChildEntriesFor(link string) []Entry {
	panic("not implemented")
	return nil
}
