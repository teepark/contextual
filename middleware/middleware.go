package middleware

import (
	"net/http"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

// when a Middleware in a Chain cancels the context, we stop processing any
// more and record how far we reached under this key so that we will only run
// the Outbounds corresponding to Inbounds which actually ran.
const reachedKey = "__middleware_reached"

// Middleware is a type that has methods for transforming a context on the way
// in, and running a function on the way out.
type Middleware interface {
	Inbound(context.Context, http.ResponseWriter, *http.Request) context.Context
	Outbound(context.Context, *http.Request)
}

// Inbound is a convenience wrapper so that a function matching the
// Middleware.Inbound signature can implement Middleware directly.
type Inbound func(context.Context, http.ResponseWriter, *http.Request) context.Context

// Inbound simply calls the function, so that it meets the Middleware
// interface.
func (inb Inbound) Inbound(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return inb(ctx, w, r)
}

// Outbound does nothing, just helps to implement the Middleware interface.
func (inb Inbound) Outbound(_ context.Context, _ *http.Request) {}

// Outbound is a convenience wrapper so that a function matching the
// Middleware.Outbound signature can implement Middleware directly.
type Outbound func(context.Context, *http.Request)

// Inbound does nothing to the context, its only purpose is to help meet the
// Middleware interface.
func (outb Outbound) Inbound(ctx context.Context, _ http.ResponseWriter, _ *http.Request) context.Context {
	return ctx
}

// Outbound simply calls the function, it is here to help implement the
// Middleware interface.
func (outb Outbound) Outbound(ctx context.Context, r *http.Request) {
	outb(ctx, r)
}

// Chain is an ordered collection of middlewares. The first middleware in the
// chain will be the outermost: its Inbound runs first, its Outbound last.
type Chain []Middleware

// Inbound implements Middleware.Inbound by running each contained Middleware's
// Inbound method in order.
func (c Chain) Inbound(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	var (
		i  int
		mw Middleware
	)
	for i, mw = range c {
		ctx = mw.Inbound(ctx, w, r)

		if canceled(ctx) {
			return context.WithValue(ctx, reachedKey, i)
		}
	}
	return ctx
}

// Outbound implements Middleware.Outbound by running each contained
// Middleware's Outbound method in reverse order.
func (c Chain) Outbound(ctx context.Context, r *http.Request) {
	start := len(c) - 1

	if canceled(ctx) {
		val := ctx.Value(reachedKey)
		if val != nil {
			start = ctx.Value(reachedKey).(int)
		}
	}

	for i := start; i >= 0; i-- {
		c[i].Outbound(ctx, r)
	}
}

// Then produces a usable contextual.Handler from the middleware chain
// plus the provided final terminating endpoint Handler.
func (c Chain) Then(h contextual.Handler) contextual.Handler {
	if h == nil {
		h = defaultHandler
	}

	return contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = c.Inbound(ctx, w, r)
		h.Serve(ctx, w, r)
		c.Outbound(ctx, r)
	})
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

func canceled(ctx context.Context) bool {
	return ctx.Err() == context.Canceled
}
