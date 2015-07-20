// Package middleware provides a concept of middleware built around Inbound and
// Outbound callbacks, and some conveniences such as a Chain type.
package middleware

import (
	"net/http"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

// Middleware is a function that transforms one Handler into another,
// possibly adding behavior before or after the inner Handler.
type Middleware func(contextual.Handler) contextual.Handler

// Chain is an ordered collection of middlewares. The first middleware in
// the chain will be the outermost: it starts first and finishes last.
type Chain []Middleware

// Then produces a usable contextual.Handler from
// the middleware chain plus an endpoint Handler.
func (c Chain) Then(h contextual.Handler) contextual.Handler {
	if h == nil {
		h = defaultHandler
	}

	for i := len(c) - 1; i >= 0; i-- {
		h = c[i](h)
	}

	return h
}

// ThenFunc performs the same operation as Then, but takes a contextual.HandlerFunc
// directly.
//  c.Then(contextual.HandlerFunc(f))
//  c.ThenFunc(f)
// The above are equivalent statements.
func (c Chain) ThenFunc(f contextual.HandlerFunc) contextual.Handler {
	return c.Then(f)
}

var defaultHandler = contextual.HandlerFunc(func(_ context.Context, w http.ResponseWriter, r *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, r)
})
