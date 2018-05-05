package tmplregistry

import (
	"github.com/gocaveman/caveman/tmpl"
	"github.com/gocaveman/caveman/webutil"
)

const (
	SeqTheme float64 = 80.0
)

// OnlyReadableFromMain determines if `Contents()` can be called from packages other than main.
var OnlyReadableFromMain = true

var reg webutil.NamedSequence

// MustRegister adds a new tmpl.Store to the registry.
func MustRegister(seq float64, name string, store tmpl.Store) {
	reg = append(reg, webutil.NamedSequenceItem{Sequence: seq, Name: name, Value: store})
}

// Contents returns the current contents of the registry as a NamedSequence.
func Contents() webutil.NamedSequence {
	if OnlyReadableFromMain {
		webutil.MainOnly(1)
	}
	return reg.SortedCopy()
}

// ContentsStore returns the registry contents as a tmpl.StackedStore.
func ContentsStore() tmpl.StackedStore {
	if OnlyReadableFromMain {
		webutil.MainOnly(1)
	}
	regc := reg.SortedCopy()
	var ret tmpl.StackedStore
	for _, item := range regc {
		ret = append(ret, item.Value.(tmpl.Store))
	}
	return ret
}
