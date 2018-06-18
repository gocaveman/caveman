// The themetester is a standalone executable intended to help internal dev and testing of Caveman themes.
// It is also intended to be easy to copy or modify for your own theme development.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/gocaveman/caveman/autowire"
	"github.com/gocaveman/caveman/renderer"
	"github.com/gocaveman/caveman/tmpl"
	"github.com/gocaveman/caveman/webutil"
	"github.com/gocaveman/caveman/webutil/handlerregistry"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gocaveman/caveman/themes/paleolithic"
	"github.com/gocaveman/caveman/themes/pleistocene"
)

func main() {

	themes := make(map[string]tmpl.Store)

	// add each theme you want to be available here
	defaultTheme := "paleolithic"
	themes["paleolithic"] = paleolithic.NewTmplStore()
	themes["pleistocene"] = pleistocene.NewTmplStore()

	pflag.StringP("http-listen", "l", ":2666", "IP:Port to listen on for HTTP")
	pflag.StringP("default-theme", "", defaultTheme, "Default theme name on startup")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	defaultTheme = viper.GetString("default-theme")
	syncStore := tmpl.NewSyncStore(themes[defaultTheme])

	themeSwitchHandler := &ThemeSwitchHandler{
		SyncStore:    syncStore,
		DefaultTheme: defaultTheme,
		ThemeMap:     themes,
	}

	hl := webutil.NewDefaultHandlerList()
	hl = append(hl, themeSwitchHandler)
	for _, item := range handlerregistry.Contents() {
		hl = append(hl, item.Value)
	}

	rend := renderer.NewFromTemplateReader(syncStore)
	autowire.Provide("", rend)
	rendHandler := renderer.NewHandler(rend)
	hl = append(hl, rendHandler)
	hl = append(hl, renderer.NotFoundHandler(rend, "/_404.gohtml"))

	err := autowire.Contents().Run()
	if err != nil {
		log.Fatalf("autowire error: %v", err)
	}

	var wg sync.WaitGroup

	httpListen := viper.GetString("http-listen")
	webutil.StartHTTPServer(&http.Server{
		Addr:    httpListen,
		Handler: hl,
	}, &wg)

	log.Printf("Web server starting at %q; use %q to set the theme", httpListen, `/api/theme-switch?id=`+defaultTheme)

	wg.Wait()
}

type ThemeSwitchHandler struct {
	SyncStore    *tmpl.SyncStore
	DefaultTheme string
	CurrentTheme string
	ThemeMap     map[string]tmpl.Store
	sync.RWMutex
}

func (h *ThemeSwitchHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (w2 http.ResponseWriter, r2 *http.Request) {

	w2, r2 = w, r

	if r.URL.Path == "/api/theme-switch" {

		h.Lock()
		defer h.Unlock()

		w.Header().Set("content-type", "application/json")
		id := r.FormValue("id")
		if _, ok := h.ThemeMap[id]; ok && id != "" {
			h.CurrentTheme = id
			h.SyncStore.SetStore(h.ThemeMap[h.CurrentTheme])
		} else {
			w.WriteHeader(400)
			fmt.Fprint(w, `{"error":{"code":400,"message":"invalid id"}}`)
			return
		}
		fmt.Fprint(w, `{"result":true}`)
		return
	}

	h.RLock()
	defer h.RUnlock()

	currentTheme := h.DefaultTheme

	// provide the theme info in the context
	var themeNameList []string
	for name := range h.ThemeMap {
		themeNameList = append(themeNameList, name)
	}
	sort.Strings(themeNameList)

	ctx := r2.Context()
	ctx = context.WithValue(ctx, "themetester.CurrentTheme", currentTheme)
	ctx = context.WithValue(ctx, "themetester.DefaultTheme", h.DefaultTheme)
	ctx = context.WithValue(ctx, "themetester.ThemeNameList", themeNameList)
	r2 = r2.WithContext(ctx)

	return
}
