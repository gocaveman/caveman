package uifiles

import (
	"fmt"
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

// Minifier describes something that can minify.
type Minifier interface {
	// Minify performce minification on the given data, reading from r and writing to w.
	// The typeName param is just the type name, e.g. "css", "js" (not the mime type).
	Minify(typeName string, w io.Writer, r io.Reader) error
}

type DefaultMinifier struct {
	M *minify.M
}

func (m *DefaultMinifier) Minify(typeName string, w io.Writer, r io.Reader) error {
	switch typeName {
	case "css":
		return m.M.Minify("text/css", w, r)
	case "js":
		return m.M.Minify("application/javascript", w, r)
	}
	return fmt.Errorf("DefaultMinifier doesn't understand type %q", typeName)
}

func NewDefaultMinifier() *DefaultMinifier {

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)

	return &DefaultMinifier{M: m}

}
