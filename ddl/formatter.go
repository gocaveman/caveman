package ddl

import "strings"

// Formatter knows how to format a Stmt into one or more SQL strings.
// Most calls to Format() will return a single SQL string, but it is
// also possible that functionality unavailable in a particular database
// will need to be emulated with multiple statements.
type Formatter interface {
	Format(stmt Stmt) ([]string, error)
}

type FormatterList []Formatter

func NewFormatterList(formatters ...Formatter) FormatterList {
	return FormatterList(formatters)
}

func quoteIdent(s, quote string) string {
	part := strings.SplitN(s, ".", 2)
	if len(part) == 2 {
		return quoteIdent(part[0], quote) + "." + quoteIdent(part[1], quote)
	}
	return quote + s + quote
}
