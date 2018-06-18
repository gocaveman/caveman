package regions

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/gocaveman/caveman/renderer"
	"github.com/gocaveman/caveman/tmpl"
	"github.com/gocaveman/caveman/webutil"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRegions(t *testing.T) {

	// start with renderer test, trim anything we can and
	// add region stuff to it

	// check enabled, request matches, ContextMeta and sequence

	assert := assert.New(t)

	fs := afero.NewMemMapFs()

	fs.MkdirAll("/test/views", 0755)
	fs.MkdirAll("/test/views/sampledir", 0755)
	fs.MkdirAll("/test/includes", 0755)

	afero.WriteFile(fs, "/test/views/sampledir/example.gohtml", []byte(`---
some: data
goes: here
---
{{template "/main-page.gohtml" .}}
{{define "body"}}body_section{{end}}
`), 0644)

	afero.WriteFile(fs, "/test/includes/main-page.gohtml", []byte(`<!doctype html>
<html><head>
{{block "style+" .}}{{end}}
</head><body>
{{block "leftnav+" .}}{{end}}
{{block "body" .}}{{end}}
{{block "script+" .}}{{end}}
</body></html>
`), 0755)

	afero.WriteFile(fs, "/test/includes/bar1.gohtml", []byte(`bar1content`), 0755)
	afero.WriteFile(fs, "/test/includes/bar2.gohtml", []byte(`bar2content`), 0755)

	tmplStore := &tmpl.HFSStore{
		FileSystems: map[string]http.FileSystem{
			tmpl.ViewsCategory:    afero.NewHttpFs(afero.NewBasePathFs(fs, "/test/views")),
			tmpl.IncludesCategory: afero.NewHttpFs(afero.NewBasePathFs(fs, "/test/includes")),
		},
	}

	rend := renderer.NewFromTemplateReader(tmplStore)

	blockDefineHandler := renderer.NewBlockDefineHandler()
	reqCtxHandler := webutil.NewRequestContextHandler()
	rendHandler := renderer.NewHandler(rend)

	regionHandler := &RegionHandler{Store: nil}
	_ = regionHandler

	hl := webutil.NewDefaultHandlerList(blockDefineHandler, reqCtxHandler, regionHandler, rendHandler)

	runReq := func(s Store, urlPath string) string {

		regionHandler.Store = s

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", urlPath, nil)
		hl.ServeHTTP(w, r)

		b, err := httputil.DumpResponse(w.Result(), true)
		if err != nil {
			panic(err)
		}
		return string(b)

	}

	var resStr string

	resStr = runReq(ListStore{
		Definition{
			DefinitionID: "def1",
			RegionName:   "leftnav+",
			Sequence:     1,
			TemplateName: "/bar1.gohtml",
		},
	}, "/sampledir/example")
	t.Logf("RESULT: %s", resStr)
	assert.Contains(resStr, "bar1content")

}
