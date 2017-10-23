package renderer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"sort"
	"strings"
)

// NewPlusModifier returns a modifier that finds all templates with a plus sign and defines new ones which are a concatenation of them.
// Useful e.g. to allow mulitple things to append script tags to the bottom of the page.
func NewPlusModifier() TemplateModifier {

	return TemplateModifierFunc(func(ctx context.Context, t *template.Template) (context.Context, *template.Template, error) {

		// build a map of all templates with plusses in them keyed off the key part (before the first plus)
		tmap := make(map[string][]string)

		allt := t.Templates()
		for _, onet := range allt {

			name := onet.Name()
			nameParts := strings.SplitN(name, "+", 2)
			if len(nameParts) != 2 {
				continue
			}

			tmap[nameParts[0]] = append(tmap[nameParts[0]], name)

		}

		// now that they are grouped, sort each set and make a template with the calls
		for k, v := range tmap {

			pname := k + "+"

			// // skip if what we are about to define is already defined - hm, what about {{block...}} needs more thought
			// if t.Lookup(pname) != nil {
			// 	continue
			// }

			sort.Strings(v)

			pt := t.New(pname)

			var buf bytes.Buffer
			fmt.Fprintf(&buf, `{{define "%s"}}`, pname)
			for _, callName := range v {
				if callName == pname {
					continue // if {{block "whatever+"}}{{end}} is being done, we must make sure the template doesn't call itself
				}
				fmt.Fprintf(&buf, `{{template "%s" .}}`, callName)
			}
			fmt.Fprintf(&buf, `{{end}}`)

			_, err := pt.Parse(buf.String())
			if err != nil {
				return ctx, t, err
			}

		}

		return ctx, t, nil
	})

}
