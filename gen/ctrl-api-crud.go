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
// See httpapi and fill stuff to properly implement PATCH

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
	globalMapGenerator["ctrl-api-crud"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

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
			*modelName = NameSnakeToCamel(targetFileName, []string{"ctrl-", "-api"}, nil)
		}
		data["ModelName"] = *modelName
		// FIXME: this breaks on JSONThing -> jSONThing
		data["ModelNameL"] = strings.ToLower((*modelName)[:1]) + (*modelName)[1:]

		data["ModelPathPart"] = strings.TrimPrefix(strings.TrimSuffix(targetFileName, ".go"), "ctrl-")

		if *storeType == "" {
			if data["PackageName"].(string) == "main" {
				*storeType = "*Store"
				data["ModelTypeName"] = *modelName
			} else {
				*storeType = "*store.Store"
				data["ModelTypeName"] = "store." + *modelName
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

const (
	{{.ModelName}}CreatePerm = "{{.ModelName}}.Create"
	{{.ModelName}}GetByIDPerm = "{{.ModelName}}.GetByID"
	{{.ModelName}}UpdatePerm   = "{{.ModelName}}.Update"
	{{.ModelName}}DeletePerm = "{{.ModelName}}.Delete"
)

type {{.ModelName}}APIRouter struct {
	APIPrefix string {{bq "autowire:\"api_prefix,optional\""}} // default: "/api" {{/* TODO: naming convention? */}}
	ModelPrefix string // default: "/{{.ModelPathPart}}"
	Controller *{{.ModelName}}APIController

	{{/*
	TODO: Renderer and path for data loading on detail page.
	Hm, need to figure out if the data loading for detail page,
	etc should be done here or in a different handler.
	*/}}
}

type {{.ModelName}}APIController struct {
	Store {{.StoreType}} {{bq "autowire:\"\""}}
}

func (h *{{.ModelName}}APIRouter) AfterWire() error {
	if h.APIPrefix == "" {
		h.APIPrefix = "/api"
	}
	if h.ModelPrefix == "" {
		h.ModelPrefix = "/{{.ModelPathPart}}"
	}
	return nil
}

func (h *{{.ModelName}}APIRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var {{.ModelNameL}} {{.ModelTypeName}}
	var {{.ModelNameL}}ID string

	ar := httpapi.NewRequest(r)

	var err error

	switch {

	/**
	 * @api {post} /api/{{.ModelPathPart}}
	 * @apiGroup {{.ModelName}}
	 */
	case ar.ParseRESTObj("POST", &{{.ModelNameL}}, h.APIPrefix+h.ModelPrefix):
		err = h.Controller.Create(w, r, ar, &{{.ModelNameL}})

	/**
	 * @api {get} /api/{{.ModelPathPart}}
	 * @apiGroup {{.ModelName}}
	 */
	case ar.ParseRESTPath("GET", h.APIPrefix+h.ModelPrefix):
		err = h.Controller.GetList(w, r, ar)

	/**
	 * @api {get} /api/{{.ModelPathPart}}/:id
	 * @apiGroup {{.ModelName}}
	 */
	case ar.ParseRESTPath("GET", h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID):
		err = h.Controller.GetByID(w, r, ar, {{.ModelNameL}}ID)

	/**
	 * @api {patch} /api/{{.ModelPathPart}}/:id
	 * @apiGroup {{.ModelName}}
	 */
	case ar.ParseRESTObjPath("PUT", &{{.ModelNameL}}, h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID) ||
		ar.ParseRESTObjPath("PATCH", &{{.ModelNameL}}, h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID):
		{{.ModelNameL}}.{{.ModelName}}ID = {{.ModelNameL}}ID
		err = h.Controller.Update(w, r, ar, &{{.ModelNameL}})

	/**
	 * @api {delete} /api/{{.ModelPathPart}}/:id
	 * @apiGroup {{.ModelName}}
	 */
	case ar.ParseRESTPath("DELETE", h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID):
		err = h.Controller.Delete(w, r, ar, {{.ModelNameL}}ID)

	}

	if err != nil {
		ar.WriteErr(w, err)
		return
	}	

}

func (h *{{.ModelName}}APIController) Create(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, {{.ModelNameL}} *{{.ModelTypeName}}) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}CreatePerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	err = h.Store.Create{{.ModelName}}({{.ModelNameL}})
	if err != nil {
		return err
	}

	w.Header().Set("Location", r.URL.Path + "/"+{{.ModelNameL}}.{{.ModelName}}ID)
	w.Header().Set("X-Id", {{.ModelNameL}}.{{.ModelName}}ID)
	ar.WriteResult(w, 201, {{.ModelNameL}})

	return nil
}

func (h *{{.ModelName}}APIController) GetList(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}GetListPerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	{{.ModelNameL}}List, err := h.Store.Find{{.ModelName}}List()
	if err != nil {
		return err
	}

	ar.WriteResult(w, 200, {{.ModelNameL}}List)

	return nil
}

func (h *{{.ModelName}}APIController) GetByID(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, {{.ModelNameL}}ID string) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}GetByIDPerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	{{.ModelNameL}}, err := h.Store.Find{{.ModelName}}ByID({{.ModelNameL}}ID)
	{{/* FIXME: need better handling of not found case */}}
	if err != nil {
		return err
	}

	ar.WriteResult(w, 200, {{.ModelNameL}})

	return nil
}

func (h *{{.ModelName}}APIController) Update(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, {{.ModelNameL}} *{{.ModelTypeName}}) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}UpdatePerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	err = h.Store.Update{{.ModelName}}({{.ModelNameL}})
	if err != nil {
		return err
	}

	ar.WriteResult(w, 200, {{.ModelNameL}})

	return nil
}

func (h *{{.ModelName}}APIController) Delete(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, {{.ModelNameL}}ID string) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}DeletePerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	err = h.OSGStore.Delete{{.ModelName}}({{.ModelNameL}}ID)
	if err != nil {
		return err
	}

	ar.WriteResult(w, 200, true)

	return nil
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

// TODO: write tests

{{/*
func Test{{.ModelName}}CRUD(t *testing.T) {

	assert := assert.New(t)
	_ = assert

}
*/}}

`, false)
			if err != nil {
				return err
			}

		}

		return nil

	})
}
