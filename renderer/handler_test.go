package renderer

import (
	"bytes"
	"log"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/spf13/afero"
)

func TestHandler(t *testing.T) {

	viewFS := afero.NewMemMapFs()

	afero.WriteFile(viewFS, "/example.gohtml", []byte(`---

some: data
goes: here

---
{{template "/main-page.gohtml" .}}

{{define "body"}}

<div>this is the body</div>

{{end}}

`), 0755)

	includeFS := afero.NewMemMapFs()

	afero.WriteFile(includeFS, "/main-page.gohtml", []byte(`<!doctype html>
<html>
<body>

before body

{{block "body" .}}{{end}}

after body

`), 0755)

	renderer := New(afero.NewHttpFs(viewFS), afero.NewHttpFs(includeFS))

	handler := NewHandler(renderer)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/example", nil)
	handler.ServeHTTP(w, r)

	b, _ := httputil.DumpResponse(w.Result(), true)
	log.Printf("RESULT:\n%s", b)

	if !bytes.Contains(b, []byte("before body")) {
		t.Fatalf("missing 'before body'")
	}

	if !bytes.Contains(b, []byte("this is the body")) {
		t.Fatalf("missing 'this is the body'")
	}

	// try a part now
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/example?part=body", nil)
	handler.ServeHTTP(w, r)

	b, _ = httputil.DumpResponse(w.Result(), true)
	log.Printf("RESULT:\n%s", b)

	if bytes.Contains(b, []byte("before body")) {
		t.Fatalf("has 'before body'!")
	}

	if !bytes.Contains(b, []byte("this is the body")) {
		t.Fatalf("missing 'this is the body'")
	}

}
