package gen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// basic crud

// advanced find methods for specific fields (do we do this by
// calling store.FindByFirstName or store.Find("first_name", ...))
// should do something with LIKE vs exact field match

// paging

// basic detail view page load - so when we generate the view it doesn't
// need a separate handler to loads its data - the common case of a single
// record by ID is already handled here

// update mulitple records with one PATCH/PUT call, so you can e.g. update a bunch of sequence numbers at once

// i18n for validation (and other?) error messages

// callback methods, with interface and New method checks to see if we implement it -
// looks like this is changing to having a ...Router and a ...Controller and a Default...Controller

// handlerregistry integration

// optional (but default) permissions code (crud perms at top, registered, and checks in code)

// prefix for api calls

// prefix/path for page data (will need the renderer stuff for this to work as expected)
// what about login check for other pages (listing page), even though they don't necessarily
// require data loading...

func init() {
	globalMapGenerator["ctrl-crud"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		storeType := fset.String("store", "", "The type of the store to use for data access (defaults to '*store.Store' or '*Store', depending on package).")
		modelName := fset.String("model", "", "The model object name, if not specified default will be deduced from file name.")
		// TODO: responder code should be an option - it's a fair amount of cruft and people shouldn't
		// be forced to have it if they don't need it; default to off
		// TODO: option for permission stuff - default to on
		renderer := fset.Bool("renderer", true, "Output renderer integration for view/edit page.")
		tests := fset.Bool("tests", true, "Create test file with test(s) for this controller.")
		targetFile, data, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		_, targetFileName := filepath.Split(targetFile)

		if *modelName == "" {
			*modelName = NameSnakeToCamel(targetFileName, []string{"ctrl-"}, nil)
		}
		data["ModelName"] = *modelName

		data["ModelPathPart"] = strings.TrimPrefix(strings.TrimSuffix(targetFileName, ".go"), "ctrl-")

		if *storeType == "" {
			if data["PackageName"].(string) == "main" {
				*storeType = "*Store"
			} else {
				*storeType = "*store.Store"
			}
		}
		data["StoreType"] = *storeType

		data["Renderer"] = *renderer
		data["Tests"] = *tests

		err = OutputGoSrcTemplate(s, data, targetFile, `
package {{.PackageName}}

import (
	"net/http"
)

func init() {

}

type {{.ModelName}}Ctrl struct {
	Store {{.StoreType}} {{bq "autowire:\"\""}}
	APIPrefix string {{bq "autowire:\"api_prefix,optional\""}} // default: "/api" TODO: naming convention
	ModelPrefix string // default: "/{{.ModelPathPart}}"

	// TODO: renderer and path for data loading on detail page
}

func (h *{{.ModelName}}Ctrl) AfterWire() error {
	if h.APIPrefix == "" {
		h.APIPrefix = "/api"
	}
	if h.ModelPrefix == "" {
		h.ModelPrefix = "/{{.ModelPathPart}}"
	}
	return nil
}

func (h *{{.ModelName}}Ctrl) ServeHTTP(w http.ResponseWriter, r *http.Request) {



}

`, false)

		if err != nil {
			return err
		}

		if *tests {

			testsTargetFile := strings.Replace(targetFile, ".go", "_test.go", 1)
			if testsTargetFile == targetFile {
				return fmt.Errorf("unable to determine test file name for %q", targetFile)
			}

			err = OutputGoSrcTemplate(s, data, testsTargetFile, `
package {{.PackageName}}

func Test{{.ModelName}}CRUD(t *testing.T) {

	assert := assert.New(t)
	_ = assert

}

`, false)
			if err != nil {
				return err
			}

		}

		return nil

	})
}
