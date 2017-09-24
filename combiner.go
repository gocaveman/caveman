package caveman

import (
	"context"
	"fmt"
	"net/http"
)

type FileCombiner struct {
	PageIndex   *PageIndex
	NextHandler http.Handler
}

func (fc *FileCombiner) SetNextHandler(next http.Handler) ChainedHandler {
	fc.NextHandler = next
	return fc
}

// func (fc *FileCombiner) AddHandler(h http.Handler) StackableHandler {
// 	return NewStackableHandler(fc, h)
// }

// func (fc *FileCombiner) SetNextHandler(next http.Handler) ChainedHandler {
// 	fc.NextHandler = next
// 	return MakeChainedHandler(next)
// }

func CtxWithCombinedPath(ctx context.Context, pathType string, resourcePath string) context.Context {
	return context.WithValue(ctx, "combined_path_"+pathType, resourcePath)
}

func (fc *FileCombiner) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// two things:
	// * if it's a page then lookup PageMeta, examine includes, work out whatever ID to give it and attach to context
	// * if it's for the combined file, then look it up and return it

	ctx := r.Context()
	ctx = CtxWithCombinedPath(ctx, "css", "/path/to/combined.css")

	if fc.NextHandler == nil {
		panic(fmt.Errorf("FileCombiner requires a NextHandler"))
	}
	fc.NextHandler.ServeHTTP(w, r.WithContext(ctx))
	// UpdateChainWR(w, r.WithContext(ctx))

}

// fc := &FileCombiner

// h = fc
// h.AddHandler

// func f() {

// 	h := 0

// 	fc := &FileCombiner{}

// 	h = h.SetNextHandler(fc)

// }

// func MakeChainedHandler(h http.Handler) ChainedHandler {

// 	// if h is a ChainedHandler, return it
// 	if ch, ok := h.(ChainedHandler); ok {
// 		return ch
// 	}

// 	// if not then
// 	return
// }
