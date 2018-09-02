package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"sync"

	"github.com/gocaveman/caveman/uiproto/files"
)

func main() {

	flag.Parse()

	// ln, err := net.Listen("tcp", "127.0.0.1:0")
	ln, err := net.Listen("tcp", "127.0.0.1:5561")
	if err != nil {
		log.Fatal(err)
	}

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		Handler: &MainHandler{},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Serve(ln)
		if err != nil {
			log.Printf("Error from HTTP server: %v", err)
		}
	}()

	mainURL := fmt.Sprintf("http://127.0.0.1:%s/index.html", port)
	log.Printf("Main URL: %s", mainURL)

	wg.Wait()

}

type MainHandler struct{}

func (ws *MainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fpath := path.Join("/static", path.Clean("/"+r.URL.Path))

	f, err := files.EmbeddedAssets.Open(fpath)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()
	fst, err := f.Stat()
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), 500)
		return
	}

	http.ServeContent(w, r, path.Base(fpath), fst.ModTime(), f)

}
