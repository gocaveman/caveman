// Database and flat file persistence for regions.
package regionsdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/gocaveman/caveman/autowire"
	"github.com/gocaveman/caveman/filesystem"
	"github.com/gocaveman/caveman/migrate"
	"github.com/gocaveman/caveman/migrate/migrateregistry"
	"github.com/gocaveman/caveman/regions"
	"github.com/gocraft/dbr"
	yaml "gopkg.in/yaml.v2"
)

// DefaultRegionsMigrations is all of our migrations for this store.
var DefaultRegionsMigrations migrate.MigrationList

func init() {

	// register in migrateregistry and with autowire for all 3 databases
	reg := func(m *migrate.SQLTmplMigration) {
		// TODO: see if we can compact these back down to one-liners
		var rm migrate.Migration
		rm = m.NewWithDriverName("sqlite3")
		DefaultRegionsMigrations = append(DefaultRegionsMigrations, rm)
		autowire.Populate(migrateregistry.MustRegister(rm))
		rm = m.NewWithDriverName("mysql")
		DefaultRegionsMigrations = append(DefaultRegionsMigrations, rm)
		autowire.Populate(migrateregistry.MustRegister(rm))
		rm = m.NewWithDriverName("postgres")
		DefaultRegionsMigrations = append(DefaultRegionsMigrations, rm)
		autowire.Populate(migrateregistry.MustRegister(rm))
	}

	reg(&migrate.SQLTmplMigration{
		// DriverNameValue set by reg
		CategoryValue: "regions",
		VersionValue:  "0001_region_definition_create", // must be unique and indicates sequence
		UpSQL: []string{
			`CREATE TABLE {{.TablePrefix}}region_definition (
				definition_id VARCHAR(255),
				region_name VARCHAR(255),
				sequence DOUBLE,
				disabled INTEGER,
				template_name VARCHAR(255),
				cond_include_paths TEXT,
				cond_exclude_paths TEXT,
				cond_template TEXT,
				context_meta TEXT,
				PRIMARY KEY (definition_id)
			)`,
		},
		DownSQL: []string{
			`DROP TABLE {{.TablePrefix}}region_definition`,
		},
	})

}

// DBStore implements Store against a database table and optionally provides persistence to a flat file.
type DBStore struct {
	DBDriver    string `autowire:"db.DriverName"`
	DBDSN       string `autowire:"db.DataSourceName"`
	TablePrefix string `autowire:"db.TablePrefix,optional"`

	// if FileSystem is non-nil then flat-file persistence is enabled in addition to database functionality
	FileSystem     filesystem.FileSystem `autowire:"db.FlatFileSystem,optional"`
	FilePathPrefix string                `autowire:"db.FlatFilePrefix,optional"`
	FilePathSuffix string                // defaults to "regions.yaml"

	conn          *dbr.Connection
	needsFileLoad bool         // set this to trigger a reload before the next operation, used at startup
	rwmu          sync.RWMutex // for file operations
}

func (s *DBStore) AfterWire() error {

	if s.FilePathSuffix == "" {
		s.FilePathSuffix = "regions.yaml"
	}

	var err error
	// FIXME: figure out how we can fill the logger here from autowire - allowing detailed database
	// logging when needed
	s.conn, err = dbr.Open(s.DBDriver, s.DBDSN, nil)
	if err != nil {
		return err
	}

	if s.FileSystem != nil {
		pdir := path.Dir(path.Join(s.FilePathPrefix, s.FilePathSuffix))
		err := s.FileSystem.MkdirAll(pdir, 0755)
		if err != nil {
			return err
		}
		s.needsFileLoad = true
	}

	return nil
}

func (s *DBStore) WriteDefinition(d regions.Definition) error {

	if err := d.IsValid(); err != nil {
		return err
	}

	err := s.condLoadFromFlatFile()
	if err != nil {
		return err
	}

	// FIXME: figure out db logging
	sess := s.conn.NewSession(nil)

	tx, err := sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	var foundDefID string
	err = tx.Select("definition_id").
		From(s.TablePrefix+"region_definition").
		Where("definition_id=?", d.DefinitionID).LoadOne(&foundDefID)
	if err == dbr.ErrNotFound {
		_, err := tx.InsertInto(s.TablePrefix+"region_definition").
			Columns(
				"definition_id",
				"region_name", "sequence", "disabled", "template_name",
				"cond_include_paths", "cond_exclude_paths", "cond_template",
				"context_meta").
			Record(&d).Exec()

		if err != nil {
			return err
		}
		foundDefID = d.DefinitionID

	} else if err != nil {
		return err
	}

	if foundDefID != d.DefinitionID {
		return fmt.Errorf("selected wrong definition ID from database, should be impossible - is column case insensitive?")
	}

	_, err = tx.Update(s.TablePrefix+"region_definition").
		Set("region_name", d.RegionName).
		Set("sequence", d.Sequence).
		Set("disabled", d.Disabled).
		Set("template_name", d.TemplateName).
		Set("cond_include_paths", d.CondIncludePaths).
		Set("cond_exclude_paths", d.CondExcludePaths).
		Set("cond_template", d.CondTemplate).
		Set("context_meta", d.ContextMeta).
		Exec()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if s.FileSystem != nil {
		return s.SaveToFlatFile()
	}
	return nil
}

func (s *DBStore) DeleteDefinition(defintionID string) error {

	err := s.condLoadFromFlatFile()
	if err != nil {
		return err
	}

	sess := s.conn.NewSession(nil)
	_, err = sess.DeleteFrom(s.TablePrefix+"region_definition").Where("definition_id=?", defintionID).Exec()
	if err != nil {
		return err
	}
	if s.FileSystem != nil {
		return s.SaveToFlatFile()
	}
	return nil
}

func (s *DBStore) AllDefinitions() (regions.DefinitionList, error) {

	err := s.condLoadFromFlatFile()
	if err != nil {
		return nil, err
	}

	var ret regions.DefinitionList
	sess := s.conn.NewSession(nil)
	_, err = sess.Select("*").From(s.TablePrefix + "region_definition").OrderAsc("definition_id").Load(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// load from file if needsFileLoad is set
func (s *DBStore) condLoadFromFlatFile() error {
	s.rwmu.RLock()
	needsFileLoad := s.needsFileLoad
	s.rwmu.RUnlock()
	if needsFileLoad {
		return s.LoadFromFlatFile()
	}
	return nil
}

// LoadFromFlatFile reads the contents of the flat file and loads into the database.
func (s *DBStore) LoadFromFlatFile() error {
	s.rwmu.Lock()
	defer s.rwmu.Unlock()
	s.needsFileLoad = false

	f, err := s.FileSystem.Open(path.Join(s.FilePathPrefix, s.FilePathSuffix))
	if err != nil {
		// need to think about this more, but for now it is not an error if the file does yet exist
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var defs regions.DefinitionList
	err = yaml.Unmarshal(b, &defs)
	if err != nil {
		return err
	}

	sess := s.conn.NewSession(nil)
	tx, err := sess.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	// remove whatever's there
	_, err = tx.DeleteFrom(s.TablePrefix + "region_definition").Exec()
	if err != nil {
		return err
	}

	// TODO: batch this so we insert a bunch at a time, might perform better
	for _, def := range defs {
		_, err := tx.InsertInto(s.TablePrefix+"region_definition").
			Columns(
				"region_name", "sequence", "disabled", "template_name",
				"cond_include_paths", "cond_exclude_paths", "cond_template",
				"context_meta").
			Record(&def).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveToFlatFile writes the contents of the database to the flat file.
func (s *DBStore) SaveToFlatFile() error {

	s.rwmu.Lock()
	defer s.rwmu.Unlock()
	s.needsFileLoad = false

	var defs regions.DefinitionList
	sess := s.conn.NewSession(nil)
	_, err := sess.Select("*").From(s.TablePrefix + "region_definition").OrderAsc("definition_id").Load(&defs)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(&defs)
	if err != nil {
		return err
	}

	p := path.Join(s.FilePathPrefix, s.FilePathSuffix)
	f, err := s.FileSystem.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
