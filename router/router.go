package router

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sort"

	"github.com/gocaveman/caveman/webutil"
)

// ContextKey gives the names of things we use in the context their own type, to ensure uniqueness.
type ContextKey string

const (
	// RequestWrittenKey is the name of the key used by RequestWritten().
	RequestWrittenKey ContextKey = "router.RequestWritten"
	// DeferChainHandlerListKey is user by DeferChainHandler().
	DeferChainHandlerListKey ContextKey = "router.DeferChainHandlerList"
)

// common sequences used for RouteSequence()
const (
	RouteSequenceFirst      float64 = -1000.0 // before everything else
	RouteSequenceSetup      float64 = -500.0  // before middleware runs
	RouteSequenceMiddleware float64 = -100.0  // modify request/response before further processing
	RouteSequenceHandler    float64 = 0.0     // main request handling
	RouteSequenceLast       float64 = 10009.0 // after everything else
)

// ChainHandler is similar to http.Handler but can allow a request to be modified and delegated to a later handler.
type ChainHandler interface {
	ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request)
}

// ChainHandlerFunc adapts a function to implement ChainHandler (same pattern as http.HandlerFunc)
type ChainHandlerFunc func(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request)

// ChainHandlerList is a slice of ChainHandler.
type ChainHandlerList []ChainHandler

// ServeHTTPChain implements ChainHandler.
func (f ChainHandlerFunc) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {
	return f(w, r)
}

// RouteHandler must be implemented by
type RouteHandler interface {
	RoutePathPrefix() string // compared with webutil.HasPathPrefix(), empty string means always called
	RouteSequence() float64  // lower numbers are called first, see constants in this file, by convention range is 0-100.0
	ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request)
}

// RouteHandlerList is a slice of RouteHandler instances.  It implements proper sorting sequence.
type RouteHandlerList []RouteHandler

// Len is part of sort.Interface.
func (rhl RouteHandlerList) Len() int {
	return len(rhl)
}

// Swap is part of sort.Interface.
func (rhl RouteHandlerList) Swap(i, j int) {
	rhl[i], rhl[j] = rhl[j], rhl[i]
}

// Less is part of sort.Interface.  Orders by sequence, path prefix and then the
// underlying type named printed as a string using fmt package's "%T".
func (rhl RouteHandlerList) Less(i, j int) bool {
	ri, rj := rhl[i], rhl[j]

	// try our very best to be deterministic...

	// if different sequences, use that
	seqi, seqj := ri.RouteSequence(), rj.RouteSequence()
	if seqi != seqj {
		return seqi < seqj
	}

	// if same, then use path prefix
	ppi, ppj := ri.RoutePathPrefix(), rj.RoutePathPrefix()
	if ppi != ppj {
		return ppi < ppj
	}

	// if still the same then use the type name
	ti, tj := fmt.Sprintf("%T", ri), fmt.Sprintf("%T", rj)
	return ti < tj
}

// DefaultRouteHandler implements RouteHandler by delegating to a ChainHandler you provide.
type DefaultRouteHandler struct {
	ChainHandler
	PathPrefix string
	Sequence   float64
}

// RoutePathPrefix returns the path prefix.
func (h *DefaultRouteHandler) RoutePathPrefix() string {
	return h.PathPrefix
}

// RouteSequence returns the sequence.
func (h *DefaultRouteHandler) RouteSequence() float64 {
	return h.Sequence
}

// HTTPRouteHandler implements RouteHandler by delegating to an http.Handler you provide.
type HTTPRouteHandler struct {
	http.Handler
	PathPrefix string
	Sequence   float64
}

// RoutePathPrefix returns the path prefix.
func (h *HTTPRouteHandler) RoutePathPrefix() string {
	return h.PathPrefix
}

// RouteSequence returns the sequence.
func (h *HTTPRouteHandler) RouteSequence() float64 {
	return h.Sequence
}

// ServeHTTPChain calls the Handler and returns w and r unmodified.
func (h *HTTPRouteHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {
	wnext, rnext = w, r
	h.Handler.ServeHTTP(w, r)
	return
}

// New will create a new empty HandlerSet ready for use.
func New() *HandlerSet {
	return &HandlerSet{}
}

// HandlerSet is a group of handlers that will receive requests.
type HandlerSet struct {
	handlerList RouteHandlerList
}

// Add will add a RouteHandler.  If it is already in the list (based on simple == comparison)
// it will not be added again.
func (hs *HandlerSet) Add(rh RouteHandler) {
	for _, h := range hs.handlerList {
		if h == rh {
			return
		}
	}
	hs.handlerList = append(hs.handlerList, rh)
	sort.Sort(hs.handlerList)
}

// ServeHTTPChain delegates requests to the RouteHandler instances in this set according to basic routing rules.
func (hs *HandlerSet) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {

	wnext, rnext = w, r

	// set up request written so it will be visible for all derived contexts
	requestWrittenVal := RequestWritten(rnext.Context()) // defaults to false
	rnext = rnext.WithContext(SetRequestWritten(rnext.Context(), requestWrittenVal))

	// defer an empty handler to make sure that is set up properly
	rnext = rnext.WithContext(DeferChainHandler(rnext.Context(), ChainHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {
			return w, r
		})))

	// wrap the response writer to call SetRequestWritten at the appropriate times
	wwrap := NewWrapResponseWriter(wnext, rnext)
	wwrap.SetWriteFunc(func(b []byte) (int, error) {
		defer SetRequestWritten(rnext.Context(), true)
		return wwrap.Parent().Write(b)
	})
	wwrap.SetWriteHeaderFunc(func(s int) {
		defer SetRequestWritten(rnext.Context(), true)
		wwrap.Parent().WriteHeader(s)
	})
	wnext = wwrap

	pclean := path.Clean("/" + r.URL.Path)

	for _, h := range hs.handlerList {
		hp := h.RoutePathPrefix()
		// empty or matching path prefix is called
		if hp == "" || webutil.HasPathPrefix(pclean, hp) {
			wnext, rnext = h.ServeHTTPChain(wnext, rnext)
		}
		if RequestWritten(rnext.Context()) {
			break
		}
	}

	// callbacks to anything that requested it
	lp, _ := rnext.Context().Value(DeferChainHandlerListKey).(*ChainHandlerList)
	if lp != nil {
		// call in reverse sequence
		for i := len(*lp) - 1; i >= 0; i-- {
			ch := (*lp)[i]
			wnext, rnext = ch.ServeHTTPChain(wnext, rnext)
		}
	}

	return
}

// ServeHTTP implements http.Handler.
func (hs *HandlerSet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs.ServeHTTPChain(w, r)
}

// RequestWritten will return the "router.RequestWritten" value from the context, if it exists otherwise false.
// This indicates that WriteHeader() has been called on the ResponseWriter and thus the response
// has been handled.
func RequestWritten(ctx context.Context) bool {
	rw, ok := ctx.Value(RequestWrittenKey).(*bool)
	if !ok {
		return false
	}
	return *rw
}

// SetRequestWritten will assign a value to "router.RequestWritten" in the context and return the new context.
// Note that if this has already been called then the context will not need to be modified and the same
// one will be returned.
func SetRequestWritten(ctx context.Context, v bool) context.Context {

	rw, ok := ctx.Value(RequestWrittenKey).(*bool)
	if !ok {
		vv := v
		rw = &vv
		ctx = context.WithValue(ctx, RequestWrittenKey, rw)
	}

	*rw = v
	return ctx
}

// DeferChainHandler adds a ChainHandler to be called when request processing is completed.
func DeferChainHandler(ctx context.Context, ch ChainHandler) context.Context {

	lp, ok := ctx.Value(DeferChainHandlerListKey).(*ChainHandlerList)
	if !ok {
		var chl ChainHandlerList
		lp = &chl
		ctx = context.WithValue(ctx, DeferChainHandlerListKey, lp)
	}

	// push to the end for simplicity and speed while registering - the calls
	// get done afterward in reverse (element 0 is called last)
	*lp = append(*lp, ch)
	return ctx
}

// PathPrefixAndSequence can be embedded in another struct to easily implement
// RoutePathPrefix and RouteSequence methods.
type PathPrefixAndSequence struct {
	PathPrefix string
	Sequence   float64 // remember to set this properly, e.g. to RouteSequenceHandler
}

// RoutePathPrefix is part of RouteHandler.
func (p PathPrefixAndSequence) RoutePathPrefix() string {
	return p.PathPrefix
}

// RouteSequence is part of RouteHandler.
func (p PathPrefixAndSequence) RouteSequence() float64 {
	return p.Sequence
}
