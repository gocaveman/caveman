package tmpl

// StackedStore implements Store by layering a list of underlying Stores.
// Write operations will be attempted on these Stores in sequence until one returns
// anything other than ErrWriteNotSupported.  The common case is to have your
// writable Store as the first in the slice and other read-only Stores can be added
// after in order to make additional templates available.  This tends to provide
// the most useful behavior.
type StackedStore []Store

func (ss StackedStore) CreateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error {
	for _, s := range ss {
		err := s.CreateTemplate(category, fileName, body, mimeType, meta)
		if err != ErrWriteNotSupported {
			return err
		}
	}
	return ErrWriteNotSupported
}

func (ss StackedStore) UpdateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error {
	for _, s := range ss {
		err := s.UpdateTemplate(category, fileName, body, mimeType, meta)
		if err != ErrWriteNotSupported {
			return err
		}
	}
	return ErrWriteNotSupported
}

func (ss StackedStore) DeleteTemplate(category, fileName string) error {
	for _, s := range ss {
		err := s.DeleteTemplate(category, fileName)
		if err != ErrWriteNotSupported {
			return err
		}
	}
	return ErrWriteNotSupported
}

func (ss StackedStore) Categories() ([]string, error) {
	foundMap := make(map[string]bool, 2)
	var ret []string
	for _, s := range ss {
		cats, err := s.Categories()
		if err != nil {
			return ret, err
		}
		for _, cat := range cats {
			if !foundMap[cat] {
				ret = append(ret, cat)
				foundMap[cat] = true
			}
		}
	}
	return ret, nil
}

func (ss StackedStore) ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error) {
	for _, s := range ss {
		body, mimeType, meta, err = s.ReadTemplate(category, fileName)
		if err != ErrNotFound {
			return
		}
	}
	return nil, "", nil, ErrNotFound
}

func (ss StackedStore) FindByPrefix(category, fileNamePrefix string, limit int) ([]string, error) {
	foundMap := make(map[string]bool)
	var ret []string
	for _, s := range ss {
		names, err := s.FindByPrefix(category, fileNamePrefix, limit)
		if err != nil {
			return ret, err
		}
		for _, name := range names {
			if !foundMap[name] {
				ret = append(ret, name)
				foundMap[name] = true
			}
		}

		if limit > 0 && len(ret) > limit {
			break
		}

	}

	if limit > 0 && len(ret) > limit {
		ret = ret[:limit]
	}

	return ret, nil
}
