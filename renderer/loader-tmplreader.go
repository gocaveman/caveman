package renderer

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"os"
)

// TmplMeta is the context key for template metadata.
const TmplMeta = "tmpl.Meta"

// Default views category name
const ViewsCategory = "views"

// Default includes category name
const IncludesCategory = "includes"

// TemplateReader corresponds to tmpl.TemplateReader.
type TemplateReader interface {
	ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error)
}

// TODO: we just read the template directly - if we want to support markdown, there
// needs to be a version of this that reads the md and converts it to a Go template

// GoTemplateReaderLoader will load Go template HTML files (.gohtml or.html) from a TemplateReader.
type GoTemplateReaderLoader struct {
	TemplateReader TemplateReader
	Category       string
}

func (l *GoTemplateReaderLoader) Load(fileName string) (io.ReadCloser, error) {

	cat := l.Category
	if cat == "" {
		cat = ViewsCategory
	}

	body, mimeType, meta, err := l.TemplateReader.ReadTemplate(cat, fileName)
	if err != nil {
		return nil, err
	}

	_, _ = mimeType, meta

	// TODO: should we check the mime type to make sure it's actually a go html template?

	return ioutil.NopCloser(bytes.NewReader(body)), nil
}

// NewTemplateMetaModifier returns a modifier that reads the template Meta and puts
// it on the context as
func NewTemplateMetaModifier(tReader TemplateReader, category string) TemplateModifier {

	if category == "" {
		category = ViewsCategory
	}

	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {

		// this is set in RendererImpl.Parse() and tell us what file name is being rendered
		fileName, ok := ctx.Value(RendererFileName).(string)
		if !ok {
			return ctx, t, nil
		}

		_, _, meta, err := tReader.ReadTemplate(category, fileName)
		if err != nil {
			if os.IsNotExist(err) { // no such template, then nop
				return ctx, t, nil
			}
			// otherwise we return the error
			return ctx, t, err
		}

		ctx = context.WithValue(ctx, TmplMeta, meta)

		return ctx, t, nil

	})
}
