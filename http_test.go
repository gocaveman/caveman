package caveman

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestChainedHandler struct {
	Out  io.Writer
	Msg  string
	next http.Handler
}

func (h *TestChainedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("ServeHTTP %q", h.Msg)
	h.Out.Write([]byte(h.Msg))
	if h.next == nil {
		panic(fmt.Errorf("nil next Msg=%q", h.Msg))
	}
	log.Printf("  calling next.ServeHTTP %q", h.Msg)
	h.next.ServeHTTP(w, r)
}

func (h *TestChainedHandler) SetNextHandler(next http.Handler) (self ChainedHandler) {
	h.next = next
	return h
}

type TestRegularHandler struct {
	Out io.Writer
	Msg string
}

func (h *TestRegularHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("ServeHTTP %q", h.Msg)
	h.Out.Write([]byte(h.Msg))
}

func TestBuildHandlerChain(t *testing.T) {

	var out bytes.Buffer

	h := BuildHandlerChain(HandlerList{
		&TestChainedHandler{Out: &out, Msg: "1."},
		&TestChainedHandler{Out: &out, Msg: "2."},
		&TestRegularHandler{Out: &out, Msg: "3."},
		&TestChainedHandler{Out: &out, Msg: "4."},
		&TestRegularHandler{Out: &out, Msg: "5."},
		&TestRegularHandler{Out: &out, Msg: "6."},
		&TestChainedHandler{Out: &out, Msg: "7."},
		&TestRegularHandler{Out: &out, Msg: "8."},
		&TestChainedHandler{Out: &out, Msg: "9."},
		&TestRegularHandler{Out: &out, Msg: "10."},
		&TestRegularHandler{Out: &out, Msg: "11."},
		&TestChainedHandler{Out: &out, Msg: "12."},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, r)

	t.Logf("TestBuildHandlerChain results: %q", out.String())

	if out.String() != "1."+"2."+"3."+"4."+"5."+"6."+"7."+"8."+"9."+"10."+"11."+"12." {
		t.Fatalf("Bad TestBuildHandlerChain output!")
	}

}
