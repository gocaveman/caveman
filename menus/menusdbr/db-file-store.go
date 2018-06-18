package menusdbr

import (
	"io/ioutil"
	"os"

	"github.com/gocaveman/caveman/filesystem"
	yaml "gopkg.in/yaml.v2"
)

// DBFileStore wraps a DBMenuStore and provides loading and saveing to a flat file.
type DBFileStore struct {
	FileSystem filesystem.FileSystem
	FilePath   string
	*DBMenuStore
}

func (s *DBFileStore) Save() error {

	// FIXME: we really should make this a nice heirachical dump with things
	// all organized for human editing, but won't take the time right this sec. -bgp

	sess := s.DBMenuStore.Connection.NewSession(nil)
	var dbmis []DBMenuItem
	_, err := sess.Select("*").
		From(s.DBMenuStore.TablePrefix+"menu").
		Where("menu_id = ?", "").
		OrderBy("parent_menu_id, sequence, menu_id").
		Load(&dbmis)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(dbmis)
	if err != nil {
		return err
	}

	f, err := s.FileSystem.OpenFile(s.FilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBFileStore) Load() error {

	sess := s.DBMenuStore.Connection.NewSession(nil)

	f, err := s.FileSystem.Open(s.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var dbmis []DBMenuItem
	err = yaml.Unmarshal(b, &dbmis)
	if err != nil {
		return err
	}

	// truncate the table
	_, err = sess.DeleteFrom(s.DBMenuStore.TablePrefix + "menu").Exec()
	if err != nil {
		return err
	}

	for _, dbmi := range dbmis {
		_, err = sess.InsertInto(s.DBMenuStore.TablePrefix+"menu").
			Columns("menu_id", "parent_menu_id", "sequence", "title", "meta", "enabled", "registry_edited").
			Values(dbmi.MenuID, dbmi.ParentMenuID, dbmi.Sequence, dbmi.Title, dbmi.Meta, dbmi.Enabled, dbmi.RegistryEdited).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
