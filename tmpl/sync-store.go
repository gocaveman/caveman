package tmpl

import "sync"

// NewSyncStore returns a SyncStore with it's underlying Store initialized.
func NewSyncStore(s Store) *SyncStore {
	return &SyncStore{
		Store: s,
	}
}

// SyncStore wraps calls to a store around a sync.RWMutex, allowing you to safely change the underlying implementation at runtime.
type SyncStore struct {
	sync.RWMutex
	Store Store
}

// SetStore safely assigns the value of Store within Lock/Unlock calls.
func (s *SyncStore) SetStore(store Store) {
	s.Lock()
	defer s.Unlock()
	s.Store = store
}

// CreateTemplate acquires an RLock() and then delegates to Store.
func (s *SyncStore) CreateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error {
	s.RLock()
	defer s.RUnlock()
	return s.Store.CreateTemplate(category, fileName, body, mimeType, meta)
}

// ReadTemplate acquires an RLock() and then delegates to Store.
func (s *SyncStore) ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.Store.ReadTemplate(category, fileName)
}

// UpdateTemplate acquires an RLock() and then delegates to Store.
func (s *SyncStore) UpdateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error {
	s.RLock()
	defer s.RUnlock()
	return s.Store.UpdateTemplate(category, fileName, body, mimeType, meta)
}

// DeleteTemplate acquires an RLock() and then delegates to Store.
func (s *SyncStore) DeleteTemplate(category, fileName string) error {
	s.RLock()
	defer s.RUnlock()
	return s.Store.DeleteTemplate(category, fileName)
}

// Categories acquires an RLock() and then delegates to Store.
func (s *SyncStore) Categories() ([]string, error) {
	s.RLock()
	defer s.RUnlock()
	return s.Store.Categories()
}

// FindByPrefix acquires an RLock() and then delegates to Store.
func (s *SyncStore) FindByPrefix(category, fileNamePrefix string, limit int) ([]string, error) {
	s.RLock()
	defer s.RUnlock()
	return s.Store.FindByPrefix(category, fileNamePrefix, limit)
}
