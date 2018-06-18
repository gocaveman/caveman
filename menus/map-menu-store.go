package menus

import (
	"sort"
	"sync"
)

func NewMapMenuStore() *MapMenuStore {
	return &MapMenuStore{
		menuItems: make(map[string]*MenuItem),
	}
}

// MapMenuStore implements Store using an in-memory map.
// It is safe for confurrent use.
type MapMenuStore struct {
	menuItems map[string]*MenuItem
	mu        sync.RWMutex
}

func (s *MapMenuStore) ReadMenuItem(id string) (*MenuItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mi, ok := s.menuItems[id]
	if !ok {
		return nil, ErrNotFound
	}
	miCopy := *mi
	return &miCopy, nil
}

func (s *MapMenuStore) CreateMenuItem(mi *MenuItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.menuItems[mi.MenuID]
	if ok {
		return ErrAlreadyExists
	}
	miCopy := *mi
	s.menuItems[miCopy.MenuID] = &miCopy
	return nil
}

func (s *MapMenuStore) UpdateMenuItem(mi *MenuItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.menuItems[mi.MenuID]
	if !ok {
		return ErrNotFound
	}
	miCopy := *mi
	s.menuItems[miCopy.MenuID] = &miCopy
	return nil
}

func (s *MapMenuStore) DeleteMenuItem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.menuItems, id)
	return nil
}

func (s *MapMenuStore) FindChildren(id string) ([]string, error) {
	// FIXME: We really should index this properly, menus are
	// read-intensive and this call will be very frequent.
	// Even if we had to rebuild the entire index on each change, it
	// would probabyl still be worth it.
	s.mu.RLock()
	defer s.mu.RUnlock()
	var retMenuItems MenuItemList
	for _, mi := range s.menuItems {
		if mi.ParentMenuID == id {
			retMenuItems = append(retMenuItems, *mi)
		}
	}
	sort.Sort(retMenuItems)
	ret := make([]string, 0, len(retMenuItems))
	for _, mi := range retMenuItems {
		ret = append(ret, mi.MenuID)
	}
	return ret, nil
}
