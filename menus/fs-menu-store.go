package menus

// // NewFSMenuStore returns a new FSMenuStore.
// func NewFSMenuStore(fileSystem filesystem.FileSystem, filePath string, registryItems MenuItemList) *FSMenuStore {
// 	ret := &FSMenuStore{
// 		FileSystem:    fileSystem,
// 		FilePath:      filePath,
// 		registryItems: MenuItemList,
// 	}
// 	return ret
// }

// // FSMenuStore implements Store by loading and saving all menu data to a YAML file.
// // It also can use a registry to provide default menu items that can be edited
// // in order to customize, without those edits being overwritten.
// type FSMenuStore struct {
// 	FileSystem filesystem.FileSystem // filesystem to store to
// 	FilePath   string                // path of file to write to (yaml)

// 	internalStore *MapMenuStore // our in memory represtation of the menus

// 	registryItems     MenuItemList    // the registry
// 	registryEditedIDs map[string]bool // the registry items that have been edited

// 	fileMu   sync.RWMutex // lock around file operations
// 	lastTime time.Time    // last timestamp related to the file, to detect if file changed on disk
// }

// // // fsMenuStoreFile is what gets yaml marshal/unmarshal'ed from the file
// // type fsMenuStoreFile struct {
// // 	DisplayMenuItem Root `json:"root" yaml:"root"`
// // 	// RegistryEditedIDs []string `json:"registry_edited_ids" yaml:"registry_edited_ids"`
// // }

// func (s *FSMenuStore) Close() error {
// 	return s.SaveFile()
// }

// func (s *FSMenuStore) SaveFile() error {
// 	s.fileMu.Lock()
// 	defer s.fileMu.Unlock()
// }

// func (s *FSMenuStore) LoadFile() error {
// 	s.fileMu.Lock()
// 	defer s.fileMu.Unlock()

// 	f, err := s.FileSystem.Open(s.FilePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()
// 	fb, err := ioutil.ReadAll(f)
// 	if err != nil {
// 		return err
// 	}

// 	var dmi DisplayMenuItem

// 	err = yaml.Unmarshal(fb, &dmi)
// 	if err != nil {
// 		return err
// 	}

// 	m := NewMapMenuStore()

// 	var handleItem func(i DisplayMenuItem) error
// 	handleItem = func(i DisplayMenuItem, parentID string) error {

// 		// overwrite the parent menu ID
// 		i.MenuItem.ParentMenuID = parentID
// 		err := m.CreateMenuItem(&i.MenuItem)
// 		if err != nil {
// 			return err
// 		}

// 		// TODO: handle the additional for registry overrides

// 		for _, child := range i.Children {
// 			err = handleItem(child, i.MenuID)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	}

// 	err = handleItem(dmi, "")
// 	if err != nil {
// 		return err
// 	}

// 	s.internalStore = m

// 	return nil
// }

// func (s *FSMenuStore) LoadFileIfChanged() error {

// 	st, err := s.FileSystem.Stat(s.FilePath)
// 	if err != nil {
// 		return err
// 	}
// 	// file is older than lastTime, we're done here
// 	if !st.ModTime().After(s.lastTime) {
// 		return nil
// 	}
// 	return s.LoadFile()
// }

// func (s *FSMenuStore) ReadMenuItem(id string) (*MenuItem, error) {
// 	err := s.LoadFileIfChanged()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s.internalStore.ReadMenuItem(id)
// }

// func (s *FSMenuStore) CreateMenuItem(mi *MenuItem) error {
// 	err := s.internalStore.CreateMenuItem(mi)
// 	if err != nil {
// 		return err
// 	}
// 	return s.SaveFile()
// }

// func (s *FSMenuStore) UpdateMenuItem(mi *MenuItem) error {
// 	err := s.internalStore.UpdateMenuItem(mi)
// 	if err != nil {
// 		return err
// 	}
// 	return s.SaveFile()
// }

// func (s *FSMenuStore) DeleteMenuItem(id string) error {
// 	err := s.internalStore.DeleteMenuItem(id)
// 	if err != nil {
// 		return err
// 	}
// 	return s.SaveFile()
// }

// func (s *FSMenuStore) FindChildren(id string) ([]string, error) {
// 	err := s.LoadFileIfChanged()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s.internalStore.FindChildren(id)
// }

// // func escapeFileName(s string) string {
// // 	var out []byte
// // 	sb := []byte(s)
// // 	for _, b := range sb {
// // 		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
// // 			out = append(out, b)
// // 			continue
// // 		}
// // 		out = append(out, []byte(fmt.Sprintf("_%X", b))...)
// // 	}
// // 	return string(out)
// // }

// // func unescapeFileName(s string) string {
// // 	var out []byte
// // 	sb := []byte(s)
// // 	for i := 0; i < len(sb); i++ {
// // 		b := sb[i]
// // 		if b != '_' {
// // 			out = append(out, b)
// // 			continue
// // 		}
// // 		if i+2 < len(sb) {
// // 			sbsubstr := string(sb[i+1 : i+3])
// // 			o, _ := hex.DecodeString(sbsubstr)
// // 			out = append(out, o...)
// // 		}
// // 		i += 2
// // 	}
// // 	return string(out)
// // }
