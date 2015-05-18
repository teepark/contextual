/*
The router package implements wrappers around
github.com/julienschmidt/httprouter.Router to enable the use of
golang.org/x/net/context.Context in endpoint handlers.
*/
package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

const paramKey = "params"

// InitFunc is a function to transform the Context before an endpoint is called.
// To short-circuit and skip the handlers entirely, return a Context that is
// already done (probably by cancelling). Be sure you also set an appropriate
// response code to the ResponseWriter.
type InitFunc func(context.Context, http.ResponseWriter, *http.Request) context.Context

// Router wraps an httprouter.Router, then for each request adds the
// httprouter's Params as the Value("params") to the context, executes a custom
// initialization function, and passes the context on to handlers that accept
// it as an argument.
type Router struct {
	router *httprouter.Router
	base context.Context
	init InitFunc
}

// NewRouter creates a new Router around a given httprouter.Router.
// All arguments may be nil, in which case the Router would wrap a Router
// created with httprouter.New(), the base context would be context.Background(),
// and the adapter would not perform any initialization of the context.
func NewRouter(router *httprouter.Router, base context.Context, init InitFunc) *Router {
	if router == nil {
		router = httprouter.New()
	}
	if base == nil {
		base = context.Background()
	}

	return &Router{router, base, init}
}

// ServeHTTP implements http.Handler
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}

// Handle adds a method/path handler with a context.Context argument
func (r *Router) Handle(method, path string, handle contextual.Handler) {
	r.router.Handle(method, path, handlerShim(r, handle))
}

// GET is a shortcut for Handle("GET", ...)
func (r *Router) GET(path string, handle contextual.Handler) {
	r.Handle("GET", path, handle)
}

// HEAD is a shortcut for Handle("HEAD", ...)
func (r *Router) HEAD(path string, handle contextual.Handler) {
	r.Handle("HEAD", path, handle)
}

// POST is a shortcut for Handle("POST", ...)
func (r *Router) POST(path string, handle contextual.Handler) {
	r.Handle("POST", path, handle)
}

// PUT is a shortcut for Handle("PUT", ...)
func (r *Router) PUT(path string, handle contextual.Handler) {
	r.Handle("PUT", path, handle)
}

// DELETE is a shortcut for Handle("DELETE", ...)
func (r *Router) DELETE(path string, handle contextual.Handler) {
	r.Handle("DELETE", path, handle)
}

// OPTIONS is a shortcut for Handle("OPTIONS", ...)
func (r *Router) OPTIONS(path string, handle contextual.Handler) {
	r.Handle("OPTIONS", path, handle)
}

// PATCH is a shortcut for Handle("PATCH", ...)
func (r *Router) PATCH(path string, handle contextual.Handler) {
	r.Handle("PATCH", path, handle)
}

func handlerShim(router *Router, ctxHandle contextual.Handler) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := context.WithValue(router.base, paramKey, p)

		if router.init != nil {
			ctx = router.init(ctx, w, r)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		ctxHandle.Serve(ctx, w, r)
	})
}

// Params retrieves the httprouter Params from a context
func Params(c context.Context) httprouter.Params {
	p := c.Value(paramKey)
	if p == nil {
		return nil
	}
	return p.(httprouter.Params)
}
