package ddl

import (
	"bytes"
	"fmt"
	"strings"
)

type MySQLFormatter struct {
	Template bool // set to true to enable template output (supports prefixes)
}

// NewMySQLFormatter returns a new MySQLFormatter. If the template argument
// is true then table prefixes (and any other templatable features)
// will be output in Go template form, for use with migrations.  Passing false
// will produce raw SQL that can be executed directly.
func NewMySQLFormatter(template bool) *MySQLFormatter {
	return &MySQLFormatter{Template: template}
}

func (f *MySQLFormatter) tmplPrefix() string {
	if f.Template {
		return "{{.TablePrefix}}"
	}
	return ""
}

func (f *MySQLFormatter) DriverName() string {
	return "mysql"
}

func (f *MySQLFormatter) Format(stmt Stmt) ([]string, error) {

	var buf bytes.Buffer

	switch st := stmt.(type) {

	case *CreateTableStmt:
		ifNotExistsStr := ""
		if st.IfNotExistsValue {
			ifNotExistsStr = "IF NOT EXISTS "
		}
		fmt.Fprintf(&buf, `CREATE TABLE %s%s (`+"\n", ifNotExistsStr, mysqlQuoteIdent(f.tmplPrefix()+st.NameValue))

		for _, col := range st.Columns {

			colstr, err := mysqlColStr(col)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&buf, "    %s,\n", colstr)
		}

		if len(st.PrimaryKeys) > 0 {
			fmt.Fprintf(&buf, "    PRIMARY KEY(")
			for idx, pk := range st.PrimaryKeys {
				fmt.Fprintf(&buf, "%s", mysqlQuoteIdent(pk))
				if idx < len(st.PrimaryKeys)-1 {
					fmt.Fprintf(&buf, ",")
				}
			}
			fmt.Fprintf(&buf, "),\n")
		}

		for _, fk := range st.ForeignKeys {
			fmt.Fprintf(&buf, "    FOREIGN KEY(%s) REFERENCES %s(%s),",
				mysqlQuoteIdent(fk.ColumnValue),
				mysqlQuoteIdent(f.tmplPrefix()+fk.OtherTableValue),
				mysqlQuoteIdent(fk.OtherColumnValue),
			)
		}

		// Use utf8mb4 as the character set for everything not explicitly
		// set on a column. Let the db choose the default collation, since
		// it will use the most recent case-insensitive unicode comparision,
		// which is usually exactly what you want.
		tableSuffixStr := " /*!50508 CHARSET=utf8mb4 */"

		// remove any trailing comma and close table definition
		fullStr := strings.TrimSuffix(strings.TrimSpace(buf.String()), ",") + "\n)" +
			tableSuffixStr
		return []string{fullStr}, nil

	case *DropTableStmt:
		fmt.Fprintf(&buf, `DROP TABLE %s`, mysqlQuoteIdent(f.tmplPrefix()+st.NameValue))
		return []string{buf.String()}, nil

	case *AlterTableRenameStmt:
		fmt.Fprintf(&buf, `ALTER TABLE %s RENAME TO %s`,
			mysqlQuoteIdent(f.tmplPrefix()+st.OldNameValue),
			mysqlQuoteIdent(f.tmplPrefix()+st.NewNameValue),
		)
		return []string{buf.String()}, nil

	case *AlterTableAddStmt:
		colStr, err := mysqlColStr(&st.DataTypeDef)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&buf, `ALTER TABLE %s ADD COLUMN %s`,
			mysqlQuoteIdent(f.tmplPrefix()+st.NameValue),
			colStr,
		)
		return []string{buf.String()}, nil

	case *CreateIndexStmt:
		uniqueStr := ""
		if st.UniqueValue {
			uniqueStr = " UNIQUE"
		}
		// for, just error on this, rather than screwing around with the MariaDB/MySQL fiasco,
		// the whole point of migrations is to avoid this crap anyway - maybe if not exists
		// functionality should just be removed...
		if st.IfNotExistsValue {
			return nil, fmt.Errorf("CREATE INDEX IF NOT EXISTS is not supported by MySQL")
		}
		ifNotExistsStr := ""
		// if st.IfNotExistsValue {
		// 	ifNotExistsStr = " IF NOT EXISTS"
		// }
		colStr := ""
		for _, colName := range st.ColumnNames {
			colStr += mysqlQuoteIdent(colName) + ","
		}
		colStr = strings.TrimRight(colStr, ",")
		fmt.Fprintf(&buf, `CREATE%s INDEX%s %s ON %s(%s)`,
			uniqueStr,
			ifNotExistsStr,
			mysqlQuoteIdent(f.tmplPrefix()+st.NameValue),
			mysqlQuoteIdent(f.tmplPrefix()+st.TableNameValue),
			colStr,
		)
		return []string{buf.String()}, nil

	case *DropIndexStmt:
		fmt.Fprintf(&buf, `DROP INDEX %s ON %s`,
			mysqlQuoteIdent(f.tmplPrefix()+st.NameValue),
			mysqlQuoteIdent(f.tmplPrefix()+st.TableNameValue),
		)
		return []string{buf.String()}, nil

	}

	return nil, fmt.Errorf("unknown statement type %T", stmt)
}

func mysqlQuoteIdent(ident string) string {
	return quoteIdent(ident, "`")
}

// https://dev.mysql.com/doc/refman/8.0/en/string-literals.html
var mysqlEncodeStringReplacer = strings.NewReplacer(`'`, `\'`, `\`, `\\`)

func mysqlEncodeString(s string) string {
	return `'` + mysqlEncodeStringReplacer.Replace(s) + `'`
}

func mysqlColStr(col *DataTypeDef) (string, error) {

	defaultStr := ""
	if col.DefaultValue != nil {
		if s, ok := col.DefaultValue.(string); ok {
			defaultStr = fmt.Sprintf(" DEFAULT %s", mysqlEncodeString(s))
		} else {
			// FIXME: we should be more careful about what escaping and formatting is used here
			// and the various possible data types
			defaultStr = fmt.Sprintf(" DEFAULT %v", col.DefaultValue)
		}
	}
	nullStr := " NOT NULL"
	if col.NullValue {
		nullStr = " NULL"
	}
	caseSensitiveStr := "" // by default it will be case insensitive, no need to specify anything
	if col.CaseSensitiveValue {
		caseSensitiveStr = " /*!50508 COLLATE utf8mb4_bin */" // case sensitive is done by changing collation to binary
	}

	// pk/fk columns are different, make them ascii and case sensitive
	keyColSuffix := " CHARACTER SET ascii COLLATE ascii_bin"

	switch col.DataTypeValue {
	case Custom:
		return fmt.Sprintf("%s %s", mysqlQuoteIdent(col.NameValue), col.CustomSQLValue), nil
	case VarCharPK:
		// always case sensitive
		return fmt.Sprintf("%s VARCHAR(64)%s%s%s", mysqlQuoteIdent(col.NameValue), keyColSuffix, nullStr, defaultStr), nil
	case BigIntAutoPK:
		return fmt.Sprintf("%s BIGINT NOT NULL AUTO_INCREMENT", mysqlQuoteIdent(col.NameValue)), nil
	case VarCharFK:
		// always case sensitive
		return fmt.Sprintf("%s VARCHAR(64)%s%s%s", mysqlQuoteIdent(col.NameValue), keyColSuffix, nullStr, defaultStr), nil
	case BigIntFK:
		return fmt.Sprintf("%s BIGINT%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case Int:
		return fmt.Sprintf("%s INTEGER%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case IntU:
		return fmt.Sprintf("%s INTEGER UNSIGNED %s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case BigInt:
		return fmt.Sprintf("%s BIGINT%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case BigIntU:
		return fmt.Sprintf("%s BIGINT UNSIGNED%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case Double:
		return fmt.Sprintf("%s REAL%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case DateTime:
		// use native DATETIME type and add the extra sub-second precision if supported
		return fmt.Sprintf("%s DATETIME/*!50604 (6) */%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case VarChar:
		// default string length - big enough to handle most human-entered values, but
		// small enough to fit under the InnoDB 767 index max byte length (i.e. if you
		// set this to 255 with utf8mb4 and try to index it the index creation will fail,
		// it's pain in the ass - 128 is a better default)
		lengthStr := "(128)"
		if col.LengthValue > 0 {
			lengthStr = fmt.Sprintf("(%d)", col.LengthValue)
		}
		return fmt.Sprintf("%s VARCHAR%s%s%s%s", mysqlQuoteIdent(col.NameValue), lengthStr, nullStr, caseSensitiveStr, defaultStr), nil
	case Text:
		lengthStr := "" // no explicit length unless specified
		if col.LengthValue > 0 {
			lengthStr = fmt.Sprintf("(%d)", col.LengthValue)
		}
		return fmt.Sprintf("%s TEXT%s%s%s%s", mysqlQuoteIdent(col.NameValue), lengthStr, nullStr, caseSensitiveStr, defaultStr), nil
	case Bool:
		return fmt.Sprintf("%s BOOLEAN%s%s", mysqlQuoteIdent(col.NameValue), nullStr, defaultStr), nil
	case Blob:
		lengthStr := "" // no explicit length unless specified
		if col.LengthValue > 0 {
			lengthStr = fmt.Sprintf("(%d)", col.LengthValue)
		}
		return fmt.Sprintf("%s BLOB%s%s%s", mysqlQuoteIdent(col.NameValue), lengthStr, nullStr, defaultStr), nil

	}

	return "", fmt.Errorf("unknown DataType: %v", col.DataTypeValue)
}
