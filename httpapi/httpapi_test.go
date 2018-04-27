package httpapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Something struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// func TestAPIStuff(t *testing.T) {

// 	var something Something

// 	r, _ := http.NewRequest("GET", "/api/something/123", nil)
// 	// ParseRequest(r).UseInput(&something)

// 	apir := NewRequest(r)

// 	apir.ParseJSONRPC("/api/jsonrpc", "get-something", &something)
// 	apir.ParseRESTPath("GET", "/api/something/%s", &something.ID)
// 	apir.ParseRESTPath("DELETE", "/api/something/%s", &something.ID)
// 	apir.ParseRESTObj("POST", &something, "/api/something")
// 	apir.ParseRESTObjPath("PATCH", &something, "/api/something/%s", &something.ID)
// 	apir.ParseRESTObj("GET", &something, "/api/something") // form values are implied here

// 	if apir.ParseJSONRPC("/api/jsonrpc", "get-something", &something) || apir.ParseRESTPath("GET", "/api/something/%s", &something.ID) {

// 		if apir.Err != nil {
// 			apir.WriteError()
// 			return
// 		}

// 		// do stuff with apir and something

// 		// we can also use Fill to copy to a db object we selected
// 		apir.Fill(obj)

// 		apir.WriteResult()

// 	}

// 	// if apir.RPCMethod == "get-something" || r.Method == "GET" && webutil.PathParse(r.URL.Path, "/api/something/%s", &something.ID) {
// 	// }

// }

func TestParseJSONRPC2_GET(t *testing.T) {

	assert := assert.New(t)

	hr, _ := http.NewRequest("GET", `/api/jsonrpc?jsonrpc=2.0&method=get-something&id=987&params={"id":"123"}`, nil)
	ar := NewRequest(hr)

	var something Something

	assert.False(ar.ParseJSONRPC2("/wrong-endpoint", "get-something", &something, true))
	assert.False(ar.ParseJSONRPC2("/api/jsonrpc", "wrong-method", &something, true))
	assert.True(ar.ParseJSONRPC2("/api/jsonrpc", "get-something", &something, true))
	assert.Equal("123", something.ID)

}

func TestParseJSONRPC2_POST(t *testing.T) {

	assert := assert.New(t)

	body := []byte(`{"jsonrpc":"2.0","method":"get-something","id":"987","params":{"id":"123"}}`)
	hr, _ := http.NewRequest("POST", `/api/jsonrpc`, bytes.NewReader(body))
	hr.Header.Set("content-type", "application/json")
	ar := NewRequest(hr)

	var something Something

	assert.False(ar.ParseJSONRPC2("/wrong-endpoint", "get-something", &something, true))
	assert.False(ar.ParseJSONRPC2("/api/jsonrpc", "wrong-method", &something, true))
	assert.True(ar.ParseJSONRPC2("/api/jsonrpc", "get-something", &something, true))
	assert.Equal("123", something.ID)

}

func TestJSONRPC2_Write(t *testing.T) {

	assert := assert.New(t)

	body := []byte(`{"jsonrpc":"2.0","method":"get-something","id":"987","params":{"id":"123"}}`)
	hr, _ := http.NewRequest("POST", `/api/jsonrpc`, bytes.NewReader(body))
	hr.Header.Set("content-type", "application/json")
	ar := NewRequest(hr)

	var something Something

	assert.True(ar.ParseJSONRPC2("/api/jsonrpc", "get-something", &something, true))
	assert.NoError(ar.Err)

	wrec := httptest.NewRecorder()
	assert.NoError(ar.WriteResult(wrec, 200, "this is \" the \nresult"))
	assert.Contains(wrec.Body.String(), `"result":"this is \" the \nresult"`)

	wrec = httptest.NewRecorder()
	assert.NoError(ar.WriteError(wrec, 409, fmt.Errorf("something strange happened")))
	// t.Logf("Error result:\n%s", wrec.Body.String())
	assert.Contains(wrec.Body.String(), `"error":{"code"`)

	wrec = httptest.NewRecorder()
	assert.NoError(ar.WriteError(wrec, 409, &ErrorDetail{Code: 1000, Message: "Blah", Data: nil}))
	// t.Logf("Error result:\n%s", wrec.Body.String())
	assert.Contains(wrec.Body.String(), `"error":{"code":1000,"message":"Blah"`)

}

func TestParseRESTPath(t *testing.T) {

	assert := assert.New(t)

	hr, _ := http.NewRequest("GET", `/api/something/123`, nil)
	ar := NewRequest(hr)

	var something Something

	assert.True(ar.ParseRESTPath("GET", "/api/something/%s", &something.ID))
	assert.Equal("123", something.ID)

}

func TestParseRESTObj(t *testing.T) {

	assert := assert.New(t)

	// try as JSON
	something := Something{}
	hr, _ := http.NewRequest("POST", `/api/something`, bytes.NewReader([]byte(`{"id":"id1","name":"name1"}`)))
	hr.Header.Set("content-type", "application/json")
	ar := NewRequest(hr)

	assert.True(ar.ParseRESTObj("POST", &something, "/api/something"))
	assert.Equal("id1", something.ID)
	assert.Equal("name1", something.Name)

	// try again as form post
	something = Something{}
	hr, _ = http.NewRequest("POST", `/api/something`, bytes.NewReader([]byte(`id=id1&name=name1`)))
	hr.Header.Set("content-type", "application/x-www-form-urlencoded")
	ar = NewRequest(hr)

	assert.True(ar.ParseRESTObj("POST", &something, "/api/something"))
	assert.Equal("id1", something.ID)
	assert.Equal("name1", something.Name)

}

func TestParseRESTObjPath(t *testing.T) {

	assert := assert.New(t)

	// try as JSON
	something := Something{}
	hr, _ := http.NewRequest("POST", `/api/something/123`, bytes.NewReader([]byte(`{"name":"name1"}`)))
	hr.Header.Set("content-type", "application/json")
	ar := NewRequest(hr)

	assert.True(ar.ParseRESTObjPath("POST", &something, "/api/something/%s", &something.ID))
	assert.Equal("123", something.ID)
	assert.Equal("name1", something.Name)

	// try again as form post
	something = Something{}
	hr, _ = http.NewRequest("POST", `/api/something/123`, bytes.NewReader([]byte(`name=name1`)))
	hr.Header.Set("content-type", "application/x-www-form-urlencoded")
	ar = NewRequest(hr)

	assert.True(ar.ParseRESTObjPath("POST", &something, "/api/something/%s", &something.ID))
	assert.Equal("123", something.ID)
	assert.Equal("name1", something.Name)

}
