package webutil

import (
	"context"
	"fmt"
	"net/http"
)

// ChainHandler is a slight variation of http.Handler which returns a new "w" and "r", allowing for them to be changed or replaced
// (and thus allowing the request context to be changed as well).
type ChainHandler interface {
	ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request)
}

// ChainHandlerFunc adapts a function to implement ChainHandler (same pattern as http.HandlerFunc)
type ChainHandlerFunc func(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request)

func (f ChainHandlerFunc) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {
	return f(w, r)
}

// HandlerList is a slice of ChainHandler or http.Handler instances.
// When ServeHTTP is called each one is called in turn, trying as ChainHandler first and if that fails then http.Handler.
// Will panic if you put any other type in the slice.
type HandlerList []interface{}

// AppendChainHandler adds a ChainHandler to the list, returning the new list.  (Type-safe wrapper around append())
func (hl HandlerList) AppendChainHandler(ch ChainHandler) HandlerList {
	return append(hl, ch)
}

// AppendHandler adds a http.Handler to the list, returning the new list.  (Type-safe wrapper around append())
func (hl HandlerList) AppendHandler(h http.Handler) HandlerList {
	return append(hl, h)
}

// ServeHTTP calls ServeHTTPChain
func (hl HandlerList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hl.ServeHTTPChain(w, r)
}

// ServeHTTPChain calls each element in the list. trying as ChainHandler first and if that fails then http.Handler.
// Will panic if you put any other type in the slice.
func (hl HandlerList) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {

	for _, h := range hl {

		if r.Context().Err() != nil {
			return
		}

		// try ChainHandler first
		ch, ok := h.(ChainHandler)
		if ok {
			w, r = ch.ServeHTTPChain(w, r)
			continue
		}

		// otherwise http.Handler
		hh, ok := h.(http.Handler)
		if ok {
			hh.ServeHTTP(w, r)
			continue
		}

		panic(fmt.Errorf("item in HandlerList (type=%t, val=%+v) is not a ChainHandler nor http.Handler", h, h))
	}

	return w, r
}

// ServeHTTPChain checks h to see if it implements ChainHandler and calls ServeHTTPChain if so.  Otherwise it falls
// back to http.Handler and ServeHTTP.  The returned w and r will be as returned by ServeHTTPChain or if
// falling back to ServeHTTP same as input.  Will panic if h does not implement either interface.
func ServeHTTPChain(h interface{}, w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {

	h1, ok := h.(ChainHandler)
	if ok {
		return h1.ServeHTTPChain(w, r)
	}

	h2, ok := h.(http.Handler)
	if ok {
		h2.ServeHTTP(w, r)
		return w, r
	}

	panic(fmt.Errorf("ServeHTTPChain: %#v is not a valid ChainHandler or http.Handler", h))

}

// NewCtxSetHandler is a helper that returns a ChainHandler which assigns a static value to the request context each time
// it is called.  (NOTE: If you want a dynamic value just implement ChainHandler yourself.  The point of this function
// is to make static values assignable as one-liners in your configuration code, e.g. making configuration values
// available in a template.)
func NewCtxSetHandler(key string, value interface{}) ChainHandler {
	return ChainHandlerFunc(func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
		return w, r.WithContext(context.WithValue(r.Context(), key, value))
	})
}

// TODO: implement this one:
// func NewCtxMapHandler(ctxMap map[string]interface{}) *CtxMapHandler {

// TODO: what about PriorityHandlerList, which is a slice of PriorityHandler structs which have a priority float 0-100
// and gets sorted before use (maybe it gets sorted each time something is added?)
// Also think about having a Name field or similar and the ability to find or remove based on names.  It might
// be very useful to find the default provider for something by its name and make some changes to it or remove it.
