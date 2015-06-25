package chain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

func tagTransformer(tag string) Transformer {
	return func(ctx context.Context, w http.ResponseWriter, _ *http.Request) context.Context {
		fmt.Fprint(w, tag)
		return ctx
	}
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

func contextValueTransformer(tag string) Transformer {
	return func(ctx context.Context, w http.ResponseWriter, _ *http.Request) context.Context {
		fmt.Fprint(w, ctx.Value(tag))
		return ctx
	}
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

func TestThenWorksWithNoTransformer(t *testing.T) {
	handler := Chain{}.Then(tagApp("simple"))

	body, err := bodyOf(handler)
	if err != nil {
		t.Fatal(err)
	}

	if body != "simple" {
		t.Fatalf("expected 'simple', got '%s'", body)
	}
}

func TestChainOrder(t *testing.T) {
	chain := Chain{
		tagTransformer("m1\n"),
		tagTransformer("m2\n"),
		tagTransformer("m3\n"),
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

	mware := func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		return context.WithValue(ctx, "key", "value\n")
	}

	chain := Chain{
		mware,
		contextValueTransformer("key"),
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
