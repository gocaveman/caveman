// FIXME: crap, this should really be "uifilesregistry" to follow the convention...
// A registry of CSS, JS files and their dependency chain.
package uiregistry

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gocaveman/caveman/webutil"
)

var ErrNotFound = errors.New("not found")

// Global is the global registry instance, modules which provide libraries should register them
// here in their init() function.
// FIXME: Instead of a variable, I think we need two function calls - one to get something that implements
// an interface for registering, and another that returns an interface that can read.  In the read call we
// can verify it is called from main - webutil.MainOnly()
var global = NewUIRegistry()

func Contents() *UIRegistry {
	return global
}

func ParseName(s string) (typ, name string, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected exactly two colon separated parts but instead found %d part(s)", len(parts))
	}
	return parts[0], parts[1], nil
}

// FIXME: what if we use http.FileSystem and a path, instead of using webutil.DataSource?
// It would be one less funky custom type and has effectively the same semantics.

func MustRegister(name string, deps []string, ds webutil.DataSource) {
	global.MustRegister(name, deps, ds)
}

func Register(name string, deps []string, ds webutil.DataSource) error {
	return global.Register(name, deps, ds)
}

// NewUIRegistry makes a new empty initialized UIRegistry.
func NewUIRegistry() *UIRegistry {
	return &UIRegistry{
		reg: make(map[string]*entry),
	}
}

// UIRegistry is a registry of JS and CSS libraries; the global instance of which is called Global in this package.
type UIRegistry struct {
	rwmu sync.RWMutex

	reg map[string]*entry
}

func (r *UIRegistry) MustRegister(name string, deps []string, ds webutil.DataSource) {
	err := r.Register(name, deps, ds)
	if err != nil {
		panic(err)
	}
}
func (r *UIRegistry) Register(name string, deps []string, ds webutil.DataSource) error {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	_, _, err := ParseName(name)
	if err != nil {
		return err
	}

	r.reg[name] = &entry{
		DataSource: ds,
		Deps:       deps,
	}

	return nil
}

// Lookup takes a single name and returns the corresponding DataSource.
func (r *UIRegistry) Lookup(name string) (webutil.DataSource, error) {

	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	if ret, ok := r.reg[name]; ok {
		return ret.DataSource, nil
	}

	return nil, ErrNotFound
}

// ResolveDeps takes one or more names and returns a list of the libraries that need to be
// included in order to resolve all dependencies, in the correct sequence.
func (r *UIRegistry) ResolveDeps(names ...string) ([]string, error) {

	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	var out []string

	// go through each requested library
	for _, name := range names {

		entry, ok := r.reg[name]
		if !ok {
			return nil, fmt.Errorf("unable to find registry entry for %q", name)
		}

		// for each dependency add it before this entry, unless it's already there
		for _, dep := range entry.Deps {
			if !stringSliceContains(out, dep) {
				out = append(out, dep)
			}
		}

		// add this entry
		out = append(out, name)

	}

	// if after including dependencies the list was changed, recurse and resolve dependencies again
	if !isSameStrings(out, names) {
		return r.ResolveDeps(out...)
	}

	// dependencies didn't change the list, we're done
	return out, nil

}

// Resolver describes a registry which is capable of resolving requests for libraries.
// Components which need to resolve libraries but not register them (i.e. stuff dealing
// with js and css files during the render path) should use this interface as the
// appropriate abstraction.
type Resolver interface {
	Lookup(name string) (webutil.DataSource, error)
	ResolveDeps(name ...string) ([]string, error)
}

// Entry describes a specific version of a library, it's dependencies and provides a way to get it's raw data (DataSource)
type entry struct {
	DataSource webutil.DataSource
	Deps       []string
}

func isSameStrings(s1 []string, s2 []string) bool {

	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

func stringSliceContains(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}
