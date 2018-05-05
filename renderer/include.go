package renderer

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"text/template/parse"
	"time"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/webutil"
)

func NewIncludeTemplateReaderModifier(tr TemplateReader, category string) TemplateModifier {
	if category == "" {
		category = IncludesCategory
	}
	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {

		includeFS := fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {

			body, _, _, err := tr.ReadTemplate(category, name)
			if err != nil {
				return nil, err
			}

			f := fsutil.NewHTTPBytesFile(name, time.Now(), body)

			return f, nil
		})

		err := tmplIncludeAll(includeFS, t)
		return ctx, t, err
	})
}

func NewIncludeFSModifier(includeFS http.FileSystem) TemplateModifier {
	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {
		err := tmplIncludeAll(includeFS, t)
		return ctx, t, err
	})
}

func tmplIncludeAll(fs http.FileSystem, t *template.Template) error {

	tlist := t.Templates()
	for _, et := range tlist {
		if et != nil && et.Tree != nil && et.Tree.Root != nil {
			err := tmplIncludeNode(fs, et, et.Tree.Root)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func tmplIncludeNode(fs http.FileSystem, t *template.Template, node parse.Node) error {

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

		f, err := fs.Open(node.Name)
		if err != nil {

			// special case for missing file, we just leave it and let it error later as an undefined template
			if os.IsNotExist(err) || err == webutil.ErrNotFound {
				return nil
			}

			return err
		}
		defer f.Close()

		t2 := t.New(node.Name)

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		_, err = t2.Parse(string(b))
		if err != nil {
			return err
		}

		// start over again, will stop recursing when there are no more templates to include
		return tmplIncludeAll(fs, t)

	case *parse.ListNode:

		if node == nil {
			return nil
		}

		for _, node := range node.Nodes {
			err := tmplIncludeNode(fs, t, node)
			if err != nil {
				return err
			}
		}

	case *parse.IfNode:
		if err := tmplIncludeNode(fs, t, node.BranchNode.List); err != nil {
			return err
		}
		if err := tmplIncludeNode(fs, t, node.BranchNode.ElseList); err != nil {
			return err
		}

	case *parse.RangeNode:
		if err := tmplIncludeNode(fs, t, node.BranchNode.List); err != nil {
			return err
		}
		if err := tmplIncludeNode(fs, t, node.BranchNode.ElseList); err != nil {
			return err
		}

	case *parse.WithNode:
		if err := tmplIncludeNode(fs, t, node.BranchNode.List); err != nil {
			return err
		}
		if err := tmplIncludeNode(fs, t, node.BranchNode.ElseList); err != nil {
			return err
		}

	}

	return nil
}
