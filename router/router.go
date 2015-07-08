/*
Package router implements wrappers around
github.com/julienschmidt/httprouter.Router to enable the use of
golang.org/x/net/context.Context in endpoint handlers.

A simple example looks like this:

	package main

	import (
		"fmt"
		"log"
		"net/http"

		"github.com/teepark/contextual"
		"github.com/teepark/contextual/router"
		"golang.org/x/net/context"
	)

	func index(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome!")
	}

	func hello(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		params := router.Params(ctx)
		fmt.Fprintf(w, "Hello, %s!\n", params.ByName("name"))
	}

	func main() {
		router := router.New(nil, nil) // take defaults
		router.GET("/", contextual.HandlerFunc(index))
		router.GET("/hello/:name", contextual.HandlerFunc(hello))

		log.Fatal(http.ListenAndServe(":8080", router))
	}

*/
package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/teepark/contextual"
	"github.com/teepark/contextual/middleware"
	"golang.org/x/net/context"
)

// ParamKey is the key in a context.Context under which the httprouter.Params
// will be stored.
const ParamKey = "params"

// Router wraps an httprouter.Router to accept contextual.Handlers. For
// each request it adds the httprouter.Params to the context as
// Value(router.ParamKey), then passes it to the appropriate (optionally
// middleware-wrapped) handler.
type Router struct {
	router *httprouter.Router
	mw     middleware.Middleware
}

// New creates a new Router around a given httprouter.Router.
// All arguments may be nil, in which case the Router would wrap a
// Router created with httprouter.New() and there would be no
// middleware applied.
func New(router *httprouter.Router, mw middleware.Middleware) *Router {
	if router == nil {
		router = httprouter.New()
	}

	return &Router{
		router: router,
		mw: mw,
	}
}

// ServeHTTP implements http.Handler
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}

// ServeFiles serves files from a given filesystem root. The path must
// end with "/*filepath". See httprouter's documentation for details.
func (router *Router) ServeFiles(path string, root http.FileSystem) {
	router.router.ServeFiles(path, root)
}

// Handle adds a method/path handler with a context.Context argument
func (router *Router) Handle(method, path string, handle contextual.Handler) {
	router.router.Handle(method, path, handlerShim(router, handle))
}

// Handle adds an http.Handler for a method/path
func (router *Router) Handler(method, path string, handler http.Handler) {
	router.router.Handler(method, path, handler)
}

// handlerFunc adds an http.HandlerFunc for a method/path
func (router *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	router.router.HandlerFunc(method, path, handler)
}

// GET is a shortcut for Handle("GET", ...)
func (router *Router) GET(path string, handle contextual.Handler) {
	router.Handle("GET", path, handle)
}

// HEAD is a shortcut for Handle("HEAD", ...)
func (router *Router) HEAD(path string, handle contextual.Handler) {
	router.Handle("HEAD", path, handle)
}

// POST is a shortcut for Handle("POST", ...)
func (router *Router) POST(path string, handle contextual.Handler) {
	router.Handle("POST", path, handle)
}

// PUT is a shortcut for Handle("PUT", ...)
func (router *Router) PUT(path string, handle contextual.Handler) {
	router.Handle("PUT", path, handle)
}

// DELETE is a shortcut for Handle("DELETE", ...)
func (router *Router) DELETE(path string, handle contextual.Handler) {
	router.Handle("DELETE", path, handle)
}

// OPTIONS is a shortcut for Handle("OPTIONS", ...)
func (router *Router) OPTIONS(path string, handle contextual.Handler) {
	router.Handle("OPTIONS", path, handle)
}

// PATCH is a shortcut for Handle("PATCH", ...)
func (router *Router) PATCH(path string, handle contextual.Handler) {
	router.Handle("PATCH", path, handle)
}

func handlerShim(router *Router, ctxHandle contextual.Handler) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := context.WithValue(context.Background(), ParamKey, p)

		var handle contextual.Handler
		if router.mw != nil {
			handle = contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				ctx = router.mw.Inbound(ctx, w, r)
				ctxHandle.Serve(ctx, w, r)
				router.mw.Outbound(ctx, r)
			})
		} else {
			handle = ctxHandle
		}

		handle.Serve(ctx, w, r)
	})
}

// Params retrieves the httprouter Params from a context
func Params(c context.Context) httprouter.Params {
	p := c.Value(ParamKey)
	if p == nil {
		return nil
	}
	return p.(httprouter.Params)
}
