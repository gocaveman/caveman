package ddl

import (
	"bytes"
	"fmt"
)

type Formatter interface {
	Format(stmt Stmt) (string, error)
	// each statement type
}

type FormatterList []Formatter

func NewFormatterList(formatters ...Formatter) FormatterList {
	return FormatterList(formatters)
}

// FIXME: this probably should move out to another package, and the import
// side effect should register and cause New() to include it by default if no formatter specified
type SQLite3Formatter struct {
	// FIXME: option for template prefix... what should the default be?
}

func NewSQLite3Formatter() *SQLite3Formatter {
	return &SQLite3Formatter{}
}

func (f *SQLite3Formatter) Format(stmt Stmt) (string, error) {

	var buf bytes.Buffer

	switch st := stmt.(type) {
	case *CreateTableStmt:
		ifNotExistsStr := ""
		if st.IfNotExistsValue {
			ifNotExistsStr = "IF NOT EXISTS "
		}
		fmt.Fprintf(&buf, `CREATE TABLE %s%s (`+"\n", ifNotExistsStr, st.NameValue)
		fmt.Fprintf(&buf, `  `+"\n", ifNotExistsStr, st.NameValue)
		fmt.Fprintf(&buf, `)`)
		return buf.String(), nil
	case *DropTableStmt:
		fmt.Fprintf(&buf, `DROP TABLE %s`, st.NameValue)
		return buf.String(), nil
	}

	return "", fmt.Errorf("unknown statement type %T", stmt)
}
