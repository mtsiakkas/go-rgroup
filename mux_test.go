package rgroup

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxMiddleware(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	h := func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	}

	g := New()
	g.Handle(http.MethodGet, h)

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

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mux.ServeHTTP(rr, req)

	if rr.Body.String() != "test: middleware" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}
}

func TestMuxAddHandlers(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	g1 := New()
	g1.Handle(http.MethodPost, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("POST /g1"), nil
	})
	g1.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET /g1"), nil
	})

	g2 := New()
	g2.Handle(http.MethodPost, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("POST /g2"), nil
	})
	g2.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET /g2"), nil
	})

	mux := NewServeMux()
	mux.Handle("/g1", g1)
	mux.Handle("/g2", g2)

	type TestRoute struct {
		method string
		route  string
	}
	routes := []TestRoute{
		{method: http.MethodPost, route: "/g1"},
		{method: http.MethodGet, route: "/g1"},
		{method: http.MethodPost, route: "/g2"},
		{method: http.MethodGet, route: "/g2"},
	}

	for _, m := range routes {
		req := httptest.NewRequest(m.method, m.route, nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)
		if rr.Body.String() != m.method+" "+m.route {
			t.Logf("unexpected response: %s", rr.Body.String())
			t.Fail()
		}
	}
}

func TestMuxNested(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	// an adapted http.Handler responds with the bytes it wrote, so the data is
	// formatted rather than asserted to a string
	m := func(h Handler) Handler {
		return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			res, _ := h(w, req)
			return Response(fmt.Sprintf("%s: middleware", res.Data)), nil
		}
	}

	g := New().Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("group"), nil
	})

	// a plain http.Handler is adapted, so the inherited middleware applies to it too
	plain := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("plain"))
	})

	sub := NewServeMux()
	sub.Handle("/group", g)
	sub.Handle("/plain", plain)

	mux := NewServeMux()
	mux.AddMiddleware(m)
	mux.Handle("/sub/", sub.SetPrefix("/sub"))

	for path, want := range map[string]string{
		"/sub/group": "group: middleware",
		"/sub/plain": "plain: middleware",
	} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))

		if rr.Body.String() != want {
			t.Logf("unexpected response for %s: %s", path, rr.Body.String())
			t.Fail()
		}
	}
}

func TestMuxPanics(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	g := New().Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	})

	tests := []struct {
		name  string
		panic string
		f     func(*HandlerMux)
	}{
		{
			name:  "nil handler",
			panic: "[rgroup.HandlerMux] nil Handler",
			f:     func(m *HandlerMux) { m.Handle("/", nil) },
		},
		{
			name:  "no path",
			panic: "[rgroup.HandlerMux] Handler without path",
			f:     func(m *HandlerMux) { m.Handle("", g) },
		},
		{
			name:  "duplicate handler",
			panic: "[rgroup.HandlerMux] duplicate Handler",
			f:     func(m *HandlerMux) { m.Handle("/", g).Handle("/", g) },
		},
		{
			name:  "nil middleware",
			panic: "[rgroup.HandlerMux] nil middleware",
			f:     func(m *HandlerMux) { m.AddMiddleware(nil) },
		},
		{
			name: "build after serve",
			f: func(m *HandlerMux) {
				m.Handle("/", g)
				m.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
				m.SetPrefix("/test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.panic
			if want == "" {
				want = "[rgroup.HandlerMux] build after serve"
			}

			defer func() {
				if r := recover(); r != want {
					t.Logf("expected panic %q, got: %v", want, r)
					t.Fail()
				}
			}()

			tt.f(NewServeMux())
		})
	}
}
