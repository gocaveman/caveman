package tmpl

// type FileSystemStore struct {
// 	FileExtMimeTypes map[string]string
// 	FileSystems      map[string]filesystem.FileSystem // key is category, value FileSystem to expose
// }

// func (s *FileSystemStore) CreateTemplate(t *Tmpl) error {
// 	return ErrWriteNotSupported
// }

// func (s *FileSystemStore) UpdateTemplate(t *Tmpl) error {
// 	return ErrWriteNotSupported
// }

// func (s *FileSystemStore) DeleteTemplate(category, fileName string) error {
// 	return ErrWriteNotSupported
// }

// func (s *FileSystemStore) Categories() ([]string, error) {
// 	var ret []string
// 	for k := range s.FileSystems {
// 		ret = append(ret, k)
// 	}
// 	return ret, nil
// }

// func (s *FileSystemStore) ReadTemplate(category, fileName string) (*Tmpl, error) {

// 	fs := s.FileSystems[category]
// 	if fs == nil {
// 		return nil, ErrNotFound
// 	}

// 	f, err := fs.Open(fileName)
// 	if err != nil {
// 		if os.IsNotExist(err) || err == ErrNotFound {
// 			return nil, ErrNotFound
// 		}
// 		return nil, err
// 	}
// 	defer f.Close()
// 	fi, err := f.Stat()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if fi.IsDir() {
// 		return nil, ErrNotFound // directories are invisible as far as ReadTemplate is concerned
// 	}

// 	meta, body, err := ParseYAMLHeadTemplate(f)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ret := &Tmpl{
// 		Category: category,
// 		FileName: fileName,
// 		Body:     body,
// 		Meta:     meta,
// 	}

// 	ret.BodyMimeType = s.FileExtMimeTypes[path.Ext(fileName)]
// 	if ret.BodyMimeType == "" {
// 		ret.BodyMimeType = DefaultFileExtMimeTypes[path.Ext(fileName)]
// 	}

// 	return ret, nil
// }

// // FindByPrefix will return a slice of file names for the specified category
// // that begin with a prefix, up to an indicated limit.  Limit -1 means all.
// // Prefix "/" will list all.  Prefix must be a directory name.
// func (s *FileSystemStore) FindByPrefix(category, fileNamePrefix string, limit int) ([]string, error) {

// 	fs := s.FileSystems[category]
// 	if fs == nil {
// 		return nil, ErrNotFound
// 	}

// 	var ret []string

// 	var done = errors.New("done")

// 	err := vfsutil.WalkFiles(fs, fileNamePrefix, func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if fi.IsDir() {
// 			return nil
// 		}
// 		ret = append(ret, path)
// 		if limit > 0 && len(ret) >= limit {
// 			return done
// 		}
// 		return nil
// 	})
// 	if err != nil && err != done {
// 		return nil, err
// 	}

// 	if limit > 0 && len(ret) > limit {
// 		ret = ret[:limit]
// 	}

// 	return ret, nil
// }
