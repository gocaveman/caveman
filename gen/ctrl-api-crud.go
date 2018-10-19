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
			*modelName = NameSnakeToCamel(targetFileName, []string{"ctrl-"}, []string{"-api.go", ".go"})
		}
		data["ModelName"] = *modelName
		// FIXME: this breaks on JSONThing -> jSONThing
		data["ModelNameL"] = strings.ToLower((*modelName)[:1]) + (*modelName)[1:]

		data["ModelPathPart"] = strings.TrimPrefix(
			strings.TrimSuffix(strings.TrimSuffix(targetFileName, ".go"), "-api"),
			"ctrl-")

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
	{{.ModelName}}FetchPerm = "{{.ModelName}}.Fetch"
	{{.ModelName}}SearchPerm = "{{.ModelName}}.Search"
	{{.ModelName}}UpdatePerm = "{{.ModelName}}.Update"
	{{.ModelName}}DeletePerm = "{{.ModelName}}.Delete"

	{{/* TODO: figure out if we should provide some variations as options
	{{.ModelName}}CreateAnyPerm = "{{.ModelName}}.CreateAny"
	{{.ModelName}}FetchAnyPerm = "{{.ModelName}}.FetchAny"
	{{.ModelName}}SearchAnyPerm = "{{.ModelName}}.SearchAny"
	{{.ModelName}}UpdateAnyPerm = "{{.ModelName}}.UpdateAny"
	{{.ModelName}}DeleteAnyPerm = "{{.ModelName}}.DeleteAny"
	*/}}
)

func init() {
	permregistry.MustAddPerm("admin", {{.ModelName}}CreatePerm)
	permregistry.MustAddPerm("admin", {{.ModelName}}FetchPerm)
	permregistry.MustAddPerm("admin", {{.ModelName}}SearchPerm)
	permregistry.MustAddPerm("admin", {{.ModelName}}UpdatePerm)
	permregistry.MustAddPerm("admin", {{.ModelName}}DeletePerm)
}

type {{.ModelName}}APIRouter struct {
	APIPrefix string {{bq "autowire:\"api_prefix,optional\""}} // default: "/api" {{/* TODO: naming convention? */}}
	ModelPrefix string // default: "/{{.ModelPathPart}}"
	Controller *{{.ModelName}}APIController

	{{/*
	TODO: Renderer and path for data loading on detail page.
	Hm, need to figure out if the data loading for detail page,
	etc should be done here or in a different handler.  Or do we
	just assume people won't need to do this... not sure, different
	schools of thought on that
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

// search{{.ModelName}}Params is the criteria for a search,
// corresponding to URL parameters
type search{{.ModelName}}Params struct {
	Criteria tmetautil.Criteria {{bq "json:\"criteria\""}}
	OrderBy tmetautil.OrderByList {{bq "json:\"order_by\""}}
	Limit int64 {{bq "json:\"limit\""}}
	Offset int64 {{bq "json:\"offset\""}}
	Related []string {{bq "json:\"related\""}}
	ReturnCount bool {{bq "json:\"return_count\""}}
}

func (h *{{.ModelName}}APIRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var {{.ModelNameL}} {{.ModelTypeName}}
	var {{.ModelNameL}}ID string
	var mapData map[string]interface{}

	ar := httpapi.NewRequest(r)

	searchParams := search{{.ModelName}}Params {
		Limit: 100,
	}
	var err error

	switch {

	/**
	 * @api {post} /api/{{.ModelPathPart}} Create {{.ModelName}}
	 * @apiGroup {{.ModelName}}
	 * @apiName create-{{.ModelPathPart}}
	 * @apiDescription Create a {{.ModelName}}
	 *
	 * @apiSuccessExample {json} Success-Response:
	 *     HTTP/1.1 200 OK
	 *     Content-Type: application/json
	 *     Location: /api/{{.ModelPathPart}}/6cSAs4i2P3PsHq2s6PZi6V
	 *     X-Id: 6cSAs4i2P3PsHq2s6PZi6V
	 *
	 *     [{
	 *         // {{.ModelName}}
	 *     }]
	 */
	case ar.ParseRESTObj("POST", &{{.ModelNameL}}, h.APIPrefix+h.ModelPrefix):
		err = h.Controller.Create(w, r, ar, &{{.ModelNameL}})

	/**
	 * @api {get} /api/{{.ModelPathPart}} Search {{plural .ModelName}}
	 * @apiGroup {{.ModelName}}
	 * @apiName search-{{.ModelPathPart}}
	 * @apiDescription List {{plural .ModelName}}
	 *
	 * @apiSuccessExample {json} Success-Response:
	 *     HTTP/1.1 200 OK
	 *     Content-Type: application/json
	 *
	 *     [{
	 *         // {{.ModelName}}
	 *     }]
	 */
	case ar.ParseRESTPath("GET", h.APIPrefix+h.ModelPrefix):
		err = httpapi.FormUnmarshal(r.URL.Query(), &searchParams)
		if err != nil {
			break
		}
		err = h.Controller.Search(w, r, ar, searchParams)

	/**
	 * @api {get} /api/{{.ModelPathPart}}/:id Fetch {{.ModelName}}
	 * @apiGroup {{.ModelName}}
	 * @apiName fetch-{{.ModelPathPart}}
	 * @apiDescription Get a {{.ModelName}} with the specified ID.  Will return an error
	 * if the object cannot be found.
	 *
	 * @apiParam {String} id ID of the {{.ModelName}} to return
	 *
	 * @apiSuccessExample {json} Success-Response:
	 *     HTTP/1.1 200 OK
	 *     Content-Type: application/json
	 *
	 *     [{
	 *         // {{.ModelName}}
	 *     }]
	 */
	case ar.ParseRESTPath("GET", h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID):
		err = httpapi.FormUnmarshal(r.URL.Query(), &searchParams)
		if err != nil {
			break
		}
		err = h.Controller.Fetch(w, r, ar, {{.ModelNameL}}ID, searchParams.Related...)

	/**
	 * @api {patch} /api/{{.ModelPathPart}}/:id Update {{.ModelName}}
	 * @apiGroup {{.ModelName}}
	 * @apiName update-{{.ModelPathPart}}
	 * @apiDescription Update a {{.ModelName}} by ID.  Will return an error
	 * if the object cannot be found or validation error occurs.  The updated
	 * object will be returned.
	 *
	 * @apiParam {String} id ID of the {{.ModelName}} to update
	 *
	 * @apiSuccessExample {json} Success-Response:
	 *     HTTP/1.1 200 OK
	 *     Content-Type: application/json
	 *
	 *     [{
	 *         // {{.ModelName}}
	 *     }]
	 */
	case ar.ParseRESTObjPath("PUT", &mapData, h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID) ||
		ar.ParseRESTObjPath("PATCH", &mapData, h.APIPrefix+h.ModelPrefix+"/%s", &{{.ModelNameL}}ID):
		{{.ModelNameL}}.{{.ModelName}}ID = {{.ModelNameL}}ID
		err = h.Controller.Update(w, r, ar, {{.ModelNameL}}ID, mapData)

	/**
	 * @api {delete} /api/{{.ModelPathPart}}/:id Delete {{.ModelName}}
	 * @apiGroup {{.ModelName}}
	 * @apiName delete-{{.ModelPathPart}}
	 * @apiDescription Delete a {{.ModelName}} by ID.
	 *
	 * @apiParam {String} id ID of the {{.ModelName}} to delete
	 *
	 * @apiSuccessExample {json} Success-Response:
	 *     HTTP/1.1 200 OK
	 *     Content-Type: application/json
	 *
	 *     {"result":true}
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

	// FIXME: we should probably be using httpapi.Fill() here to ensure
	// consistent behavior with Update()

	err := h.Store.Create{{.ModelName}}(r.Context(), {{.ModelNameL}})
	if err != nil {
		return err
	}

	w.Header().Set("Location", r.URL.Path + "/"+{{.ModelNameL}}.{{.ModelName}}ID)
	w.Header().Set("X-Id", {{.ModelNameL}}.{{.ModelName}}ID)
	ar.WriteResult(w, 201, {{.ModelNameL}})

	return nil
}

// TODO: we need our fancy find, it should have:
// * List of relations to return
// * Make sure to set the deadline, and a default and a max limit on the records, offset for paging
// * Where conditions expressed as nested struct (probably something that should go in dbutil pkg).
//   There can also be something on that where struct that allows the Store to easily say
//   "I need an exact match (optionally or LIKE prefix% with at least N chars) on
//   at least one of these fields", to enforce that an index is being
//   used (if desired - which it should be by default).
// * What about rate limiting for security purposes, and API usage? (related but not necessarily the same)

func (h *{{.ModelName}}APIController) Search(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, params search{{.ModelName}}Params) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}SearchPerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	// make a separate context with timeout so search doesn't run too long
	queryCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	maxLimit := int64(5000) // never allow more than this many records
	if params.Limit + params.Offset > maxLimit {
		return fmt.Errorf("limit plus offset exceeded maximum value")
	}

	// check all of the field and relation names to prevent SQL injection or other unexpected behavior
    err := params.Criteria.CheckFieldNames(h.Store.Meta.For({{.ModelTypeName}}{}).SQLFields(true)...)
    if err != nil {
        return err
    }
    err = params.OrderBy.CheckFieldNames(h.Store.Meta.For({{.ModelTypeName}}{}).SQLFields(true)...)
    if err != nil {
        return err
    }
    err = h.Store.Meta.For({{.ModelTypeName}}{}).CheckRelationNames(params.Related...)
    if err != nil {
        return err
    }

	ret := make(map[string]interface{}, 3)

	if params.ReturnCount {
		count, err := h.Store.Search{{.ModelName}}Count(queryCtx,
			params.Criteria, params.OrderBy,
			maxLimit)
		if err != nil {
			return err
		}
		ret["count"] = count
	}

	resultList, err := h.Store.Search{{.ModelName}}(queryCtx,
		params.Criteria, params.OrderBy,
		params.Limit, params.Offset,
		params.Related...)
	if err != nil {
		return err
	}
	ret["result_list"] = resultList
	ret["result_length"] = len(resultList)

	{{/*
	// TODO: It would be nice to sanely implement paging.  I think using
	// headers to convey paging is a crappy way to go - harder for clients
	// expecting JSON to handle.  The idea of "caller needs 10 on a page
	// so ask for 11 and if you get it you know there's another page" should
	// work fine but should be internalized here so that complexity isn't 
	// pushed back to the caller.  So maybe it's just limit, offset that gets
	// passed in and the result {result_list:...:has_more:true} or something. 
	// Do more looking at other APIs and find someone doing it well.
	// https://www.moesif.com/blog/technical/api-design/REST-API-Design-Filtering-Sorting-and-Pagination/

	// We also should support pagination based on keys (i.e. "10 next records after X")
	// so it doesn't choke on large datasets - I can see use cases for both approaches
	*/}}

	{{/* FIXME: so do we return this error or not? if we return it
		then the caller will double-output the error, if not we just
		hid the error - maybe we need a log statement specifically
		for this case, think it through and drop it in */}}

	ar.WriteResult(w, 200, ret)

	return nil
}

func (h *{{.ModelName}}APIController) Fetch(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, {{.ModelNameL}}ID string, related ...string) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}FetchPerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	var {{.ModelNameL}} {{.ModelTypeName}}
	err := h.Store.Fetch{{.ModelName}}(r.Context(), &{{.ModelNameL}}, {{.ModelNameL}}ID, related...)
	if webutil.IsNotFound(err) {
		return &httpapi.ErrorDetail{Code: 404, Message: "not found"}
	}
	if err != nil {
		return err
	}

	// TODO: sometimes return values need to be vetted/scrubbed, both
	// here and in Search, before being returned, where should this
	// call go?  Might be better actually to ensure at the store
	// layer we can just avoid loading the fields we don't need...
	// Although that might be a specific case for certain data types that we
	// don't want to have in everything by default.

	ar.WriteResult(w, 200, {{.ModelNameL}})

	return nil
}

func (h *{{.ModelName}}APIController) Update(w http.ResponseWriter, r *http.Request, ar *httpapi.APIRequest, {{.ModelNameL}}ID string, mapData map[string]interface{}) error {

	if !userctrl.ReqUserHasPerm(r, {{.ModelName}}UpdatePerm) {
		return &httpapi.ErrorDetail{Code: 403, Message: "access denied"}
	}

	{{/*
	// FIXME: THIS IS STILL NEEDED - ESPECIALLY THE ERRORS THAT ROLL UP TO THE UI
	// TODO: Validation!
	// Tags in model
	// Check in store calls
	// Additional validation in controller
	// Errors that roll all the way up to the UI - look at responses codes, how to indicate what validation went wrong etc
	// Translatable
	*/}}

	// TODO: add version number check, so we can do optimistic locking all the
	// way up to the form/UI
	var {{.ModelNameL}} {{.ModelTypeName}}
	err := h.Store.Fetch{{.ModelName}}(r.Context(), &{{.ModelNameL}}, {{.ModelNameL}}ID)
	if err != nil {
		return err
	}

	// patch the fillable fields
	err = httpapi.Fill(&{{.ModelNameL}}, mapData)
	if err != nil {
		return err
	}

	err = h.Store.Update{{.ModelName}}(r.Context(), &{{.ModelNameL}})
	if webutil.IsNotFound(err) {
		return &httpapi.ErrorDetail{Code: 404, Message: "not found"}
	}
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

	err := h.Store.Delete{{.ModelName}}(r.Context(), &{{.ModelTypeName}}{ {{.ModelName}}ID: {{.ModelNameL}}ID } )
	if webutil.IsNotFound(err) {
		return &httpapi.ErrorDetail{Code: 404, Message: "not found"}
	}
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
