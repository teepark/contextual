package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

func tagMiddleware(tag string) Middleware {
	return Inbound(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		fmt.Fprint(w, tag)
		return ctx
	})
}

func tagApp(tag string) contextual.Handler {
	return contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, tag)
	})
}

func contextValueApp(tag string) contextual.Handler {
	return contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, ctx.Value(tag))
	})
}

func contextValueMiddleware(tag string) Middleware {
	return Inbound(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		fmt.Fprint(w, ctx.Value(tag))
		return ctx
	})
}

func runHandler(h contextual.Handler, r *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.Serve(context.Background(), rr, r)
	return rr
}

func bodyOf(h contextual.Handler) (string, error) {
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return "", err
	}
	return runHandler(h, r).Body.String(), nil
}

func TestThenWorksWithNoMiddleware(t *testing.T) {
	handler := Chain{}.Then(tagApp("simple"))

	body, err := bodyOf(handler)
	if err != nil {
		t.Fatal(err)
	}

	if body != "simple" {
		t.Fatalf("expected 'simple', got '%s'", body)
	}
}

func TestChainInboundOrder(t *testing.T) {
	chain := Chain{
		tagMiddleware("m1\n"),
		tagMiddleware("m2\n"),
		tagMiddleware("m3\n"),
	}
	handler := chain.Then(tagApp("endpoint"))

	body, err := bodyOf(handler)
	if err != nil {
		t.Fatal(err)
	}

	expected := "m1\nm2\nm3\nendpoint"
	if body != expected {
		t.Fatalf("expected '%q', got '%q'", expected, body)
	}
}

func TestChainOutboundOrder(t *testing.T) {
	sl := make([]int, 0, 3)

	m1 := Outbound(func(ctx context.Context, r *http.Request) {
		sl = append(sl, 1)
	})
	m2 := Outbound(func(ctx context.Context, r *http.Request) {
		sl = append(sl, 2)
	})
	m3 := Outbound(func(ctx context.Context, r *http.Request) {
		sl = append(sl, 3)
	})

	ch := Chain{m1, m2, m3}
	handler := ch.Then(nil)

	req, err := http.NewRequest("GET", "http://localhost/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	runHandler(handler, req)

	if len(sl) != 3 ||
		sl[0] != 3 ||
		sl[1] != 2 ||
		sl[2] != 1 {
		t.Fatalf("expected [3, 2, 1], got %v", sl)
	}
}

func TestNilTreatedAsDefault(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "foo handler")
	}))
	mux.HandleFunc("/bar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "bar handler")
	}))

	trueDefault := http.DefaultServeMux
	http.DefaultServeMux = mux
	defer func() {
		http.DefaultServeMux = trueDefault
	}()

	mware := Inbound(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		return context.WithValue(ctx, "key", "value\n")
	})

	chain := Chain{
		mware,
		contextValueMiddleware("key"),
	}
	handler := chain.Then(nil)

	req, err := http.NewRequest("GET", "http://localhost/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := "value\nfoo handler"
	rr := runHandler(handler, req)
	got := rr.Body.String()
	if expected != got {
		t.Fatalf("expected '%s', got '%s'", expected, got)
	}
}
