package migratedbr

import (
	"fmt"

	"github.com/gocraft/dbr"
)

func New(driverName, dsn string) (*DbrVersioner, error) {
	return NewTable(driverName, dsn, "migration_state")
}

func NewTable(driverName, dsn, tableName string) (*DbrVersioner, error) {

	conn, err := dbr.Open(driverName, dsn, nil)
	if err != nil {
		return nil, err
	}

	// TODO: might need to do variations on this but for now this should work for mysql, postgres and sqlite3
	// FIXME: category changed to 128 due to some obscure MariaDB encoding issue, needs more thought
	// https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=886756
	_, err = conn.DB.Exec(`
CREATE TABLE IF NOT EXISTS ` + tableName + ` (
	category varchar(128),
	version varchar(255),
	status varchar(255),
	PRIMARY KEY (category)
)
`)
	if err != nil {
		return nil, err
	}

	return &DbrVersioner{
		Connection: conn,
		TableName:  tableName,
	}, nil
}

type DbrVersioner struct {
	Connection *dbr.Connection
	TableName  string
}

func (v *DbrVersioner) Categories() ([]string, error) {

	sess := v.Connection.NewSession(nil)
	recs := []struct {
		Category string `db:"category"`
	}{}
	n, err := sess.Select("category").From(v.TableName).Load(&recs)
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0, n)
	for _, rec := range recs {
		ret = append(ret, rec.Category)
	}

	return ret, nil
}

func (v *DbrVersioner) Version(category string) (string, error) {

	versionName := ""

	sess := v.Connection.NewSession(nil)
	err := sess.Select("version").From(v.TableName).Where("category = ?", category).LoadOne(&versionName)

	// treat missing row as empty version
	if err == dbr.ErrNotFound {
		return "", nil
	}

	return versionName, err
}

func (v *DbrVersioner) StartVersionChange(category, currentVersion string) error {

	sess := v.Connection.NewSession(nil)

	// tx, err := v.DB.Begin()
	// if err != nil {
	// 	return err
	// }
	// defer tx.Rollback()

	versionName := ""

	err := sess.Select("version").From(v.TableName).Where("category = ?", category).LoadOne(&versionName)
	if err == dbr.ErrNotFound {

		_, err := sess.InsertInto(v.TableName).Columns("category", "version", "status").Values(category, "", "none").Exec()
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	// row := tx.QueryRow(`SELECT version FROM `+v.TableName+` WHERE category = ?`, category)
	// err = row.Scan(&versionName)
	// if err == sql.ErrNoRows {
	// 	_, err := tx.Exec(`INSERT INTO `+v.TableName+`(category, version, status) VALUES(?,?,?)`, category, "", "none")
	// 	if err != nil {
	// 		return err
	// 	}
	// } else if err != nil {
	// 	return err
	// }

	if versionName != currentVersion {
		return fmt.Errorf("incorrect version, found %q expected %q", versionName, currentVersion)
	}

	res, err := sess.Update(v.TableName).
		Set("status", "inprogress").
		Where(dbr.And(
			dbr.Eq("category", category),
			dbr.Eq("status", "none"),
			dbr.Eq("version", currentVersion),
		)).
		Exec()

	if err != nil {
		return err
	}

	// res, err := tx.Exec(`UPDATE `+v.TableName+` SET status = ? WHERE category = ? AND status = ? AND version = ?`, "inprogress", category, "none", currentVersion)
	// if err != nil {
	// 	return err
	// }

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("StartVersionChange UPDATE statement returned num rows affected %d (expected 1)", n)
	}

	// err = tx.Commit()
	// if err != nil {
	// 	return err
	// }

	return nil

}

func (v *DbrVersioner) EndVersionChange(category, newVersionName string) error {

	// tx, err := v.DB.Begin()
	// if err != nil {
	// 	return err
	// }
	// defer tx.Rollback()

	sess := v.Connection.NewSession(nil)

	res, err := sess.Update(v.TableName).
		Set("version", newVersionName).
		Set("status", "none").
		Where(dbr.And(dbr.Eq("category", category), dbr.Eq("status", "inprogress"))).
		Exec()

	// res, err := tx.Exec(`UPDATE `+v.TableName+` SET version = ?, status = ? WHERE category = ? AND status = ?`, newVersionName, "none", category, "inprogress")
	// if err != nil {
	// 	return err
	// }

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return fmt.Errorf("EndVersionChange UPDATE statement returned num rows affected %d (expected 1)", n)
	}

	// err = tx.Commit()
	// if err != nil {
	// 	return err
	// }

	return nil
}
