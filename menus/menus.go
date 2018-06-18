package menus

import (
	"fmt"

	"github.com/gocaveman/caveman/webutil"
)

var ErrNotFound = webutil.ErrNotFound
var ErrAlreadyExists = fmt.Errorf("already exists")

type Store interface {
	ReadMenuItem(id string) (*MenuItem, error)
	CreateMenuItem(mi *MenuItem) error
	UpdateMenuItem(mi *MenuItem) error
	DeleteMenuItem(id string) error
	// FIXME: Do we need a recursive delete?  If not, do we allow
	// deletes for ones that have children?  We could just block that
	// action and provide a separate naked function to do a recurive delete.
	FindChildren(id string) ([]string, error)
}

type MenuItemReader interface {
	ReadMenuItem(id string) (*MenuItem, error)
	FindChildren(id string) ([]string, error)
}

type MenuItem struct {
	MenuID       string                      `json:"menu_id" yaml:"menu_id" db:"menu_id"`
	ParentMenuID string                      `json:"parent_menu_id" yaml:"parent_menu_id" db:"parent_menu_id"`
	Sequence     float64                     `json:"sequence" yaml:"sequence" db:"sequence"`
	Title        string                      `json:"title" yaml:"title" db:"title"`
	Meta         webutil.SimpleStringDataMap `json:"meta" yaml:"meta" db:"meta"`
	Enabled      bool                        `json:"enabled" yaml:"enabled" db:"enabled"`
}

type MenuItemList []MenuItem

// TODO: We could probably add some builder pattern stuff here so you can do things
// like WithChildren() or other things that would make statically providing some menus
// in a package to the registry easier to do.

func (a MenuItemList) Len() int      { return len(a) }
func (a MenuItemList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MenuItemList) Less(i, j int) bool {
	if a[i].Sequence == a[j].Sequence {
		return a[i].MenuID < a[j].MenuID
	}
	return a[i].Sequence < a[j].Sequence
}

// TODO: MenuScopeHandler to provide MenuScope in context,
// and then register it in the handlerregistry and autowire.

// MenuScope provides display facilities for use in templates.
type MenuScope struct {
	MenuItemReader `autowire:""`
}

func (ms *MenuScope) Display(id string) (DisplayMenuItem, error) {
	return BuildDisplayMenuItem(ms.MenuItemReader, id)
}

type DisplayMenuItem struct {
	MenuItem
	Children   []DisplayMenuItem      `json:"children,omitempty" yaml:"children,omitempty"`
	Additional map[string]interface{} `json:"additional,omitempty" yaml:"additional,omitempty"`
}

func BuildDisplayMenuItem(r MenuItemReader, id string) (ret DisplayMenuItem, err error) {

	var mi *MenuItem
	mi, err = r.ReadMenuItem(id)
	if err != nil {
		return
	}
	ret.MenuItem = *mi

	var childIDs []string
	childIDs, err = r.FindChildren(id)
	if err != nil {
		return
	}

	for _, childID := range childIDs {
		var child DisplayMenuItem
		child, err = BuildDisplayMenuItem(r, childID)
		if err != nil {
			return
		}
		ret.Children = append(ret.Children, child)
	}

	return
}
