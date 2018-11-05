// Utilities for making REST and JSON-RPC over HTTP APIs.
//
// JSON-RPC 2.0 follows the guidelines laid out in http://www.simple-is-better.org/json-rpc/transport_http.html
// NOTE: exception about error responses and how while it's up to the developer the pattern of using
// HTTP response codes in addition to RPC result codes is a workable approach and provides the benefits of
// REST-style repsonse codes while retaining the benefit of a complex error object that can still be
// inspected as needed.
//
// And REST is mostly path matching on the path you provide to each call.
//
// Advanced things to review: Documentation generation based on comments (separate package but should
// play well with this one).  Code-generating clients? (Difficult but not necessarily impossible.)
//
package httpapi

import (
	"encoding/json"
	"mime"
	"net/http"

	"github.com/gocaveman/caveman/weberrors"
	"github.com/gocaveman/caveman/webutil"
)

const (
	TYPE_JSONRPC_2_0 = "jsonrpc-2.0"
	TYPE_REST        = "rest"
)

type APIRequest struct {
	Type string // "jsonrpc-2.0" or "rest"
	// Name    string        // "get-customer" or "GET:/api/customer/123"
	// RPCMethod string        // "get-customer" or "GET:/api/customer/123"
	Input interface{} // input object
	// ID      interface{}   // used only for jsonrpc-2.0 and included in the response if present
	Request *http.Request // the original request

	JSONRPC2Request *JSONRPC2Request // if non-nil then it's the body of the request as previously parsed as a JSONRPC2Request

	Err error // if an error occurred during parsing it is placed here

	// if non-nil it is the contents of Request.Body after being read in full; the first
	// parse attempt will read from Request.Body and store in BodyBytes and subsequence attempts will just use BodyBytes
	// BodyBytes []byte
}

func NewRequest(r *http.Request) *APIRequest {

	// figure out type and name
	// id if present

	ret := &APIRequest{}
	ret.Request = r

	return ret
}

type JSONRPC2Request struct {
	JSONRPC interface{}     `json:"jsonrpc,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Returns true if the request meets the basic criteria for a JSON-RPC 2.0 request, taking into account
// the variations in calling style.  JSON or other parsing errors will be stored in Err if present.
// If false is returned it means the request does not match the criteria and false APIRequest will remain unmodified
// unless a JSON parse error was encounted while trying to match the method name.
func (apir *APIRequest) ParseJSONRPC2(endpointPath string, rpcMethodName string, inputObj interface{}, allowGET bool) bool {

	r := apir.Request

	if r.URL.Path != endpointPath {
		return false
	}

	if r.Method == "GET" {

		if !allowGET {
			return false
		}

		q := r.URL.Query()

		if q.Get("jsonrpc") != "2.0" {
			return false
		}

		rpcMethod := q.Get("method")
		if rpcMethod != rpcMethodName {
			return false
		}

		// at this point it matches and we fill everything out and return true
		// anything else that goes wrong is put into apir.Err

		apir.Type = "jsonrpc-2.0"

		apir.JSONRPC2Request = &JSONRPC2Request{JSONRPC: "2.0"}

		id := q.Get("id")
		if id != "" {
			apir.JSONRPC2Request.ID = id
		}

		apir.JSONRPC2Request.Method = rpcMethod

		params := q.Get("params")
		apir.JSONRPC2Request.Params = json.RawMessage(params)

		err := json.Unmarshal([]byte(apir.JSONRPC2Request.Params), inputObj)
		if err != nil {
			apir.Err = err
			return true
		}

		return true
	}

	if r.Method == "POST" {

		ct, _, _ := mime.ParseMediaType(r.Header.Get("content-type"))
		if ct != "application/json" {
			return false
		}

		if apir.JSONRPC2Request == nil {

			apir.Type = "jsonrpc-2.0"

			apir.JSONRPC2Request = &JSONRPC2Request{}

			defer r.Body.Close()
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(apir.JSONRPC2Request)
			if err != nil {
				apir.Err = err
				return false // it's false in this case because we don't know if the method name matches yet
			}

		}

		// check method name
		if apir.JSONRPC2Request.Method != rpcMethodName {
			return false
		}

		// try to unmarshal
		err := json.Unmarshal([]byte(apir.JSONRPC2Request.Params), inputObj)
		if err != nil {
			apir.Err = err
			return true
		}

		return true
	}

	return false
}

func (apir *APIRequest) ParseRESTPath(httpMethod, pathFormat string, pathArgs ...interface{}) bool {

	r := apir.Request

	if r.Method != httpMethod {
		return false
	}

	err := webutil.PathParse(r.URL.Path, pathFormat, pathArgs...)
	if err != nil {
		return false
	}

	apir.Type = TYPE_REST

	return true
}

func (apir *APIRequest) ParseRESTObj(httpMethod string, obj interface{}, path string) bool {

	r := apir.Request

	if r.Method != httpMethod {
		return false
	}

	if r.URL.Path != path {
		return false
	}

	ct, _, _ := mime.ParseMediaType(r.Header.Get("content-type"))
	if ct == "application/json" {

		apir.Type = TYPE_REST
		apir.Input = obj

		defer apir.Request.Body.Close()
		dec := json.NewDecoder(apir.Request.Body)
		err := dec.Decode(obj)
		if err != nil {
			apir.Err = err
			return true
		}

		return true

	} else if ct == "application/x-www-form-urlencoded" || httpMethod == "GET" || httpMethod == "HEAD" {

		apir.Type = TYPE_REST
		apir.Input = obj
		apir.Err = r.ParseForm()
		if apir.Err != nil {
			return true
		}

		apir.Err = FormUnmarshal(r.Form, obj)
		return true

	}

	return false
}

func (apir *APIRequest) ParseRESTObjPath(httpMethod string, obj interface{}, pathFormat string, pathArgs ...interface{}) bool {

	r := apir.Request

	if r.Method != httpMethod {
		return false
	}

	err := webutil.PathParse(r.URL.Path, pathFormat, pathArgs...)
	if err != nil {
		return false
	}

	ct, _, _ := mime.ParseMediaType(r.Header.Get("content-type"))
	if ct == "application/json" {

		apir.Type = TYPE_REST
		apir.Input = obj

		defer apir.Request.Body.Close()
		dec := json.NewDecoder(apir.Request.Body)
		err = dec.Decode(obj)
		if err != nil {
			apir.Err = err
			return true
		}

		return true

	} else if ct == "application/x-www-form-urlencoded" || httpMethod == "GET" || httpMethod == "HEAD" {

		apir.Type = TYPE_REST
		apir.Input = obj
		apir.Err = r.ParseForm()
		if apir.Err != nil {
			return true
		}

		apir.Err = FormUnmarshal(r.Form, obj)
		return true

	}

	return false

}

// apir.ParseRESTPath("GET", "/api/something/%s", &something.ID)
// apir.ParseRESTPath("DELETE", "/api/something/%s", &something.ID)
// apir.ParseRESTObj("POST", &something, "/api/something")
// apir.ParseRESTObjPath("PATCH", &something, "/api/something/%s", &something.ID)
// apir.ParseRESTObj("GET", &something, "/api/something") // form values are implied here

func (apir *APIRequest) Fill(obj interface{}) error {
	return Fill(obj, apir.Input)
}

func (apir *APIRequest) WriteResult(w http.ResponseWriter, code int, obj interface{}) error {

	if apir.Type == TYPE_JSONRPC_2_0 {

		// for JSONRPC2 we wrap it with an object with a single "result" property
		result := map[string]interface{}{
			"result": obj,
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(code)
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		err := enc.Encode(result)

		return err

	}

	// else if apir.Type == TYPE_REST {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(obj)

	return err

}

// type ErrorDetail struct {
// 	Code    int         `json:"code"`
// 	Message string      `json:"message"`
// 	Data    interface{} `json:"data,omitempty"`
// }

// func (e *ErrorDetail) Error() string {
// 	return e.Message
// }

// func (e *ErrorDetail) ErrorCode() int {
// 	return e.Code
// }

// type ErrorCoder interface {
// 	ErrorCode() int
// }

// FIXME: what about adding a method that acts like HTTPError - think that through
// I think what's needed here is that by default everything acts like HTTPError
// but we have interface(s) that allow the public message and error code and detail
// data to be provided.

func (apir *APIRequest) WriteCodeErr(w http.ResponseWriter, code int, err error) error {
	return apir.writeErr(w, err, code,
		weberrors.ErrorMessage(err),
		weberrors.ErrorData(err),
		weberrors.ErrorHeaders(err),
	)
}

func (apir *APIRequest) WriteErr(w http.ResponseWriter, err error) error {
	return apir.writeErr(w, err,
		weberrors.ErrorCode(err),
		weberrors.ErrorMessage(err),
		weberrors.ErrorData(err),
		weberrors.ErrorHeaders(err),
	)
}

func (apir *APIRequest) writeErr(w http.ResponseWriter, err error, code int, message string, data interface{}, h http.Header) error {

	detail := weberrors.New(err, code, message, data, h)
	// detail := &weberrors.Detail{
	// 	Err:     err,
	// 	Code:    code,
	// 	Message: message,
	// 	Data:    data,
	// }

	w.Header().Set("content-type", "application/json")
	httpCode := code
	if httpCode < 100 || httpCode >= 600 {
		httpCode = 500
	}
	// assign any headers from h not already set
	for k, v := range h {
		if _, ok := w.Header()[k]; !ok {
			w.Header()[k] = v
		}
	}
	w.WriteHeader(httpCode)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	if apir.Type == TYPE_JSONRPC_2_0 {

		// jsonrpc2 gets wrapped in another object with "error" property
		err := enc.Encode(map[string]interface{}{
			"error": detail,
		})
		return err

	}

	err = enc.Encode(detail)
	return err
}

// FIXME: use weberrors instead of our own ErrorDetail DONE
// FIXME: make WriteError have the correct behavior for code, message, data NEEDS TESTING
// FIXME: add something to enable logging upon WriteError(); also be sure to log an ID and send
// X-Id so everything can be matched up.
