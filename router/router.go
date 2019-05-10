package router

import (
	"fmt"
	"net/http"
	"strings"
)

//Router Router
type Router struct {
	Routes          []*Route
	NotFoundHandler http.Handler
	PanicHandler    PanicHandler
}

//Route ...
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

//PanicHandler ...
type PanicHandler func(http.ResponseWriter, *http.Request)

//New returns a new router
func New(NotFoundHandler http.Handler, PanicHandler PanicHandler) *Router {

	var emptyRoutes []*Route

	return &Router{
		emptyRoutes,
		NotFoundHandler,
		PanicHandler,
	}
}

//RegisterRoute ...
func (r *Router) RegisterRoute(method, path string, handler http.HandlerFunc) {
	r.Routes = append(r.Routes, &Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

//rcvr Recover function
func (r *Router) rcvr(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req)
	}
}

func (r *Router) QueryParams(req *http.Request) (ret map[string]string) {

	values := req.URL.Query()

	for k, v := range values{

		fmt.Println(k)

		for i := range v{

			fmt.Println(v[i])
		}

	}

	return

}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if r.PanicHandler != nil {
		defer r.rcvr(w, req)
	}

	for _, route := range r.Routes {
		if match(req, route) {
			route.Handler.ServeHTTP(w, req)
			return
		}
	}

	w.WriteHeader(404)
	_, _ = w.Write([]byte("Not Found"))
	return
	//r.NotFoundHandler.ServeHTTP(w, req)
	//return
}

func match(req *http.Request, route *Route) bool {

	//if the method isn't the same, return immediately

	// e.g. POST == POST
	requestMethod := strings.ToUpper(req.Method)
	routeMethod := strings.ToUpper(route.Method)

	if requestMethod != routeMethod {
		return false
	}

	requestPath := strings.TrimRight(req.URL.Path, "/")
	routePath := strings.TrimRight(route.Path, "/")

	// e.g. "/api/users == /api/users"
	if routePath == requestPath {
		return true
	}

	requestPathParts := strings.SplitN(requestPath, "/", -1)
	routePathParts := strings.SplitN(routePath, "/", -1)

	if len(requestPathParts) != len(routePathParts) {
		return false
	}

	matches := true

	for i := range routePathParts {

		if !matches { //no need to continue
			break
		}

		if strings.HasPrefix(routePathParts[i], ":") {
			continue
		}

		matches = requestPathParts[i] == routePathParts[i]

	}

	// e.g. /api/user/:id == /api/user/123
	return matches

}
