// Manage and apply/unapply database schema changes, grouped by category and db driver.
package migrate

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
	"text/template"
)

// multiple named migrations
// must be able to test migrations
// in a cluster it must not explode when mulitiple servers try to migrate at the same time

// registry? how do the various components ensure their migrations get done

// what about "minimum veresion required for this code to run"

// Versioner interface is implemented by things that record which version is applied to a database and category.
type Versioner interface {
	Categories() ([]string, error)
	Version(category string) (string, error)
	StartVersionChange(category, curVersionName string) error
	EndVersionChange(category, newVersionName string) error
}

// NewRunner creates a Runner.
func NewRunner(driverName, dsn string, versioner Versioner, migrations MigrationList) *Runner {
	return &Runner{
		DriverName: driverName,
		DSN:        dsn,
		Versioner:  versioner,
		Migrations: migrations,
	}
}

// Runner is used to apply migration changes.
type Runner struct {
	DriverName string
	DSN        string
	Versioner
	Migrations MigrationList
}

// FIXME: Runner should also be able to tell us if there are outstanding migrations
// that need to be run. In local environments we'll probably just call RunAllUpToLatest()
// at startup, but for staging and production we'll instead just check at startup
// and show messages if migrations are needed.

// CheckResult is a list of CheckResultItem
type CheckResult []CheckResultItem

// CheckResultItem gives the current and latest versions for a specific driver/dsn/category.
type CheckResultItem struct {
	DriverName     string
	DSN            string
	Category       string
	CurrentVersion string
	LatestVersion  string
}

// IsCurrent returns true if latest version is current version.
func (i CheckResultItem) IsCurrent() bool {
	return i.CurrentVersion == i.LatestVersion
}

// CheckAll checks all categories and compares the current version to the latest version
// and returns the results.  If you pass true then all results will be returned, otherwise
// only results where the version is not current will be returned.
func (r *Runner) CheckAll(returnAll bool) (CheckResult, error) {

	var res CheckResult

	ms := r.Migrations.WithDriverName(r.DriverName)

	cats := ms.Categories()
	for _, cat := range cats {
		msc := ms.WithCategory(cat).Sorted()
		latestVersion := ""
		if len(msc) > 0 {
			latestVersion = msc[len(msc)-1].Version()
		}
		currentVersion, err := r.Versioner.Version(cat)
		if err != nil {
			return nil, err
		}
		res = append(res, CheckResultItem{
			DriverName:     r.DriverName,
			DSN:            r.DSN,
			Category:       cat,
			CurrentVersion: currentVersion,
			LatestVersion:  latestVersion,
		})
	}

	if returnAll {
		return res, nil
	}

	var res2 CheckResult
	for _, item := range res {
		if !item.IsCurrent() {
			res2 = append(res2, item)
		}
	}

	return res2, nil

}

// RunAllUpToLatest runs all migrations for all categories up to the latest version.
func (r *Runner) RunAllUpToLatest() error {
	cats := r.Migrations.Categories()
	for _, cat := range cats {
		err := r.RunUpToLatest(cat)
		if err != nil {
			return err
		}
	}
	return nil
}

// RunUpToLatest runs all migrations for a specific category up to the latest.
func (r *Runner) RunUpToLatest(category string) error {

	mc := r.Migrations.WithDriverName(r.DriverName).WithCategory(category).Sorted()

	if len(mc) == 0 {
		return nil
	}

	ver := mc[len(mc)-1].Version()

	return r.RunUpTo(category, ver)
}

// RunTo will run the migrations for a speific category up or down to a specific version.
func (r *Runner) RunTo(category, targetVersion string) error {

	if targetVersion == "" {
		return fmt.Errorf("RunTo with empty target version not allowed, call RunDownTo() explicitly")
	}

	mc := r.Migrations.WithDriverName(r.DriverName).WithCategory(category).Sorted()
	curVer, err := r.Versioner.Version(category)
	if err != nil {
		return err
	}

	if curVer == targetVersion {
		return nil
	}

	// empty cur ver always means it's an up
	if curVer == "" {
		return r.RunUpTo(category, targetVersion)
	}

	curIdx := -1
	tgtIdx := -1
	for i := range mc {
		if mc[i].Version() == curVer {
			curIdx = i
		}
		if mc[i].Version() == targetVersion {
			tgtIdx = i
		}
	}

	if curIdx < 0 {
		return fmt.Errorf("current version %q not found in category %q", curVer, category)
	}
	if tgtIdx < 0 {
		return fmt.Errorf("target version %q not found in category %q", targetVersion, category)
	}

	if curIdx > tgtIdx {
		return r.RunDownTo(category, targetVersion)
	}

	return r.RunUpTo(category, targetVersion)
}

// RunUpTo runs migrations up to a specific version. Will only run up, will error if
// this version is lower than the current one.
func (r *Runner) RunUpTo(category, targetVersion string) error {

	curVer, err := r.Versioner.Version(category)
	if err != nil {
		return err
	}
	ml := r.Migrations.WithDriverName(r.DriverName).WithCategory(category).Sorted()
	if len(ml) == 0 {
		return nil
	}

	if !ml.HasVersion(targetVersion) {
		return fmt.Errorf("version %q not found", targetVersion)
	}

	active := curVer == "" // start active if empty current version
	for _, m := range ml {

		if curVer == m.Version() {
			active = true
			continue
		}

		if !active {
			continue
		}

		err := r.Versioner.StartVersionChange(category, curVer)
		if err != nil {
			return err
		}

		err = m.ExecUp(r.DSN)
		if err != nil {
			// try to revert the version
			err2 := r.Versioner.EndVersionChange(category, curVer)
			if err2 != nil { // just log the error in this case, so the orignal error is preserved
				log.Printf("EndVersionChange returned error: %v", err2)
			}
			return fmt.Errorf("ExecUp(%q) error: %v", r.DSN, err)
		}

		// Update version to the migration we just ran.
		// NOTE: This will leave things in an inconsistent state if it errors but nothing we can do...
		err = r.Versioner.EndVersionChange(category, m.Version())
		if err != nil {
			return err
		}

		curVer = m.Version()

		if targetVersion == m.Version() {
			break
		}

	}

	return nil

}

// RunUpTo runs migrations down to a specific version. Will only run down, will error if
// this version is higher than the current one.
func (r *Runner) RunDownTo(category, targetVersion string) error {

	// log.Printf("RunDownTo %q %q", category, targetVersion)

	curVer, err := r.Versioner.Version(category)
	if err != nil {
		return err
	}
	ml := r.Migrations.WithDriverName(r.DriverName).WithCategory(category).Sorted()
	sort.Sort(sort.Reverse(ml))
	if len(ml) == 0 {
		return nil
	}

	if targetVersion != "" && !ml.HasVersion(targetVersion) {
		return fmt.Errorf("version %q not found", targetVersion)
	}

	active := false
	for mlidx, m := range ml {

		// check for target version, in which case we're done
		if targetVersion == m.Version() {
			break
		}

		// if we're on current version, mark active and continue
		if curVer == m.Version() {
			active = true
		}

		if !active {
			continue
		}

		err := r.Versioner.StartVersionChange(category, curVer)
		if err != nil {
			return err
		}

		err = m.ExecDown(r.DSN)
		if err != nil {
			// try to revert the version
			err2 := r.Versioner.EndVersionChange(category, curVer)
			if err2 != nil { // just log the error in this case, so the orignal error is preserved
				log.Printf("EndVersionChange returned error: %v", err2)
			}
			return fmt.Errorf("ExecDown(%q) error: %v", r.DSN, err)
		}

		// Update version to the NEXT migration in the sequence or empty string if at the end
		nextLowerVersion := ""
		if mlidx+1 < len(ml) {
			nextLowerVersion = ml[mlidx+1].Version()
		}

		// NOTE: This will leave things in an inconsistent state if it errors but nothing we can do...
		err = r.Versioner.EndVersionChange(category, nextLowerVersion)
		if err != nil {
			return err
		}

		curVer = m.Version()

	}

	return nil
}

// Migration represents a driver name, category and version and functionality to perform an
// "up" and "down" to and from this version.  See SQLMigration and FuncsMigration for implementations.
type Migration interface {
	DriverName() string
	Category() string
	Version() string
	ExecUp(dsn string) error
	ExecDown(dsn string) error
}

// MigrationList is a slice of Migration
type MigrationList []Migration

func (p MigrationList) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for _, m := range p {
		fmt.Fprintf(&buf, `{"type":%q,"driverName":%q,"category":%q,"version":%q},`,
			fmt.Sprintf("%T", m), m.DriverName(), m.Category(), m.Version())
	}
	if len(p) > 0 {
		buf.Truncate(buf.Len() - 1) // remove trailing comma
	}
	buf.WriteString("]")
	return buf.String()
}

func (p MigrationList) Len() int      { return len(p) }
func (p MigrationList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p MigrationList) Less(i, j int) bool {
	mli := p[i]
	mlj := p[j]
	return mli.DriverName()+" | "+mli.Category()+"|"+mli.Version() <
		mlj.DriverName()+" | "+mlj.Category()+"|"+mlj.Version()

}

// HasVersion returns true if this has has an item with this version name.
func (ml MigrationList) HasVersion(ver string) bool {
	vers := ml.Versions()
	for _, v := range vers {
		if v == ver {
			return true
		}
	}
	return false
}

// Categories returns a unique list of categories from these Migrations.
func (ml MigrationList) Categories() []string {
	var ret []string
	catMap := make(map[string]bool)
	for _, m := range ml {
		cat := m.Category()
		if !catMap[cat] {
			catMap[cat] = true
			ret = append(ret, cat)
		}
	}
	sort.Strings(ret)
	return ret
}

// Versions returns a unique list of versions from these Migrations.
func (ml MigrationList) Versions() []string {
	var ret []string
	verMap := make(map[string]bool)
	for _, m := range ml {
		ver := m.Version()
		if !verMap[ver] {
			verMap[ver] = true
			ret = append(ret, ver)
		}
	}
	sort.Strings(ret)
	return ret
}

// WithDriverName returns a new list filtered to only include migrations with the specified driver name.
func (ml MigrationList) WithDriverName(driverName string) MigrationList {

	var ret MigrationList

	for _, m := range ml {
		if m.DriverName() == driverName {
			ret = append(ret, m)
		}
	}

	return ret
}

// WithCategory returns a new list filtered to only include migrations with the specified category.
func (ml MigrationList) WithCategory(category string) MigrationList {

	var ret MigrationList

	for _, m := range ml {
		if m.Category() == category {
			ret = append(ret, m)
		}
	}

	return ret
}

// ExcludeCategory returns a new MigrationList without records for the specified category.
func (ml MigrationList) ExcludeCategory(category string) MigrationList {
	var ret MigrationList
	for _, m := range ml {
		if m.Category() != category {
			ret = append(ret, m)
		}
	}
	return ret
}

// Sorted returns a sorted copy of the list.  Sequence is by driver, category and then version.
func (ml MigrationList) Sorted() MigrationList {

	ml2 := make(MigrationList, len(ml))
	copy(ml2, ml)

	sort.Sort(ml2)

	return ml2

}

// LoadSQLMigrationsHFS is like LoadMigrations but loads from an http.FileSystem, so you can control the file source.
func LoadSQLMigrationsHFS(hfs http.FileSystem, dir string) (MigrationList, error) {

	f, err := hfs.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	migMap := make(map[string]*SQLMigration)

	for _, fi := range fis {

		// skip dirs
		if fi.IsDir() {
			continue
		}

		fname := path.Base(fi.Name())

		// skip anything not a .sql file
		if path.Ext(fname) != ".sql" {
			continue
		}

		parts := strings.Split(strings.TrimSuffix(fname, ".sql"), "-")
		if !(len(parts) == 4 && (parts[3] == "up" || parts[3] == "down")) {
			return nil, fmt.Errorf("LoadSQLMigrationsHFS(hfs=%v, dir=%q): filename %q is wrong format, expected exactly 4 parts separated by dashes and the last part must be 'up' or 'down'", hfs, dir, fname)
		}

		key := strings.Join(parts[:3], "-")

		// check for existing migration so we can fill in either up or down
		var sqlMigration *SQLMigration
		if migMap[key] != nil {
			sqlMigration = migMap[key]
		} else {
			sqlMigration = &SQLMigration{
				DriverNameValue: parts[0],
				CategoryValue:   parts[1],
				VersionValue:    parts[2],
			}
		}
		migMap[key] = sqlMigration

		// figure out up/down part
		var stmts *[]string
		if parts[3] == "up" {

			if sqlMigration.UpSQL != nil {
				return nil, fmt.Errorf("LoadSQLMigrationsHFS(hfs=%v, dir=%q): filename %q - more than one up migration found", hfs, dir, fname)
			}

			stmts = &sqlMigration.UpSQL

		} else { // down

			if sqlMigration.DownSQL != nil {
				return nil, fmt.Errorf("LoadSQLMigrationsHFS(hfs=%v, dir=%q): filename %q - more than one down migration found", hfs, dir, fname)
			}

			stmts = &sqlMigration.DownSQL

		}

		f, err := hfs.Open(path.Join(dir, fname))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// make sure it's non-nil
		(*stmts) = make([]string, 0)

		r := bufio.NewReader(f)
		var thisStmt bytes.Buffer
		for {
			line, err := r.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("LoadSQLMigrationsHFS(hfs=%v, dir=%q): filename %q error reading file: %v", hfs, dir, fname, err)
			}
			thisStmt.Write(line)

			// check for end of statement
			if bytes.HasSuffix(bytes.TrimSpace(line), []byte(";")) {
				thisStmtStr := thisStmt.String()
				if len(strings.TrimSpace(thisStmtStr)) > 0 {
					(*stmts) = append((*stmts), thisStmtStr)
				}
				thisStmt.Truncate(0)
			}

		}

		thisStmtStr := thisStmt.String()
		if len(strings.TrimSpace(thisStmtStr)) > 0 {
			(*stmts) = append((*stmts), thisStmtStr)
		}

	}

	var ret MigrationList
	for _, m := range migMap {
		ret = append(ret, m)
	}
	ret = ret.Sorted()

	return ret, nil
}

// LoadSQLMigrations loads migrations from the specified directory.  File names are
// expected to be in exactly four parts each separated with a dash and have a .sql extension:
// `driver-category-version-up.sql` is the format for up migrations, the corresponding down
// migration is the same but with 'down' instead of 'up'.  Another example:
//
// mysql-users-2017120301_create-up.sql
//
// mysql-users-2017120301_create-down.sql
//
// Both the up and down files must be present for a migration or an error will be returned.
// Files are plain text with SQL in them.  Each line that ends with a semicolon (ignoring whitespace after)
// will be treated as a separate SQL statement.
func LoadSQLMigrations(dir string) (MigrationList, error) {
	return LoadSQLMigrationsHFS(http.Dir(dir), "/")
}

// SQLMigration implements Migration with a simple slice of SQL strings for the up and down migration steps.
// Common migrations which are just one or more static SQL statements can be implemented easily using SQLMigration.
type SQLMigration struct {
	DriverNameValue string
	CategoryValue   string
	VersionValue    string
	UpSQL           []string
	DownSQL         []string
}

// NewWithDriverName as a convenience returns a copy with DriverNameValue set to the specified value.
func (m *SQLMigration) NewWithDriverName(driverName string) *SQLMigration {
	ret := *m
	ret.DriverNameValue = driverName
	return &ret
}

func (m *SQLMigration) DriverName() string { return m.DriverNameValue }
func (m *SQLMigration) Category() string   { return m.CategoryValue }
func (m *SQLMigration) Version() string    { return m.VersionValue }

func (m *SQLMigration) exec(dsn string, stmts []string) error {

	db, err := sql.Open(m.DriverNameValue, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	for n, s := range stmts {
		_, err := db.Exec(s)
		if err != nil {
			return fmt.Errorf("SQLMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) Exec on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, s)
		}
	}

	return nil
}

func (m *SQLMigration) ExecUp(dsn string) error {
	return m.exec(dsn, m.UpSQL)
}

func (m *SQLMigration) ExecDown(dsn string) error {
	return m.exec(dsn, m.DownSQL)
}

// SQLTmplMigration implements Migration with a simple slice of strings which are
// interpreted as Go templates with SQL as the up and down migration steps.
// This allows you to customize the SQL with things like table prefixes.
// Template are executed using text/template and the SQLTmplMigration instance
// is passed as the data to the Execute() call.
type SQLTmplMigration struct {
	DriverNameValue string
	CategoryValue   string
	VersionValue    string
	UpSQL           []string
	DownSQL         []string

	// a common reason to use SQLTmplMigration is be able to configure the table prefix
	TablePrefix string `autowire:"db.TablePrefix,optional"`
	// other custom data needed by the template(s) can go here
	Data interface{}
}

// NewWithDriverName as a convenience returns a copy with DriverNameValue set to the specified value.
func (m *SQLTmplMigration) NewWithDriverName(driverName string) *SQLTmplMigration {
	ret := *m
	ret.DriverNameValue = driverName
	return &ret
}

func (m *SQLTmplMigration) DriverName() string { return m.DriverNameValue }
func (m *SQLTmplMigration) Category() string   { return m.CategoryValue }
func (m *SQLTmplMigration) Version() string    { return m.VersionValue }

func (m *SQLTmplMigration) tmplExec(dsn string, stmts []string) error {

	db, err := sql.Open(m.DriverNameValue, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	for n, s := range stmts {

		t := template.New("sql")
		t, err := t.Parse(s)
		if err != nil {
			return fmt.Errorf("SQLTmplMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) template parse on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, s)
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, m)
		if err != nil {
			return fmt.Errorf("SQLTmplMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) template execute on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, s)
		}

		newS := buf.String()

		_, err = db.Exec(newS)
		if err != nil {
			return fmt.Errorf("SQLTmplMigration (driverName=%q, category=%q, version=%q, stmtidx=%d) Exec on dsn=%q failed with error: %v\nSQL Statement:\n%s",
				m.DriverNameValue, m.CategoryValue, m.VersionValue, n, dsn, err, newS)
		}
	}

	return nil
}

func (m *SQLTmplMigration) ExecUp(dsn string) error {
	return m.tmplExec(dsn, m.UpSQL)
}

func (m *SQLTmplMigration) ExecDown(dsn string) error {
	return m.tmplExec(dsn, m.DownSQL)
}

// NewFuncsMigration makes and returns a new FuncsMigration pointer with the data you provide.
func NewFuncsMigration(driverName, category, version string, upFunc, downFunc MigrationFunc) *FuncsMigration {
	return &FuncsMigration{
		DriverNameValue: driverName,
		CategoryValue:   category,
		VersionValue:    version,
		UpFunc:          upFunc,
		DownFunc:        downFunc,
	}
}

type MigrationFunc func(driverName, dsn string) error

// FuncsMigration is a Migration implementation that simply has up and down migration functions.
type FuncsMigration struct {
	DriverNameValue string
	CategoryValue   string
	VersionValue    string
	UpFunc          MigrationFunc
	DownFunc        MigrationFunc
}

func (m *FuncsMigration) DriverName() string { return m.DriverNameValue }
func (m *FuncsMigration) Category() string   { return m.CategoryValue }
func (m *FuncsMigration) Version() string    { return m.VersionValue }
func (m *FuncsMigration) ExecUp(dsn string) error {
	return m.UpFunc(m.DriverNameValue, dsn)
}
func (m *FuncsMigration) ExecDown(dsn string) error {
	return m.DownFunc(m.DriverNameValue, dsn)
}
