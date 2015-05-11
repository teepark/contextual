package contextual

import (
	"net/http"

	"golang.org/x/net/context"
)

// Handler is like an http.Handler, but also accepts a context.Context to use
type Handler interface {
	Serve(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc allows a func that can *be* Handler.Serve()
// to actually implement the Handler interface directly
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// Serve is the shim that lets a HandlerFunc implement Handler
func (h HandlerFunc) Serve(c context.Context, w http.ResponseWriter, r *http.Request) {
	h(c, w, r)
}
