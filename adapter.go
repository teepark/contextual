package contextual

import (
	"net/http"

	"golang.org/x/net/context"
)

// Adapter serves as an http.Handler, but does so by delegating the request
// handling to a contextual.Handler given a specific base context.Context.
type Adapter struct {
	c       context.Context
	handler Handler
}

// NewAdapter creates an Adapter (valid http.Handler) from a
// contextual.Handler and a base context
func NewAdapter(h Handler, c context.Context) *Adapter {
	if c == nil {
		c = context.Background()
	}
	return &Adapter{c, h}
}

// ServeHTTP implements the http.Handler interface
func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handler.Serve(a.c, w, r)
}
