package ddl

import (
	"bytes"
	"fmt"
	"strings"
)

// FIXME: this probably should move out to another package, and the import
// side effect should register and cause New() to include it by default if no formatter specified
type SQLite3Formatter struct {
	// FIXME: option for template prefix... what should the default be?
}

func NewSQLite3Formatter() *SQLite3Formatter {
	return &SQLite3Formatter{}
}

func (f *SQLite3Formatter) Format(stmt Stmt) ([]string, error) {

	var buf bytes.Buffer

	switch st := stmt.(type) {

	case *CreateTableStmt:
		ifNotExistsStr := ""
		if st.IfNotExistsValue {
			ifNotExistsStr = "IF NOT EXISTS "
		}
		fmt.Fprintf(&buf, `CREATE TABLE %s%s (`+"\n", ifNotExistsStr, sqlite3QuoteIdent(st.NameValue))

		skipPKBlock := false
		for _, col := range st.Columns {

			// due to syntactic funk, we need to declare the primary key on the column for
			// autoincrement functionality and cannot have a separate PRIMARY KEY(field) block
			if col.DataTypeValue == BigIntAutoPK {
				skipPKBlock = true
			}

			colstr, err := sqlite3ColStr(col)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&buf, "    %s,\n", colstr)
		}

		if (!skipPKBlock) && len(st.PrimaryKeys) > 0 {
			fmt.Fprintf(&buf, "    PRIMARY KEY(")
			for idx, pk := range st.PrimaryKeys {
				fmt.Fprintf(&buf, "%s", sqlite3QuoteIdent(pk))
				if idx < len(st.PrimaryKeys)-1 {
					fmt.Fprintf(&buf, ",")
				}
			}
			fmt.Fprintf(&buf, "),\n")
		}

		for _, fk := range st.ForeignKeys {
			fmt.Fprintf(&buf, "    FOREIGN KEY(%s) REFERENCES %s(%s),",
				sqlite3QuoteIdent(fk.ColumnValue),
				sqlite3QuoteIdent(fk.OtherTableValue),
				sqlite3QuoteIdent(fk.OtherColumnValue),
			)
		}

		withoutRowidStr := ""
		for _, col := range st.Columns {
			if col.DataTypeValue == VarCharPK { // varchar primary key triggers WITHOUT ROWID
				withoutRowidStr = " WITHOUT ROWID"
				break
			}
		}
		if len(st.PrimaryKeys) > 1 { // multiple pks triggers WITHOUT ROWID
			withoutRowidStr = " WITHOUT ROWID"
		}

		// remove any trailing comma and close table definition
		fullStr := strings.TrimSuffix(strings.TrimSpace(buf.String()), ",") + "\n)" +
			withoutRowidStr
		return []string{fullStr}, nil

	case *DropTableStmt:
		fmt.Fprintf(&buf, `DROP TABLE %s`, sqlite3QuoteIdent(st.NameValue))
		return []string{buf.String()}, nil

	case *AlterTableRenameStmt:
		fmt.Fprintf(&buf, `ALTER TABLE %s RENAME TO %s`,
			sqlite3QuoteIdent(st.OldNameValue),
			sqlite3QuoteIdent(st.NewNameValue),
		)
		return []string{buf.String()}, nil

	case *AlterTableAddStmt:
		colStr, err := sqlite3ColStr(&st.DataTypeDef)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&buf, `ALTER TABLE %s ADD COLUMN %s`,
			sqlite3QuoteIdent(st.NameValue),
			colStr,
		)
		return []string{buf.String()}, nil

	case *CreateIndexStmt:
		uniqueStr := ""
		if st.UniqueValue {
			uniqueStr = " UNIQUE"
		}
		ifNotExistsStr := ""
		if st.IfNotExistsValue {
			ifNotExistsStr = " IF NOT EXISTS"
		}
		colStr := ""
		for _, colName := range st.ColumnNames {
			colStr += sqlite3QuoteIdent(colName) + ","
		}
		colStr = strings.TrimRight(colStr, ",")
		fmt.Fprintf(&buf, `CREATE%s INDEX%s %s ON %s(%s)`,
			uniqueStr,
			ifNotExistsStr,
			sqlite3QuoteIdent(st.NameValue),
			sqlite3QuoteIdent(st.TableNameValue),
			colStr,
		)
		return []string{buf.String()}, nil

	case *DropIndexStmt:
		fmt.Fprintf(&buf, `DROP INDEX %s`,
			sqlite3QuoteIdent(st.NameValue),
			// NOTE: SQLite3 does not need or allow the table name
		)
		return []string{buf.String()}, nil

	}

	return nil, fmt.Errorf("unknown statement type %T", stmt)
}

func sqlite3QuoteIdent(ident string) string {
	return quoteIdent(ident, `"`)
}

func sqlite3EncodeString(s string) string {
	// https://www.sqlite.org/faq.html
	return `'` + strings.Replace(s, `'`, `''`, -1) + `'`
}

func sqlite3ColStr(col *DataTypeDef) (string, error) {

	defaultStr := ""
	if col.DefaultValue != nil {
		if s, ok := col.DefaultValue.(string); ok {
			defaultStr = fmt.Sprintf(" DEFAULT %s", sqlite3EncodeString(s))
		} else {
			// FIXME: we should be more careful about what escaping and formatting is used here
			// and the various possible data types
			defaultStr = fmt.Sprintf(" DEFAULT %v", col.DefaultValue)
		}
	}
	// sqlite3 ignores lengths, don't bother: https://www.sqlite.org/datatype3.html
	// lengthStr := func(defaultLen int) string {
	// 	if col.LengthValue <= 0 {
	// 		if defaultLen > 0 {
	// 			return fmt.Sprintf("(%d)", defaultLen)
	// 		}
	// 		return ""
	// 	}
	// 	return fmt.Sprintf("(%d)", col.LengthValue)
	// }
	nullStr := " NOT NULL"
	if col.NullValue {
		nullStr = " NULL"
	}
	caseSensitiveStr := " COLLATE NOCASE"
	if col.CaseSensitiveValue {
		caseSensitiveStr = "" // will default to binary
	}

	switch col.DataTypeValue {
	case Custom:
		return fmt.Sprintf("%s %s", sqlite3QuoteIdent(col.NameValue), col.CustomSQLValue), nil
	case VarCharPK:
		// always case sensitive
		return fmt.Sprintf("%s VARCHAR%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case BigIntAutoPK:
		return fmt.Sprintf("%s INTEGER PRIMARY KEY AUTOINCREMENT", sqlite3QuoteIdent(col.NameValue)), nil
	case VarCharFK:
		// always case sensitive
		return fmt.Sprintf("%s VARCHAR%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case BigIntFK:
		return fmt.Sprintf("%s INTEGER%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case Int:
		return fmt.Sprintf("%s INTEGER%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case IntU:
		return fmt.Sprintf("%s UNSIGNED INTEGER%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case BigInt:
		return fmt.Sprintf("%s INTEGER%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case BigIntU:
		return fmt.Sprintf("%s UNSIGNED INTEGER%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case Double:
		return fmt.Sprintf("%s REAL%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case DateTime:
		// datetime values need to be text for things to work correctly with SQLite3
		return fmt.Sprintf("%s TEXT%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case VarChar:
		return fmt.Sprintf("%s VARCHAR%s%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, caseSensitiveStr, defaultStr), nil
	case Text:
		return fmt.Sprintf("%s TEXT%s%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, caseSensitiveStr, defaultStr), nil
	case Bool:
		// same as INTEGER but whatever
		return fmt.Sprintf("%s BOOLEAN%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case Blob:
		return fmt.Sprintf("%s BLOB%s%s", sqlite3QuoteIdent(col.NameValue), nullStr, defaultStr), nil

	}

	return "", fmt.Errorf("unknown DataType: %v", col.DataTypeValue)
}
