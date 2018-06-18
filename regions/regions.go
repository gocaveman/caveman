package regions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/gocaveman/caveman/webutil"
)

var ErrWriteNotSupported = errors.New("write not supported")

// BlockDefiner lets us define blocks.  Matches renderer.BlockDefiner
type BlockDefiner interface {
	BlockDefine(name, content string, condition func(ctx context.Context) (bool, error))
}

// RegionHandler analyzes the regions against the current request and uses BlockDefiner (see renderer package) to create the appropriate blocks.
type RegionHandler struct {
	Store Store `autowire:""`
}

func (h *RegionHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	if h.Store == nil {
		panic("RegionHandler.Store is nil")
	}

	ctx := r.Context()
	blockDefiner, ok := ctx.Value("renderer.BlockDefiner").(BlockDefiner)
	if !ok {
		// TODO: not sure what the rule is on this - should we fail, just log, or what;
		// log for now...
		log.Printf("RegionHandler: `renderer.BlockDefiner` is not set on context, cannot define blocks.  Are you missing the renderer.BlockDefineHandler in your handler list?")
		return w, r
	}

	defs, err := h.Store.AllDefinitions()
	if err != nil {
		webutil.HTTPError(w, r, err, "error getting region definitions", 500)
		return w, r
	}

	// should be handled now...
	// log.Printf("FIXME: we really should be doing this check at a time when both the template and pageinfo meta are available, so those can be used in the condition... which would mean maybe BlockDefiner needs to allow for a condition or something...")

	regionNames := defs.RegionNames()
	for _, regionName := range regionNames {
		// FIXME: what if regionName is empty or doesn't start with plus - need to think
		// the error conditions through and do a better job.
		regionDefs := defs.ForRegionName(regionName).SortedCopy()
		for i := range regionDefs {

			def := regionDefs[i]

			if def.Disabled {
				continue
			}

			// log.Printf("FIXME: so this can't work like this we need to be able to delay the check using the BlockDefiner condition stuff, which means context-only")
			// if !def.MatchesRequest(r) {
			// 	continue
			// }

			// name formatted so plusser adds to correct region, sorts in sequence and is unique
			name := fmt.Sprintf("%s0%04d_%s", regionName, i, def.DefinitionID)

			// use default context or if ContextMeta then include call for that
			ctxStr := `$`
			if len(def.ContextMeta) > 0 {
				b, err := json.Marshal(def.ContextMeta)
				if err != nil {
					webutil.HTTPError(w, r, err, fmt.Sprintf("error marshaling JSON data for region defintion %q", def.DefinitionID), 500)
					return w, r
				}
				ctxStr = fmt.Sprintf(`(WithValueMapJSON $ %q)`, b)
			}

			// condition check is delayed until later when everything possible is available on the context
			// FIXME: condition should have error return
			condition := func(ctx context.Context) (bool, error) {
				return def.MatchesContext(ctx)
			}

			content := fmt.Sprintf(`{{template %q (WithValue %s "regions.DefinitionID" %q)}}`, def.TemplateName, ctxStr, def.DefinitionID)
			// log.Printf("Doing BlockDefine(%q,...)\n%s", name, content)
			blockDefiner.BlockDefine(name, content, condition)

		}
	}

	return w, r
}

type Store interface {
	WriteDefinition(d Definition) error
	DeleteDefinition(defintionID string) error
	AllDefinitions() (DefinitionList, error)
}

// Definition describes a single entry in a region.
type Definition struct {
	DefinitionID string  `json:"definition_id" yaml:"definition_id" db:"definition_id"` // must be globally unique, recommend prefixing with package name for built-in ones
	RegionName   string  `json:"region_name" yaml:"region_name" db:"region_name"`       // name of the region this gets appended to, must end with a "+" to match template region names (see renderer/plusser.go)
	Sequence     float64 `json:"sequence" yaml:"sequence" db:"sequence"`                // order of this in the region, by convention this is between 0 and 100 but can be any number supported by float64
	Disabled     bool    `json:"disabled" yaml:"disabled" db:"disabled"`                // mainly here so we can override a definition as disabled
	TemplateName string  `json:"template_name" yaml:"template_name" db:"template_name"` // name/path of template to include

	CondIncludePaths string `json:"cond_include_paths" yaml:"cond_include_paths" db:"cond_include_paths"` // http request paths to include (wildcard pattern if starts with "/" or regexp, multiple are whitespace separated)
	CondExcludePaths string `json:"cond_exclude_paths" yaml:"cond_exclude_paths" db:"cond_exclude_paths"` // http request paths to exclude (wildcard pattern if starts with "/" or regexp, multiple are whitespace separated)
	CondTemplate     string `json:"cond_template" yaml:"cond_template" db:"cond_template"`                // custom condition as text/template, if first non-whitespace character is other than '0' means true, otherwise false, take precedence over other conditions

	ContextMeta webutil.SimpleStringDataMap `json:"context_meta" yaml:"context_meta" db:"context_meta"` // static data that gets set on the context during the call
}

func (d *Definition) IsValid() error {

	if d.DefinitionID == "" {
		return fmt.Errorf("DefinitionID is empty")
	}
	if d.RegionName == "" {
		return fmt.Errorf("RegionName is empty")
	}
	if math.IsNaN(d.Sequence) || math.IsInf(d.Sequence, 0) {
		return fmt.Errorf("Sequence number is not valid")
	}
	if d.TemplateName == "" {
		return fmt.Errorf("TemplateName is empty")
	}
	if _, err := condPaths(d.CondIncludePaths).checkPath("/"); err != nil {
		return fmt.Errorf("CondIncludePaths error: %v", err)
	}
	if _, err := condPaths(d.CondExcludePaths).checkPath("/"); err != nil {
		return fmt.Errorf("CondExcludePaths error: %v", err)
	}

	// FIXME: need to extract out the thing that deals with CondTemplate and at least
	// parse it here

	return nil
}

type condPaths string

func (ps condPaths) checkPath(p string) (bool, error) {

	p = path.Clean("/" + p)

	parts := strings.Fields(string(ps))

	found := false

	for _, part := range parts {
		if part == "" {
			continue
		}

		// wildcard
		if part[0] == '/' {

			expr := `^` + strings.Replace(regexp.QuoteMeta(part), `\*`, `.*`, -1) + `$`
			re, err := regexp.Compile(expr)
			if err != nil {
				return false, fmt.Errorf("checkPath wildcard regexp compile error (`%s`): %v", expr, err)
			}

			if re.MatchString(p) {
				found = true
				break
			}

		} else { // otherwise regexp

			re, err := regexp.Compile(part)
			if err != nil {
				return false, fmt.Errorf("checkPath direct regexp compile error (`%s`): %v", part, err)
			}

			if re.MatchString(p) {
				found = true
				break
			}

		}

	}

	return found, nil

}

// MatchesContext checks the condition against the HTTP request in a context.
func (d *Definition) MatchesContext(ctx context.Context) (bool, error) {

	r, ok := ctx.Value("http.Request").(*http.Request)
	if !ok {
		// log.Printf("NO REQUEST")
		return false, nil
	}

	// check CondTemplate only if set
	if d.CondTemplate != "" {

		t, err := template.New("_").Parse(d.CondTemplate)
		if err != nil {
			return false, fmt.Errorf("MatchesContext error while parsing condition template: %v\nTEMPLATE: %s", err, d.CondTemplate)
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, ctx)
		if err != nil {
			return false, fmt.Errorf("MatchesContext error while executing condition template: %v\nTEMPLATE: %s", err, d.CondTemplate)
		}

		bufStr := strings.TrimSpace(buf.String())
		return len(bufStr) > 0 && bufStr[0] != '0', nil

	}

	var err error

	// do include/exclude thing
	ret := true
	if strings.TrimSpace(d.CondIncludePaths) != "" {
		// if any include condition then require one include to match
		ret, err = condPaths(d.CondIncludePaths).checkPath(r.URL.Path)
		if err != nil {
			return false, err
		}
	}
	// if exclude then veto prior state
	v, err := condPaths(d.CondExcludePaths).checkPath(r.URL.Path)
	if err != nil {
		return false, err
	}
	if v {
		ret = false
	}

	return ret, nil

}

type DefinitionList []Definition

func (p DefinitionList) Len() int           { return len(p) }
func (p DefinitionList) Less(i, j int) bool { return p[i].Sequence < p[j].Sequence }
func (p DefinitionList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// RegionNames returns a deduplicated slice of the RegionName fields of this list.
func (dl DefinitionList) RegionNames() []string {
	ret := make([]string, 0, 32)
	retMap := make(map[string]bool, 32)
	for _, d := range dl {
		if !retMap[d.RegionName] {
			retMap[d.RegionName] = true
			ret = append(ret, d.RegionName)
		}
	}
	sort.Strings(ret)
	return ret
}

// ForRegionName returns the definitions that match a region name.
func (dl DefinitionList) ForRegionName(regionName string) DefinitionList {
	var ret DefinitionList
	for _, d := range dl {
		if d.RegionName == regionName {
			ret = append(ret, d)
		}
	}
	return ret
}

// SortedCopy returns a copy that is sorted by Sequence.
func (dl DefinitionList) SortedCopy() DefinitionList {
	ret := make(DefinitionList, len(dl))
	copy(ret, dl)
	sort.Sort(ret)
	return ret
}

// ListStore implements Store on top of a slice of Defintions.  Read-only.
// Makes it simple to implement a static Store in a theme or other Go code.
type ListStore []Definition

func (s ListStore) WriteDefinition(d Definition) error {
	return ErrWriteNotSupported
}
func (s ListStore) DeleteDefinition(defintionID string) error {
	return ErrWriteNotSupported
}
func (s ListStore) AllDefinitions() (DefinitionList, error) {
	return DefinitionList(s), nil
}

// StackedStore implements Store on top of a slice of other Stores.
// WriteDefinition and DeleteDefinition call each member of the slice in sequence
// until one returns nil or an error other than ErrWriteNotSupported.
// AllDefinitions returns a combined list of defintions from all stores
// with the Stores earlier in the list taking precedence, e.g. a Definition
// with a specific DefinitionID at Store index 0 will be returns instead of
// a Definition with the same ID at Store index 1.
type StackedStore []Store

func (ss StackedStore) WriteDefinition(d Definition) error {
	for _, s := range ss {
		err := s.WriteDefinition(d)
		if err == ErrWriteNotSupported {
			continue
		}
		return err
	}
	return ErrWriteNotSupported
}

func (ss StackedStore) DeleteDefinition(defintionID string) error {
	for _, s := range ss {
		err := s.DeleteDefinition(defintionID)
		if err == ErrWriteNotSupported {
			continue
		}
		return err
	}
	return ErrWriteNotSupported
}

func (ss StackedStore) AllDefinitions() (DefinitionList, error) {
	var ret DefinitionList
	retMap := make(map[string]bool)
	for _, s := range ss {
		defs, err := s.AllDefinitions()
		if err != nil {
			return nil, err
		}
		for _, def := range defs {
			if !retMap[def.DefinitionID] {
				ret = append(ret, def)
				retMap[def.DefinitionID] = true
			}
		}
	}
	return ret, nil
}
