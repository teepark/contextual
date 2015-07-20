package router

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teepark/contextual"
	"github.com/teepark/contextual/middleware"
	"golang.org/x/net/context"
)

func TestRouterRoutes(t *testing.T) {
	r := New(nil, nil)
	r.Handle("GET", "/", contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "I'm a GET")
	}))

	r.Handle("POST", "/", contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "I'm a POST")
	}))

	r.Handle("GET", "/other/path", contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "another path")
	}))

	s := httptest.NewServer(r)
	defer s.Close()

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal("GET", err)
	}

	msg, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal("read get", err)
	}

	if string(msg) != "I'm a GET" {
		t.Fatal("mismatch on GET", string(msg))
	}

	resp, err = http.Post(s.URL, "text/plain", nil)
	if err != nil {
		t.Fatal("POST", err)
	}

	msg, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal("read get", err)
	}

	if string(msg) != "I'm a POST" {
		t.Fatal("mismatch on POST", string(msg))
	}

	resp, err = http.Get(s.URL + "/other/path")
	if err != nil {
		t.Fatal("GET", err)
	}

	msg, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal("read get", err)
	}

	if string(msg) != "another path" {
		t.Fatal("mismatch on GET /other/path", string(msg))
	}
}

type routerMethodSpec struct {
	name string
	f    func(*Router) func(string, contextual.Handler)
}

var routerMethodSpecs = []routerMethodSpec{
	{"GET", func(r *Router) func(string, contextual.Handler) { return r.GET }},
	{"HEAD", func(r *Router) func(string, contextual.Handler) { return r.HEAD }},
	{"POST", func(r *Router) func(string, contextual.Handler) { return r.POST }},
	{"PUT", func(r *Router) func(string, contextual.Handler) { return r.PUT }},
	{"DELETE", func(r *Router) func(string, contextual.Handler) { return r.DELETE }},
	{"OPTIONS", func(r *Router) func(string, contextual.Handler) { return r.OPTIONS }},
	{"PATCH", func(r *Router) func(string, contextual.Handler) { return r.PATCH }},
}

func TestRouterMethodFuncs(t *testing.T) {
	// TODO: useful comments in this function, lots of high-level indirection to explain
	r := New(nil, nil)

	handlerForMethod := func(method string) contextual.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if method == "HEAD" {
				w.WriteHeader(204)
			} else {
				io.WriteString(w, method)
			}
		}
	}

	for _, spec := range routerMethodSpecs {
		spec.f(r)("/"+spec.name, handlerForMethod(spec.name))
	}

	s := httptest.NewServer(r)
	defer s.Close()

	for _, spec := range routerMethodSpecs {
		request, err := http.NewRequest(spec.name, s.URL+"/"+spec.name, nil)
		if err != nil {
			t.Fatalf("failed to create request (%s): %v", spec.name, err)
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("client.Do (%s): %v", spec.name, err)
		}

		msg, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("read %s: %v", spec.name, err)
		}

		if spec.name == "HEAD" {
			if response.StatusCode != 204 {
				t.Fatal("HEAD response status not 204:", response.StatusCode)
			}
		} else if response.StatusCode != 200 {
			t.Fatalf("response status (%s): %d", spec.name, response.StatusCode)
		}

		if spec.name == "HEAD" {
			if string(msg) != "" {
				t.Fatal("received response body for HEAD:", string(msg))
			}
		} else if string(msg) != spec.name {
			t.Fatalf("mismatch on %s: '%s'", spec.name, string(msg))
		}
	}
}

func TestInboundMiddlewareRuns(t *testing.T) {
	initer := middleware.Middleware(func(h contextual.Handler) contextual.Handler {
		return contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			h.Serve(context.WithValue(ctx, "key", "value"), w, r)
		})
	})

	router := New(nil, initer)

	router.GET("/", contextual.HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		ival := c.Value("key")
		if ival == nil {
			t.Fatal("missing value")
		}
		if val, ok := ival.(string); !ok {
			t.Fatal("non-string value", ival)
		} else {
			io.WriteString(w, val)
		}
	}))

	s := httptest.NewServer(router)
	defer s.Close()

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal("GET", err)
	}

	msg, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal("read:", err)
	}

	if string(msg) != "value" {
		t.Fatal("mismatch:", string(msg))
	}
}
