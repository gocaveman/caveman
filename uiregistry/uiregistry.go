package uiregistry

import (
	"errors"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/spf13/afero"
)

var ErrNotFound = errors.New("not found")

// Global is the global registry instance, modules which provide libraries should register them
// here in their init() function.
var Global = NewUIRegistry()

// NewUIRegistry makes a new empty initialized UIRegistry.
func NewUIRegistry() *UIRegistry {
	return &UIRegistry{
		typeNameMap: make(map[string]map[string]EntryPtrList),
		fileNameMap: make(map[string]*Entry),
	}
}

// UIRegistry is a registry of JS and CSS libraries; the global instance of which is called Global in this package.
type UIRegistry struct {
	typeNameMap map[string]map[string]EntryPtrList
	fileNameMap map[string]*Entry
	rwmu        sync.RWMutex
}

// RegReader describes a registry which is capable of resolving requests for libraries.
// Components which need to resolve libraries but not register them (i.e. stuff dealing
// with js and css files during the render path) should use this interface as the
// appropriate abstraction.
type RegReader interface {
	Resolve(typ string, spec string) (Entry, error)
	ResolveFile(fileName string) (Entry, error)
	ResolveAll(typ string, specs []string) ([]Entry, error)
}

// Entry describes a specific version of a library, it's dependencies and provides a way to get it's raw data (DataSource)
type Entry struct {
	Type     string   // type of file, e.g. "js", "css" (also potentially "map")
	Name     string   // name of package, e.g. "jquery"
	Version  string   // semver format, e.g. "1.2.3"
	FileName string   // name of file for use in URLs, e.g. "jquery-1.2.3.js"
	Deps     []string // dependencies in Spec format
	// FIXME: what about adding some sort of "priority" value here so you can specify overrides and say "yes I really want to override this version of 'jquery'" or whatever
	DataSource DataSource // source of file data
}

type EntryPtrList []*Entry

func (p EntryPtrList) Len() int      { return len(p) }
func (p EntryPtrList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p EntryPtrList) Less(i, j int) bool {
	vi, _ := semver.Parse(p[i].Version)
	vj, _ := semver.Parse(p[j].Version)
	return vi.Compare(vj) < 0
}

type DataSource interface {
	// OpenData opens a readable stream of data for the file.
	OpenData() (ReadSeekCloser, error)
}

type ReadSeekCloser interface {
	io.Closer
	io.Reader
	io.Seeker
}

func NewFileDataSource(fs afero.Fs, path string) (DataSource, error) {
	// fs.Open(name)
}

func NewBytesDataSource(b []byte) DataSource {

}

// Register is called to add an entry to the registry (a specific single version of a library)
func (r *UIRegistry) Register(entry Entry) error {

	// make sure version looks valid
	_, err := semver.Parse(entry.Version)
	if err != nil {
		return err
	}

	// write lock
	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	nMap := r.typeNameMap[entry.Type]
	if nMap == nil {
		nMap = make(map[string]EntryPtrList)
	}

	entries := nMap[entry.Name]

	// FIXME: what about duplicates???
	entries = append(entries, &entry)

	sort.Sort(entries)

	nMap[entry.Name] = entries

	r.typeNameMap[entry.Type] = nMap

	// FIXME: check for duplicate here
	r.fileNameMap[entry.FileName] = &entry

	return nil
}

func (r *UIRegistry) Resolve(typ string, spec string) (ret Entry, retErr error) {
	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	nMap := r.typeNameMap[typ]
	if nMap == nil {
		return ret, ErrNotFound
	}

	sSpec := Spec(spec)
	name := sSpec.Name()
	sRange, err := sSpec.Range()
	if err != nil {
		return ret, err
	}

	entries := nMap[name]
	for _, entry := range entries {
		ver, _ := semver.Parse(entry.Version)
		if sRange(ver) {
			return *entry, nil
		}
	}

	return ret, nil
}

func (r *UIRegistry) ResolveFile(fileName string) (ret Entry, retErr error) {
	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	entry := r.fileNameMap[fileName]
	if entry != nil {
		return *entry, nil
	}

	return ret, ErrNotFound
}

func (r *UIRegistry) ResolveAll(typ string, specs []string) ([]Entry, error) {
	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	panic("TO BE IMPLEMENTED")

	return nil, nil
}

// Spec format "package@semverrange"
type Spec string

func (s Spec) Name() string {
	return strings.Split(string(s), "@")[0]
}

func (s Spec) Range() (semver.Range, error) {
	parts := strings.Split(string(s), "@")
	if len(parts) < 2 {
		return semver.ParseRange("*")
	}

	return semver.ParseRange(parts[1])
}
