// Package chain provides a convenient way to chain functions that transform a
// net/context.Context.
package chain

import (
	"net/http"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

// Transformer is a function that modifies a context based on the contents of a
// *Request. It can optionally bail out early by returning a canceled context,
// and signal failure to the client via the ResponseWriter.
type Transformer func(context.Context, http.ResponseWriter, *http.Request) context.Context

// Chain is an ordered collection of Transformer functions.
type Chain []Transformer

// Apply runs all the transformers in a chain successively, returning the
// resulting Context. If any of the Transformers returned a canceled Context,
// this function stops at that point. The base Context arg may be nil, in which
// case context.Background() will be used.
func (ch Chain) Apply(base context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	ctx := base
	if ctx == nil {
		ctx = context.Background()
	}

	for _, t := range ch {
		ctx = t(ctx, w, r)

		if ctx.Err() == context.Canceled {
			break
		}
	}

	return ctx
}

// Then returns a contextual.Handler that will first transform the context by
// the full Chain, then call the argument Handler if the context hasn't been
// canceled by then.
func (ch Chain) Then(h contextual.Handler) contextual.Handler {
	if h == nil {
		h = defaultHandler
	}

	return contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = ch.Apply(ctx, w, r)

		if ctx.Err() != context.Canceled {
			h.Serve(ctx, w, r)
		}
	})
}

var defaultHandler = contextual.HandlerFunc(func(_ context.Context, w http.ResponseWriter, r *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, r)
})
