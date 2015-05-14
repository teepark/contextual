package router

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teepark/contextual"
	"golang.org/x/net/context"
)

func TestRouterRoutes(t *testing.T) {
	r := NewRouter(nil, nil, nil)
	r.GET("/", contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "I'm a GET")
	}))

	r.POST("/", contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "I'm a POST")
	}))

	r.GET("/other/path", contextual.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
