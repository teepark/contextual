// Package chain provides a convenient way to chain contextual Handlers.
// It is based on alice (github.com/justinas/alice).
package chain

import (
	"net/http"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

// Middleware is a function that wraps one contextual.Handler to create another,
// potentially performing operations before and/or after calling the wrapped
// Handler.
type Middleware func(contextual.Handler) contextual.Handler

// Chain is an ordered collection of middlewares. The first middleware in the
// chain will be the outermost -- the opposite of what you'd get by reducing
// from the beginning of the chain to the end.
type Chain []Middleware

// Append adds one or more additional middlewares to the end of a chain. The
// appended middlewares will be the ones invoked last/innermost (not like
// taking an existing handler and wrapping it).
func (c Chain) Append(middlewares ...Middleware) Chain {
	return append(c, middlewares...)
}

// Then produces a usable contextual.Handler from the middleware chain
// plus the provided final terminating endpoint Handler.
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
