package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testHandler struct{}

func (t *testHandler) post(w http.ResponseWriter, req *http.Request) {

	b := []byte("HELLO WORLD")

	w.WriteHeader(200)
	_, _ = w.Write(b)

}

func TestRouter(t *testing.T) {

	var r Router

	var testHandler testHandler

	r.RegisterRoute("GET", "/api/users/:id", testHandler.post)

	w := httptest.NewRecorder()

	request := httptest.NewRequest("GET", "/api/users/123", nil)

	r.ServeHTTP(w, request)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fail()
	}

	requestNotFound := httptest.NewRequest("POST", "/api/users/123", nil)

	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, requestNotFound)

	respNotFound := w2.Result()

	if respNotFound.StatusCode != 404 {
		t.Fail()
	}

	requestWithParams := httptest.NewRequest("POST", "/api/users/123?param=test&param2=test2", nil)

	params := r.QueryParams(requestWithParams)

	for _, param := range params {
		t.Log(param)
	}

}
