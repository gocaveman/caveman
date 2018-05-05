// The renderer package provides functionality to convert the contents of files into servable URLs using Go HTML templating.
// As well as provide the "view layer" for MVC applications (i.e. from a handler you can call the renderer and tell it to
// render a page for you.) Rendering is consist of the following parts:
//
// A Loader is responsible for taking a filename and returning a stream of bytes that can be interpreted by Go's html/template package.
// FileExtLoader allows for easily calling different loaders based on file extension.  GohtmlLoader reads in a file and strips off
// the YAML meta section (delimited by +++ before and after, and replacing it with a comment so the line numbers stay corret) and
// returns the rest.  The markdown package provides an implementation which reads markdown and converts to Go HTML template language.
//
// A Renderer is able to parse a filename (it has a Loader on it which is uses to load files) and return a Template instance
// (from html/template).  The TemplateModifier interface can be used to customize settings such as the FuncMap used and delimiters;
// there is one that gets called before parse and one after, the default behaviors are to set the default FuncMap before and
// after to look for template names that start with "require " and call the require method on uifiles.FileSet from the context.
//
// A Renderer can be called directly from an http.Handler in order to render a specific page and this is the recommended way
// for handlers which need to do work and then render a page to function.  You can either Parse() and then execute the template
// yourself or ParseAndExecute() provides a more convenient way for the common case.
//
// A FileNamer takes a path from a URL and returns a list of possible filenames.  This distinction between URL paths and filenames
// is important - a Renderer is only aware of exact filenames.  Whereas a RenderHandler (below) uses a FileNamer to convert from
// the path that came in on a request to the underlying filename and they are not the same.  Examples include "/" -> "/index.gohtml",
// "/somepage" -> "/somepage.gohtml".  See NewDefaultFileNamer() for details.
//
// The RenderHandler provides functionality to use a FileNamer and a Renderer to serve pages in the manner you'd expect - you
// request a file and the page is served back to the browser.
//
package renderer

import (
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gocaveman/caveman/webutil"
)

// RendererFileName is the context key name for the file name of the top level template being rendered.
const RendererFileName = "renderer.FileName"

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

type TemplateModifier interface {
	TemplateModify(ctx context.Context, t *template.Template) (context.Context, *template.Template, error)
}

type TemplateModifierFunc func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error)

func (f TemplateModifierFunc) TemplateModify(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {
	return f(ctx, t)
}

func NewDefaultFuncMapModifier() TemplateModifier {
	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {
		t = t.Funcs(DEFAULT_FUNCMAP)
		return ctx, t, nil
	})
}

type TemplateModifierList []TemplateModifier

func (l TemplateModifierList) TemplateModify(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {
	var err error
	for _, m := range l {
		ctx, t, err = m.TemplateModify(ctx, t)
		if err != nil {
			return ctx, t, err
		}
	}
	return ctx, t, nil
}

// Renderer is the interface for rendering templates.
type Renderer interface {
	Parse(ctx context.Context, filename string) (context.Context, *template.Template, error)
	ParseAndExecute(ctx context.Context, filename string, w io.Writer, tmplCtx interface{}) error
	ParseAndExecuteHTTP(w http.ResponseWriter, r *http.Request, filename string)
}

// RendererImpl implements Renderer.
type RendererImpl struct {
	Loader      Loader
	BeforeParse TemplateModifier
	AfterParse  TemplateModifier
}

// NewFromTemplateReader creates a Renderer from a TemplateReader.
func NewFromTemplateReader(tReader TemplateReader) *RendererImpl {

	return &RendererImpl{
		Loader: &GoTemplateReaderLoader{
			TemplateReader: tReader,
			Category:       ViewsCategory,
		},
		BeforeParse: NewDefaultFuncMapModifier(),
		AfterParse: TemplateModifierList{
			NewIncludeTemplateReaderModifier(tReader, IncludesCategory),
			NewRequireModifier(),
			NewPlusModifier(),
			NewTemplateMetaModifier(tReader, ""),
		},
	}

}

// DEPRECATED: Use NewFromTemplateReader() instead.
// New returns an Renderer implementation with the default config.
func NewFromFSs(fileFS http.FileSystem, includeFS http.FileSystem) *RendererImpl {
	goLoader := NewGohtmlLoader(fileFS)
	return &RendererImpl{
		Loader: NewFileExtLoader(nil).
			WithExt(".gohtml", goLoader).
			WithExt(".html", goLoader),
		BeforeParse: NewDefaultFuncMapModifier(),
		AfterParse: TemplateModifierList{
			NewIncludeFSModifier(includeFS),
			NewRequireModifier(),
			NewPlusModifier(),
		},
	}
}

// Parse will load and parse the specified file, performing all before and after actions.
// The file name you pass is associated with the context as "renderer.FileName" (RendererFileName).
func (r *RendererImpl) Parse(ctx context.Context, filename string) (context.Context, *template.Template, error) {

	var err error
	t := template.New(filename)

	// make the template file name available to the modifiers and whatever else
	ctx = context.WithValue(ctx, RendererFileName, filename)

	if r.BeforeParse != nil {
		ctx, t, err = r.BeforeParse.TemplateModify(ctx, t)
		if err != nil {
			return ctx, t, err
		}
	}

	rc, err := r.Loader.Load(filename)
	if err != nil {
		return ctx, t, err
	}
	defer rc.Close()

	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return ctx, t, err
	}

	t, err = t.Parse(string(b))
	if err != nil {
		return ctx, t, err
	}

	if r.AfterParse != nil {
		ctx, t, err = r.AfterParse.TemplateModify(ctx, t)
		if err != nil {
			return ctx, t, err
		}
	}

	return ctx, t, nil
}

// ParseAndExecute will call Parse() and then execute the template.  tmplCtx is "data" argument to t.ExecuteTemplate,
// if you pass nil for it then the resulting context after Parse() will be used (recommended).
func (r *RendererImpl) ParseAndExecute(ctx context.Context, filename string, w io.Writer, tmplCtx interface{}) error {

	ctx, t, err := r.Parse(ctx, filename)
	if err != nil {
		return err
	}

	if tmplCtx == nil {
		tmplCtx = ctx
	}

	return t.ExecuteTemplate(w, filename, tmplCtx)

}

// ParseAndExecuteHTTP is similar to ParseAndExecute but works with ResponseWriter and Request.
// If unset, it will set the content-type header to text/html and cache-control to no-store.
// Errors are reported with webutil.HTTPError instead of being returned.
func (ri *RendererImpl) ParseAndExecuteHTTP(w http.ResponseWriter, r *http.Request, filename string) {

	if w.Header().Get("content-type") == "" {
		w.Header().Set("content-type", "text/html")
	}

	if w.Header().Get("cache-control") == "" {
		w.Header().Set("cache-control", "no-store")
	}

	ctx := r.Context()

	err := ri.ParseAndExecute(ctx, filename, w, nil)
	if err != nil {
		webutil.HTTPError(w, r, err, "render error", 500)
		return
	}

}
