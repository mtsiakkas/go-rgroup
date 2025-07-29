package rgroup

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerMiddleware(t *testing.T) {

	h := Handler(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	})

	m := func(h Handler) Handler {
		return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			res, _ := h(w, req)
			resm := Response(res.Data.(string) + ": middleware")
			return resm, nil
		}
	}

	hm := h.applyMiddleware([]Middleware{m})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	res, err := hm(rr, req)
	if err != nil {
		t.Logf("unexpected error: %s", err)
		t.Fail()
	}

	if d, ok := res.Data.(string); !ok || d != "test: middleware" {
		t.Logf("unexpected response: %s", d)
		t.Fail()
	}
}

func TestToHandlerFunc(t *testing.T) {

	h := Handler(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	})

	srv := httptest.NewServer(h.ToHandlerFunc())

	client := srv.Client()
	res, err := client.Get(srv.URL)
	if err != nil {
		t.Logf("unexpected error: %s", err)
		t.Fail()
	}

	d, err := io.ReadAll(res.Body)
	if err != nil {
		t.Logf("unexpected error: %s", err)
		t.Fail()
	}

	if string(d) != "test" {
		t.Logf("unexpected response: %s", d)
		t.Fail()
	}
}
