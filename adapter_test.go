package contextual

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

func TestAdapterAdapts(t *testing.T) {
	greeting := "hello, world"
	h := HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, greeting)
	})
	a := NewAdapter(h, nil)
	s := httptest.NewServer(a)
	defer s.Close()

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal("GET", err)
	}

	hello, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal("read", err)
	}
	if string(hello) != greeting {
		t.Fatal("mismatch", string(hello))
	}
}

func TestBaseCtxPropogates(t *testing.T) {
	key, value := "testkey", "message value"
	base := context.WithValue(context.Background(), key, value)

	h := HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ival := ctx.Value(key)
		if ival == nil {
			t.Fatal("missing context value")
		}
		val, ok := ival.(string)
		if !ok {
			t.Fatal("wrong type for value", ival)
		}
		io.WriteString(w, val)
	})
	a := NewAdapter(h, base)
	s := httptest.NewServer(a)
	defer s.Close()

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Fatal("GET", err)
	}

	msg, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal("read", err)
	}
	if string(msg) != value {
		t.Fatal("mismatch", string(msg))
	}
}
