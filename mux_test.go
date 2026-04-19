package rgroup

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxMiddleware(t *testing.T) {

	h := func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	}

	g := New()
	g.Get(h)

	m := func(h Handler) Handler {
		return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			res, _ := h(w, req)
			resm := Response(res.Data.(string) + ": middleware")
			return resm, nil
		}
	}

	mux := NewServeMux()
	mux.Handle("/", g)
	mux.AddMiddleware(m)
	mm := mux.Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mm.ServeHTTP(rr, req)

	if rr.Body.String() != "test: middleware" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}
}

func TestMuxAddHandlers(t *testing.T) {
	g1 := New()
	g1.Post(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("POST /g1"), nil
	})
	g1.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET /g1"), nil
	})

	g2 := New()
	g2.Post(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("POST /g2"), nil
	})
	g2.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET /g2"), nil
	})

	mux := NewServeMux()
	mux.Handle("/g1", g1)
	mux.Handle("/g2", g2)

	mm := mux.Make()

	type TestRoute struct {
		method string
		route  string
	}
	routes := []TestRoute{
		TestRoute{method: http.MethodPost, route: "/g1"},
	}

	for _, m := range routes {
		req := httptest.NewRequest(m.method, m.route, nil)
		rr := httptest.NewRecorder()

		mm.ServeHTTP(rr, req)
		if rr.Body.String() != m.method+" "+m.route {
			t.Logf("unexpected response: %s", rr.Body.String())
			t.Fail()
		}
	}
}
