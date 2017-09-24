package caveman

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"text/template/parse"

	"github.com/russross/blackfriday"

	yaml "gopkg.in/yaml.v2"
)

// TODO: think the file cache problem through - do we just do the cache slug thing
// or should we try to get more sophisticated...  Hm - yeah we actually need to confront
// CSS and JS combining as well...  Need a manager thingy for this but it will look at the PageMeta

// support md and html(go templates) OK
// make top meta part pluggable ??
// make a big struct with default stuff in it but also way to do custom things OK (NEEDS FIELD BUILD-OUT)
// alternate templating systems possible OK
// hybrid systems possible too (md will be like this) OK (NEEDS BUILD-OUT)
// stripping off ".html" and ".md" is pluggable but default behavior OK
// using "index.html" or "index.md" as default for dir is also pluggable but default OK
// separate fs+root to include templates from OK
// default is to scan views and provide those as the page list but that should be customizable with simple "exclude this", "rewrite this", "add this" sort of stuff
// should support layering with mulitple underlying filesystems to read from (although this can probably be implemented with a layered implementation of http.FileSystem, rather than putting that code in here)

// NewDefaultRedirects returns a helper handler to normalize URLs by 301 redirecting
// common variations that would otherwise create duplicate URLs.
func NewDefaultRedirects() http.Handler {
	return http.HandlerFunc(DefaultRedirects)
}

func DefaultRedirects(w http.ResponseWriter, r *http.Request) {

	qs := ""
	if r.URL.RawQuery != "" {
		qs = "?" + r.URL.RawQuery
	}

	p := path.Clean("/" + r.URL.Path)

	// if it's not already cleaned, redirect to cleaned version
	if p != r.URL.Path {
		http.Redirect(w, r, p+qs, 301)
		return
	}

	if strings.HasSuffix(p, ".gohtml") {
		p := path.Clean("/" + strings.TrimSuffix(p, ".gohtml"))
		http.Redirect(w, r, p+qs, 301)
		return
	}

	if strings.HasSuffix(p, ".html") {
		p := path.Clean("/" + strings.TrimSuffix(p, ".html"))
		http.Redirect(w, r, p+qs, 301)
		return
	}

	if strings.HasSuffix(p, ".md") {
		p := path.Clean("/" + strings.TrimSuffix(p, ".md"))
		http.Redirect(w, r, p+qs, 301)
		return
	}

	if strings.HasSuffix(p, "/index") {
		p := path.Clean("/" + strings.TrimSuffix(p, "/index"))
		http.Redirect(w, r, p+qs, 301)
		return
	}

}

// NewDefaultRenderer creates a renderer with sensible defaults and supports Go templates and markdown.
func NewDefaultRenderer(bodyFs http.FileSystem, includeFs http.FileSystem) *Renderer {

	return &Renderer{
		BodyFs:    bodyFs,
		IncludeFs: includeFs,
		BodyTemplaters: map[string]BodyTemplater{
			".html":   &GoBodyTemplater{},
			".gohtml": &GoBodyTemplater{},
			".md":     NewMarkdownBodyTemplater(),
		},
		PathExpandFunc: DefaultPathExpand,
	}
}

func DefaultPathExpand(p string) []string {
	ext := path.Ext(p)
	if ext == "" {
		return []string{
			path.Clean("/" + p + ".gohtml"),
			path.Clean("/" + p + ".html"),
			path.Clean("/" + p + ".md"),
			path.Clean("/" + p + "/index.gohtml"),
			path.Clean("/" + p + "/index.html"),
			path.Clean("/" + p + "/index.md"),
		}
	}
	return []string{path.Clean("/" + p)}
}

type Renderer struct {
	BodyFs         http.FileSystem          // read body templates from here
	IncludeFs      http.FileSystem          // read included templates from here
	BodyTemplaters map[string]BodyTemplater // map of file extension to BodyTemplater
	PathExpandFunc func(string) []string    // translate a path into a list of other possible ones that should be checked (in sequence) and the first one found rendered
}

// ServeHTTP implements http.Handler
func (rr *Renderer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr.ServeHTTPPath(w, r, r.URL.Path)
}

type RenderHandler interface {
	ServeHTTPPath(w http.ResponseWriter, r *http.Request, p string)
}

// ServeHTTPPath implemented RenderHandler and is like ServeHTTP but lets you override which file it will look for.
// This is a simple way to "rewrite" URLs based on custom logic - you just call the
// Renderer with the appropriate custom path.
func (rr *Renderer) ServeHTTPPath(w http.ResponseWriter, r *http.Request, p string) {

	p = path.Clean("/" + p)

	ps := rr.PathExpandFunc(p)

	for _, p := range ps {

		ext := path.Ext(p)
		bt := rr.BodyTemplaters[ext]
		if bt == nil {
			continue
		}

		f, err := rr.BodyFs.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		pageMeta, body, err := bt.BodyTemplate(f)
		if err != nil {
			HTTPError(w, r, err, "error executing body templater", 500)
			break
		}
		// log.Printf("pageMeta: %#v", pageMeta)

		t, err := template.New("__main").Parse(body)
		if err != nil {
			HTTPError(w, r, err, "body template parse error", 500)
			break
		}

		err = TmplIncludeAll(rr.IncludeFs, t)
		if err != nil {
			HTTPError(w, r, err, "include template parse error", 500)
			break
		}

		ctx := r.Context()

		ctx = CtxWithPageMeta(ctx, pageMeta)

		log.Printf("FIXME: figure out FuncMap - sensible defaults with a way to add or override")

		// t.Funcs(template.FuncMap())
		// FIXME: this is placeholder
		t.Funcs(map[string]interface{}{
			"html": func(s string) template.HTML { return template.HTML(s) },
		})

		// log.Printf("page-meta: %#v", ctx.Value("page-meta"))

		// tctx := struct {
		// 	context.Context
		// 	Request *http.Request
		// }{
		// 	Context: ctx,
		// 	Request: r,
		// }

		// FIXME: do we need this? or is there already a request in there - looks like not, but think this through
		ctx = context.WithValue(ctx, "req", r)

		var wb bytes.Buffer

		// FIXME: I think we need to catch panic here...

		err = t.ExecuteTemplate(&wb, "__main", ctx)
		if err != nil {
			HTTPError(w, r, err, "error executing template", 500)
			break
		}

		// trim leading space
		outbytes := bytes.TrimLeft(wb.Bytes(), " \n\r\t")

		// make sure to set content type (the auto-sniffer thing does not work with gzip)
		w.Header().Set("content-type", "text/html")

		w.Write(outbytes)

		break

	}

}

func ParseMetaAndText(b []byte, meta interface{}) (tmpltext []byte, e error) {

	// check for yaml header
	if bytes.HasPrefix(b, []byte("+++")) {

		b = bytes.TrimPrefix(b, []byte("+++"))

		parts := bytes.SplitN(b, []byte("\n+++"), 2)

		// as a cute little hack here - we count the number of lines in the yaml header
		// part and put that many blank spaces ahead of the template, so on errors
		// the line numbers match up; and these spaces get stripped off in the page rendering code;
		// ugly but workable

		tmpltext = bytes.Join([][]byte{
			bytes.Repeat([]byte("\n"), len(bytes.Split(parts[0], []byte("\n")))),
			parts[1],
		}, nil)

		if meta != nil {

			err := yaml.Unmarshal(parts[0], meta)
			if err != nil {
				return nil, err
			}

		}

	} else {
		// no yaml header, makes it real simple
		tmpltext = b
	}

	return tmpltext, nil

}

// BodyTemplater reads the body of a page and returns a string with the appropriate Go template.
// Pages that are already Go templates can be returned as-is but this interface allows us to
// implement markdown or other languages by converting them to Go templates.
type BodyTemplater interface {
	BodyTemplate(io.Reader) (*PageMeta, string, error)
}

type GoBodyTemplater struct{}

func (t *GoBodyTemplater) BodyTemplate(r io.Reader) (*PageMeta, string, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	meta := &PageMeta{}
	text, err := ParseMetaAndText(b, meta)

	return meta, string(text), err
}

func NewMarkdownBodyTemplater() *MarkdownBodyTemplater {
	return &MarkdownBodyTemplater{
		TemplatePrefix: `{{template "page.html" .}}{{define "body"}}`,
		TemplateSuffix: `{{end}}`,
	}
}

type MarkdownBodyTemplater struct {
	TemplatePrefix string
	TemplateSuffix string
}

func (t *MarkdownBodyTemplater) BodyTemplate(r io.Reader) (*PageMeta, string, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	meta := &PageMeta{}
	text, err := ParseMetaAndText(b, meta)

	// process text md->html
	output := blackfriday.MarkdownCommon([]byte(text))

	return meta, t.TemplatePrefix + string(output) + t.TemplateSuffix, err

}

// TmplIncludeAll - see https://stackoverflow.com/questions/20716726/golang-text-html-template-call-other-templates-with-dynamic-name/42969242#42969242
func TmplIncludeAll(fs http.FileSystem, t *template.Template) error {

	tlist := t.Templates()
	for _, et := range tlist {
		if et != nil && et.Tree != nil && et.Tree.Root != nil {
			err := TmplIncludeNode(fs, et, et.Tree.Root)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TmplIncludeNode - see https://stackoverflow.com/questions/20716726/golang-text-html-template-call-other-templates-with-dynamic-name/42969242#42969242
func TmplIncludeNode(fs http.FileSystem, t *template.Template, node parse.Node) error {

	if node == nil {
		return nil
	}

	switch node := node.(type) {

	case *parse.TemplateNode:
		if node == nil {
			return nil
		}

		// if template is already defined, do nothing
		tlist := t.Templates()
		for _, et := range tlist {
			if node.Name == et.Name() {
				return nil
			}
		}

		t2 := t.New(node.Name)

		f, err := fs.Open(node.Name)
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		_, err = t2.Parse(string(b))
		if err != nil {
			return err
		}

		// start over again, will stop recursing when there are no more templates to include
		return TmplIncludeAll(fs, t)

	case *parse.ListNode:

		if node == nil {
			return nil
		}

		for _, node := range node.Nodes {
			err := TmplIncludeNode(fs, t, node)
			if err != nil {
				return err
			}
		}

	case *parse.IfNode:
		if err := TmplIncludeNode(fs, t, node.BranchNode.List); err != nil {
			return err
		}
		if err := TmplIncludeNode(fs, t, node.BranchNode.ElseList); err != nil {
			return err
		}

	case *parse.RangeNode:
		if err := TmplIncludeNode(fs, t, node.BranchNode.List); err != nil {
			return err
		}
		if err := TmplIncludeNode(fs, t, node.BranchNode.ElseList); err != nil {
			return err
		}

	case *parse.WithNode:
		if err := TmplIncludeNode(fs, t, node.BranchNode.List); err != nil {
			return err
		}
		if err := TmplIncludeNode(fs, t, node.BranchNode.ElseList); err != nil {
			return err
		}

	}

	return nil
}

// WalkRenderAndIndex is a helper to glue together the Renderer and PageIndex,
// since these two are intentionally kept unaware of each other.  This walks the
// file system and calls the Renderer for each file found, adding the approrpiate
// PageMeta to the PageIndex.
// FIXME: fileExts moves to some kinda of Options struct and we also need a way to regexp exclude,
// as well as specify removal of file extension.
func WalkRenderAndIndex(bodyFs http.FileSystem, fileExts []string, renderer Renderer, pageIndex *PageIndex) error {

	// bodyFs.Open(name)
	panic("not implemented")
	return nil

}
