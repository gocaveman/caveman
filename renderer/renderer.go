// The renderer package provides functionality to convert the contents of files into servable URLs using Go HTML templating.
// Rendering is consist of the following parts:
//
// A Loader is responsible for taking a filename and returning a stream of bytes that can be interpreted by Go's html/template package.
// FileExtLoader allows for easily calling different loaders based on file extension.  GohtmlLoader reads in a file and strips off
// the YAML meta section (delimited by +++ before and after, and replacing it with a comment so the line numbers stay corret) and
// returns the rest.  The markdown package provides an implementation which reads markdown and converts to Go HTML template language.
//
// A Renderer is able to parse a filename (it has a Loader on it which is uses to load files) and return a Template instance
// (from html/template).  The TemplateMaker interface can be used to customize settings such as the FuncMap used and delimiters.
// A Renderer can be called directly from an http.Handler in order to render a specific page and this is the recommended way
// for handlers which need to do work and then render a page to function.  You can either Parse() and then execute the template
// yourself or ParseAndExecute() provides a more convenient way for the common case.
//
// A FileNamer takes a path from a URL and returns a list of possible filenames.  This distinction between URL paths and filenames
// is important - a Renderer is only aware of exact filenames.  Whereas a RenderHandler (below) uses a FileName to convert from
// the path that came in on a request to the underlying filename and they are not the same.  Examples include "/" -> "/index.gohtml",
// "/somepage" -> "/somepage.gohtml".  See NewDefaultFileNamer() for details.
//
// The RenderHandler provides functionality to use a FileNamer and a Renderer to serve pages in the manner you'd expect - you
// request a file and the page is served back to the browser.
//
package renderer

import (
	"html/template"
	"io"
)

// some specific things to break out:
// - path <-> filename mapping, needs to be bidirectional and pluggable
//   filenames can be multiple options
// - file contents -> go template conversion (need a default one plus funky stuff like markdown);
//   also should support redirects, e.g. /index -> / (but also look at this in relation to the
//   /section/ vs /section issue, which may not be ble to be handled here but might require a separate 404
//   handler - sort out the difference...)
// - what do we do about metadata... that one is odd, i'm tempted to completely leave it out of this step...
//   it could be that the pages module thing runs before and picks the metadata from whatever it's doing
//   and attaches it to the context and that's it, renderer has nothing to do with it.
// - require functionality - but this should be pluggable, rather than having a UIRequirer right on Renderer,
//   we could do some sort of generic post processor (after parse and before exec) that is enabled by default
// - shoudl we have a way to customize the template that gets created (i.e. to define other templates
//   or... )?  possibly, but without understanding
//   the use this is likely overkill for a first version - could easily add later.
// - the process of parsing and resolving templates against the fs can actually be totaly separate from the default
//   ServeHTTP functionlity

type TemplateMaker interface {
	NewTemplate() *template.Template
}

type TempateMakerFunc func() *template.Template

func (f *TemplateMakerFunc) NewTemplate() *template.Template {
	return f()
}

type Renderer interface {
	Parse(filename string) (*template.Template, error)
	Execute(t *template.Template, w io.Writer, ctx interface{}) error
	ParseAndExecute(filename string, w io.Writer, ctx interface{}) error
}

type renderer struct {
	// need something on here which can be called after parse and before execute (not sure if at end of parse
	// or beginning of execute or separately named step is better); but we need this in order to implement
	// UI requiring - and we need to have the UI requiring easily enabled;
	// also, whatever that step is should not be hidden from things that don't want to call Execute()
	// ParseAndExecute().
}

func New() Renderer {

}
