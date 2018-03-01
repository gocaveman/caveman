package i18n

import (
	"log"
	"strings"

	"github.com/gocaveman/caveman/webutil"
)

// Translator translates text without knowledge of the target locale, it is provided in the call.
type Translator interface {
	// Translate a string into the first locale found in the list.  Note that the group
	// and key are case sensitive but the locales are not.
	// ErrNotFound must be returned if the key cannot be found in any of the locales given.
	Translate(g, k string, locales ...string) (string, error)
}

// MapKey is the key for MapTranslator
type MapKey struct {
	Group  string
	Key    string
	Locale string
}

// MapTranslator is a simple in-memory map Translator implementation.
type MapTranslator struct {
	Values map[MapKey]string
}

// NewMapTranslator returns a new empty MapTranslator.
func NewMapTranslator() *MapTranslator {
	return &MapTranslator{Values: make(map[MapKey]string)}
}

func MergeMapTranslators(mts ...*MapTranslator) *MapTranslator {
	ret := NewMapTranslator()
	for _, mt := range mts {
		for k, v := range mt.Values {
			ret.Values[k] = v
		}
	}
	return ret
}

// SetEntry assigns a value by its group, key and locale.
func (mt *MapTranslator) SetEntry(g, k, l, v string) {
	mt.Values[MapKey{Group: g, Key: k, Locale: strings.ToLower(l)}] = v
}

// GetEntry retrieves a value by its group, key and locale. Empty string if not there.
func (mt *MapTranslator) GetEntry(g, k, l string) string {
	return mt.Values[MapKey{Group: g, Key: k, Locale: strings.ToLower(l)}]
}

// CheckEntry retrieves a value by its group, key and locale. Empty string if not there
// but also returns a bool as true if there or false if not.
func (mt *MapTranslator) CheckEntry(g, k, l string) (string, bool) {
	ret, ok := mt.Values[MapKey{Group: g, Key: k, Locale: strings.ToLower(l)}]
	return ret, ok
}

// Translate implements the Translator interface.
func (mt *MapTranslator) Translate(g, k string, locales ...string) (string, error) {

	for _, l := range locales {
		ret, ok := mt.CheckEntry(g, k, l)
		if ok {
			return ret, nil
		}
	}

	return k, ErrNotFound
}

// NamedSequenceTranslator implements Translator using a NamedSequence of other translators
// to delegate to in sequence.  If Debug is true it will produce detailed log output for every
// call to Translate and which item succeeded.
type NamedSequenceTranslator struct {
	NamedSequence webutil.NamedSequence
	Debug         bool
}

func NewNamedSequenceTranslator(ns webutil.NamedSequence, debug bool) *NamedSequenceTranslator {
	ns = ns.SortedCopy()
	return &NamedSequenceTranslator{NamedSequence: ns, Debug: debug}
}

func (t *NamedSequenceTranslator) Translate(g, k string, locales ...string) (string, error) {

	for _, nsi := range t.NamedSequence {
		// TODO: look at optimizing this - if it's called a lot we might be able to squeeze some perf here
		itemt := nsi.Value.(Translator)
		ret, err := itemt.Translate(g, k, locales...)
		if err == ErrNotFound {
			continue
		}

		if t.Debug {
			log.Printf("Translate(g=%q, k=%q, l=%q) returning (err=%v) for (sequence=%v, name=%q): %q", g, k, locales, err, nsi.Sequence, nsi.Name, ret)
		}

		return ret, err
	}

	if t.Debug {
		log.Printf("Translate(g=%q, k=%q, l=%q) not found in named sequence", g, k, locales)
	}

	return k, ErrNotFound
}
