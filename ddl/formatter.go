package ddl

import "strings"

// Formatter knows how to format a Stmt into one or more SQL strings.
// Most calls to Format() will return a single SQL string, but it is
// also possible that functionality unavailable in a particular database
// will need to be emulated with multiple statements.
type Formatter interface {
	Format(stmt Stmt) ([]string, error) // do formatting on a statement
	DriverName() string                 // get the driver name for this formatter, e.g. "mysql", "sqlite3"
}

type FormatterList []Formatter

func quoteIdent(s, quote string) string {
	part := strings.SplitN(s, ".", 2)
	if len(part) == 2 {
		return quoteIdent(part[0], quote) + "." + quoteIdent(part[1], quote)
	}
	return quote + s + quote
}
