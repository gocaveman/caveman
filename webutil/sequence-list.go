package webutil

import "sort"

type SequenceListItem struct {
	Sequence float64
	Name     string
	Value    interface{}
}

// SequenceList represents a list of items that each have a sequence (by convention it's 0 to 100, but that's not enforced), a name and a value.
// All mutating operations provided here are guaranteed to copy the original without modifying it.  Intended for use by global registries of
// things where you have a list of default things that builds up and it gets customized during application startup.
type SequenceList []SequenceListItem

func (p SequenceList) Len() int           { return len(p) }
func (p SequenceList) Less(i, j int) bool { return p[i].Sequence < p[j].Sequence }
func (p SequenceList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// TODO: sort by Sequence, add/replace/find/remove by name, convert to []interface{}

func (pl SequenceList) Copy() SequenceList {
	ret := make(SequenceList, len(pl))
	copy(ret, pl)
	return ret
}

// SortedCopy returns a copy of this SequenceList sorted by Sequence (ascending).
func (pl SequenceList) SortedCopy() SequenceList {
	ret := pl.Copy()
	sort.Sort(ret)
	return ret
}

// InterfaceSlice returns the Values as `[]interface{}`
func (pl SequenceList) InterfaceSlice() []interface{} {
	ret := make([]interface{}, 0, len(pl))
	for _, item := range pl {
		ret = append(ret, item.Value)
	}
	return ret
}

// // NamedIndex returns the index in this of the item with this name.  Returns -1 if not found.
// func (pl SequenceList) NamedIndex(name string) int {

// }

// func (pl SequenceList) NamedValue(name string) interface{} {
// }

// func (pl SequenceList) RemoveByName(name string) (SequenceList, bool) {

// }

// func (pl SequenceList) ReplaceValueByName(name string, val interface{}) (SequenceList, bool) {
// }

// func (pl SequenceList) Add(seq float64, name string, val interface{}) SequenceList {
// }

// func (pl SequenceList) MustAddUnique(seq float64, name string, val interface{}) SequenceList {
// }
