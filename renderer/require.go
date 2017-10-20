package renderer

import (
	"context"
	"fmt"
	"html/template"
	"strings"
)

type requirer interface {
	Require(name string) error
}

// NewRequireModifier returns a TemplateModifier that looks for a "uifiles.FileSet" on the context with a Require(string)error method.
// If it's found then each template definition that starts with "require " is converted into a Require() call.
// Provides clean integration with uifiles without directly depending on it.
func NewRequireModifier() TemplateModifier {
	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {

		fs := ctx.Value("uifiles.FileSet")
		if fs != nil {

			rq, ok := fs.(requirer)
			if !ok {
				return ctx, t, fmt.Errorf("context has a 'uifiles.FileSet' but it does not implement Require(string)error : %#v", fs)
			}

			allt := t.Templates()
			for _, onet := range allt {
				tmplname := onet.Name()
				if strings.HasPrefix(tmplname, "require ") {
					rqname := strings.TrimPrefix(tmplname, "require ")
					err := rq.Require(rqname)
					if err != nil {
						return ctx, t, fmt.Errorf("error attempting require of %q: %v", rqname, err)
					}
				}
			}

		}

		return ctx, t, nil
	})
}
