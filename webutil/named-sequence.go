package webutil

import "sort"

type NamedSequenceItem struct {
	Sequence float64
	Name     string
	Value    interface{}
}

// NamedSequence represents a list of items that each have a sequence (by convention it's 0 to 100, but that's not enforced), a name and a value.
// All mutating operations provided here are guaranteed to copy the original without modifying it.  Intended for use by global registries of
// things where you have a list of default things that builds up and it gets customized during application startup.
type NamedSequence []NamedSequenceItem

func (p NamedSequence) Len() int           { return len(p) }
func (p NamedSequence) Less(i, j int) bool { return p[i].Sequence < p[j].Sequence }
func (p NamedSequence) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// TODO: sort by Sequence, add/replace/find/remove by name, convert to []interface{}

func (pl NamedSequence) Copy() NamedSequence {
	ret := make(NamedSequence, len(pl))
	copy(ret, pl)
	return ret
}

// SortedCopy returns a copy of this NamedSequence sorted by Sequence (ascending).
func (pl NamedSequence) SortedCopy() NamedSequence {
	ret := pl.Copy()
	sort.Sort(ret)
	return ret
}

// InterfaceSlice returns the Values as `[]interface{}`
func (pl NamedSequence) InterfaceSlice() []interface{} {
	ret := make([]interface{}, 0, len(pl))
	for _, item := range pl {
		ret = append(ret, item.Value)
	}
	return ret
}

// // NamedIndex returns the index in this of the item with this name.  Returns -1 if not found.
// func (pl NamedSequence) NamedIndex(name string) int {

// }

// func (pl NamedSequence) NamedValue(name string) interface{} {
// }

// func (pl NamedSequence) RemoveByName(name string) (NamedSequence, bool) {

// }

// func (pl NamedSequence) ReplaceValueByName(name string, val interface{}) (NamedSequence, bool) {
// }

// func (pl NamedSequence) Add(seq float64, name string, val interface{}) NamedSequence {
// }

// func (pl NamedSequence) MustAddUnique(seq float64, name string, val interface{}) NamedSequence {
// }
