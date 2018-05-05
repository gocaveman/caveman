package renderer

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/gocaveman/caveman/tmpl"
	"github.com/spf13/afero"
)

func TestHandlerFromTemplateReader(t *testing.T) {

	fs := afero.NewMemMapFs()

	fs.MkdirAll("/test/views", 0755)
	fs.MkdirAll("/test/includes", 0755)

	afero.WriteFile(fs, "/test/views/example.gohtml", []byte(`---

some: data
goes: here

---
{{template "/main-page.gohtml" .}}

{{define "body"}}
{{$meta := First (.Value "pageinfo.Meta") (.Value "tmpl.Meta")}}

<div>this is the body: {{$meta.some}}</div>

{{end}}

`), 0644)

	afero.WriteFile(fs, "/test/includes/main-page.gohtml", []byte(`<!doctype html>
<html>
<body>

before body

{{block "body" .}}{{end}}

after body

`), 0755)

	tmplStore := &tmpl.HFSStore{
		FileSystems: map[string]http.FileSystem{
			ViewsCategory:    afero.NewHttpFs(afero.NewBasePathFs(fs, "/test/views")),
			IncludesCategory: afero.NewHttpFs(afero.NewBasePathFs(fs, "/test/includes")),
		},
	}

	renderer := NewFromTemplateReader(tmplStore)

	// TODO: test pageinfo.Meta and doing a render from another custom handler (not via this handler).
	// Maybe it belongs in another file...

	handler := NewHandler(renderer)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/example", nil)
	handler.ServeHTTP(w, r)

	b, _ := httputil.DumpResponse(w.Result(), true)
	log.Printf("RESULT:\n%s", b)

	if !bytes.Contains(b, []byte("before body")) {
		t.Fatalf("missing 'before body'")
	}

	if !bytes.Contains(b, []byte("this is the body: data")) {
		t.Fatalf("missing 'this is the body: data'")
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

	if !bytes.Contains(b, []byte("this is the body: data")) {
		t.Fatalf("missing 'this is the body: data'")
	}

}

// DEPRECATED - we're moving away from this and toward the TemplateReader approach.
func TestHandlerFromFSs(t *testing.T) {

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

	renderer := NewFromFSs(afero.NewHttpFs(viewFS), afero.NewHttpFs(includeFS))

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
