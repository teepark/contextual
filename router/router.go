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

// RouterAdapter wraps an httprouter.Router, then for each request adds the
// httprouter's Params as the Value("params") to the context, executes a custom
// initialization function, and passes the context on to handlers that accept
// it as an argument.
type RouterAdapter struct {
	// The Router that will be used to route incoming requests to handlers
	Router *httprouter.Router

	// The Context to use as a starting point for all endpoint handlers
	Base context.Context

	// A function to transform the Context before the endpoint is called. To
	// short-circuit and skip the handlers entirely, return a Context that is
	// already done (probably cancelled). Be sure you also send an appropriate
	// response code to the ResponseWriter.
	Init func(context.Context, http.ResponseWriter, *http.Request) context.Context
}

// NewRouterAdapter creates a new RouterAdapter around a given httprouter.Router.
// All arguments may be nil, in which case the RouterAdapter would wrap a Router
// created with httprouter.New(), the base context would be context.Background(),
// and the adapter would not perform any initialization of the context.
func NewRouterAdapter(router *httprouter.Router, base context.Context,
	init func(context.Context, http.ResponseWriter, *http.Request) context.Context) *RouterAdapter {
	if router == nil {
		router = httprouter.New()
	}
	if base == nil {
		base = context.Background()
	}

	return &RouterAdapter{router, base, init}
}

// ServeHTTP implements http.Handler
func (ra *RouterAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ra.Router.ServeHTTP(w, r)
}

// Handle adds a method/path handler with a context.Context argument
func (ra *RouterAdapter) Handle(method, path string, handle contextual.Handler) {
	ra.Router.Handle(method, path, handlerShim(ra, handle))
}

// GET is a shortcut for Handle("GET", ...)
func (ra *RouterAdapter) GET(path string, handle contextual.Handler) {
	ra.Handle("GET", path, handle)
}

// HEAD is a shortcut for Handle("HEAD", ...)
func (ra *RouterAdapter) HEAD(path string, handle contextual.Handler) {
	ra.Handle("HEAD", path, handle)
}

// POST is a shortcut for Handle("POST", ...)
func (ra *RouterAdapter) POST(path string, handle contextual.Handler) {
	ra.Handle("POST", path, handle)
}

// PUT is a shortcut for Handle("PUT", ...)
func (ra *RouterAdapter) PUT(path string, handle contextual.Handler) {
	ra.Handle("PUT", path, handle)
}

// DELETE is a shortcut for Handle("DELETE", ...)
func (ra *RouterAdapter) DELETE(path string, handle contextual.Handler) {
	ra.Handle("DELETE", path, handle)
}

// OPTIONS is a shortcut for Handle("OPTIONS", ...)
func (ra *RouterAdapter) OPTIONS(path string, handle contextual.Handler) {
	ra.Handle("OPTIONS", path, handle)
}

// PATCH is a shortcut for Handle("PATCH", ...)
func (ra *RouterAdapter) PATCH(path string, handle contextual.Handler) {
	ra.Handle("PATCH", path, handle)
}

func handlerShim(ra *RouterAdapter, ctxHandle contextual.Handler) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := context.WithValue(ra.Base, paramKey, p)

		if ra.Init != nil {
			ctx = ra.Init(ctx, w, r)
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
