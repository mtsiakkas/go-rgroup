package rgroup

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverPanic(t *testing.T) {
	t.Run("recovered", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			panic("test_panic")
		})

		var logged *LoggerData
		g.SetLogger(func(l *LoggerData) { logged = l })

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		g.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}

		if logged == nil || logged.Error == nil {
			t.Log("expected the panic to reach the logger")
			t.FailNow()
		}

		var pe *PanicError
		if !errors.As(logged.Error, &pe) {
			t.Logf("expected a PanicError, got: %s", logged.Error)
			t.FailNow()
		}

		if pe.Value() != "test_panic" {
			t.Logf("unexpected panic value: %v", pe.Value())
			t.Fail()
		}

		if len(pe.Stack()) == 0 {
			t.Log("expected a stack trace")
			t.Fail()
		}
	})

	t.Run("middleware", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("test"), nil
		})
		g.AddMiddleware(func(h Handler) Handler {
			return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
				panic("test_panic")
			}
		})
		g.SetLogger(func(l *LoggerData) {})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		g.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}
	})

	t.Run("ToHandlerFunc", func(t *testing.T) {
		t.Cleanup(resetConfig)

		h := Handler(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			panic("test_panic")
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		captureErrorLog(func() { h.ToHandlerFunc()(rr, req) })

		if rr.Code != http.StatusInternalServerError {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}
	})

	t.Run("abort handler", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		defer func() {
			if r := recover(); r != http.ErrAbortHandler {
				t.Logf("expected http.ErrAbortHandler to propagate, got: %v", r)
				t.Fail()
			}
		}()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			panic(http.ErrAbortHandler)
		})

		g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})

	t.Run("disabled", func(t *testing.T) {
		t.Cleanup(resetConfig)

		defer func() {
			if r := recover(); r != "test_panic" {
				t.Logf("expected the panic to propagate, got: %v", r)
				t.Fail()
			}
		}()

		Config.SetRecoverPanics(false)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			panic("test_panic")
		})

		g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestMiddleware(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	h := func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	}

	m := func(h Handler) Handler {
		return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			res, _ := h(w, req)
			resm := Response(res.Data.(string) + ": middleware")
			return resm, nil
		}
	}

	g := New().Handle(http.MethodGet, h).AddMiddleware(m)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	g.ServeHTTP(rr, req)

	if rr.Body.String() != "test: middleware" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}

	// the group is frozen by the first request, so it cannot be built any further
	defer func() {
		if r := recover(); r != "[rgroup.HandlerGroup] build after serve" {
			t.Logf("expected a build after serve panic, got: %v", r)
			t.Fail()
		}
	}()

	g.AddMiddleware(m)
}

func TestMiddlewareOrder(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	m := func(name string) Middleware {
		return func(h Handler) Handler {
			return func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
				res, _ := h(w, req)
				return Response(res.Data.(string) + ": " + name), nil
			}
		}
	}

	g := New()
	// middleware added before the handler applies to it as well
	g.AddMiddleware(m("first"))
	g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	})
	g.AddMiddleware(m("second"))

	rr := httptest.NewRecorder()
	g.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Body.String() != "test: first: second" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}
}

func TestHandle(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	g := New()
	g.Handle("BATCH", func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("BATCH"), nil
	})
	g.Handle(http.MethodPost, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("POST"), nil
	})
	// the method is normalized to upper case
	g.Handle("get", func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET"), nil
	})
	g.Handle(http.MethodPut, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("PUT"), nil
	})
	g.Handle(http.MethodPatch, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("PATCH"), nil
	})
	g.Handle(http.MethodDelete, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("DELETE"), nil
	})

	for _, m := range []string{"BATCH", "POST", "GET", "PUT", "PATCH", "DELETE"} {
		req := httptest.NewRequest(m, "/", nil)
		rr := httptest.NewRecorder()

		g.ServeHTTP(rr, req)
		if rr.Body.String() != m {
			t.Logf("unexpected response: %s", rr.Body.String())
			t.Fail()
		}
	}
}

func TestHandlePanics(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	h := func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	}

	tests := []struct {
		name  string
		panic string
		f     func(*HandlerGroup)
	}{
		{
			name:  "nil handler",
			panic: "[rgroup.HandlerGroup] nil Handler",
			f:     func(g *HandlerGroup) { g.Handle(http.MethodGet, nil) },
		},
		{
			name:  "no method",
			panic: "[rgroup.HandlerGroup] Handler without method",
			f:     func(g *HandlerGroup) { g.Handle("", h) },
		},
		{
			name:  "duplicate handler",
			panic: "[rgroup.HandlerGroup] duplicate Handler",
			f:     func(g *HandlerGroup) { g.Handle(http.MethodGet, h).Handle("get", h) },
		},
		{
			name:  "nil middleware",
			panic: "[rgroup.HandlerGroup] nil middleware",
			f:     func(g *HandlerGroup) { g.AddMiddleware(nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != tt.panic {
					t.Logf("expected panic %q, got: %v", tt.panic, r)
					t.Fail()
				}
			}()

			tt.f(New())
		})
	}
}

func TestUnlockedConfig(t *testing.T) {
	t.Cleanup(resetConfig)

	defer func() {
		if r := recover(); r != "[rgroup] config not locked" {
			t.Logf("expected a config not locked panic, got: %v", r)
			t.Fail()
		}
	}()

	New()
}

func TestOptions(t *testing.T) {
	t.Run("default handler", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("GET"), nil
		})
		g.Handle(http.MethodPost, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("POST"), nil
		})

		if g.MethodsAllowed() != "GET,OPTIONS,POST" {
			t.Logf("unexpected methods allowed: %s", g.MethodsAllowed())
			t.Fail()
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/", nil)

		g.ServeHTTP(rr, req)

		if rr.Body.String() != "" {
			t.Logf("unexpected options response: %s", rr.Body.String())
			t.Fail()
		}

		if rr.Header().Get("Allow") != "GET,OPTIONS,POST" {
			t.Logf("unexpected options header: %s", rr.Header().Get("Allow"))
			t.Fail()
		}

		rr = httptest.NewRecorder()
		g.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, "/", nil))

		if rr.Code != http.StatusMethodNotAllowed {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}

		if rr.Header().Get("Allow") != "GET,OPTIONS,POST" {
			t.Logf("unexpected allow header: %s", rr.Header().Get("Allow"))
			t.Fail()
		}
	})

	t.Run("custom handler", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("GET"), nil
		})
		g.Handle(http.MethodOptions, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("OPTIONS"), nil
		})

		// a handled OPTIONS is not repeated in the allowed methods
		if g.MethodsAllowed() != "GET,OPTIONS" {
			t.Logf("unexpected methods allowed: %s", g.MethodsAllowed())
			t.Fail()
		}

		rr := httptest.NewRecorder()
		g.ServeHTTP(rr, httptest.NewRequest(http.MethodOptions, "/", nil))

		if rr.Body.String() != "OPTIONS" {
			t.Logf("unexpected options response: %s", rr.Body.String())
			t.Fail()
		}
	})

	t.Run("logging", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("GET"), nil
		})

		// OPTIONS requests are not logged by default
		res := captureOutput(func() {
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodOptions, "/", nil))
		})
		if res != "" {
			t.Logf("unexpected log output: %s", res)
			t.Fail()
		}
	})

	t.Run("logging enabled", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.SetLogOptionsRequests(true)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("GET"), nil
		})

		res := captureOutput(func() {
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodOptions, "/", nil))
		})
		if !strings.Contains(res, "OPTIONS") {
			t.Logf("unexpected log output: %s", res)
			t.Fail()
		}
	})
}

func TestEmptyGroup(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	g := New()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	g.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Logf("unexpected status: %s", rr.Result().Status)
		t.Fail()
	}

	if g.MethodsAllowed() != "OPTIONS" {
		t.Logf("unexpected methods allowed: %s", g.MethodsAllowed())
		t.Fail()
	}
}

func TestFrozenGroup(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	h := func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("test"), nil
	}

	tests := []struct {
		name string
		f    func(*HandlerGroup)
	}{
		{name: "Handle", f: func(g *HandlerGroup) { g.Handle(http.MethodPost, h) }},
		{name: "AddMiddleware", f: func(g *HandlerGroup) { g.AddMiddleware(func(h Handler) Handler { return h }) }},
		{name: "SetLogger", f: func(g *HandlerGroup) { g.SetLogger(func(l *LoggerData) {}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != "[rgroup.HandlerGroup] build after serve" {
					t.Logf("expected a build after serve panic, got: %v", r)
					t.Fail()
				}
			}()

			g := New().Handle(http.MethodGet, h)
			g.SetLogger(func(l *LoggerData) {})
			g.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			tt.f(g)
		})
	}
}

func TestGroupLogger(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	g := New()
	g.SetLogger(func(ld *LoggerData) {
		fmt.Printf("LOGGER: %s", ld.Message())
	})

	g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response(nil).WithMessage("test logger"), nil
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	log := captureOutput(func() { g.ServeHTTP(rr, req) })
	if log != "LOGGER: test logger" {
		t.Logf("unexpected log output: %s", log)
		t.Fail()
	}
}

func TestGroupErrorResponse(t *testing.T) {
	t.Run("rgroup.Error", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return nil, Error(http.StatusNotAcceptable)
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		g.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotAcceptable {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}

	})

	t.Run("error", func(t *testing.T) {
		t.Cleanup(resetConfig)
		Config.Lock()

		g := New()
		g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return nil, errors.New("test error")
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		log := captureErrorLog(func() { g.ServeHTTP(rr, req) })

		if rr.Code != http.StatusInternalServerError {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}
		if !strings.Contains(log, "test error") {
			t.Logf("unexpected error message: %s", log)
			t.Fail()
		}

	})
}

func TestNetHttpHandler(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.Lock()

	g := New()
	g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET"), nil
	})
	g.Handle(http.MethodPost, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("POST").WithHTTPStatus(http.StatusCreated), nil
	})

	srv := httptest.NewServer(g)
	defer srv.Close()

	client := srv.Client()
	resGet, err := client.Get(srv.URL)
	if err != nil {
		t.Logf("failed to call test server: %s", err)
		t.FailNow()
	}

	bodyGet, err := io.ReadAll(resGet.Body)
	resGet.Body.Close()

	if err != nil {
		t.Logf("failed to read response body")
		t.FailNow()
	}

	if string(bodyGet) != "GET" {
		t.Logf("unexpected response: %s", string(bodyGet))
		t.Fail()
	}

	resPost, err := client.Post(srv.URL, "", nil)
	if err != nil {
		t.Logf("failed to call test server: %s", err)
		t.FailNow()
	}

	bodyPost, err := io.ReadAll(resPost.Body)
	resPost.Body.Close()

	if err != nil {
		t.Logf("failed to read response body")
		t.FailNow()
	}

	if resPost.StatusCode != http.StatusCreated {
		t.Logf("unexpected status: %s", http.StatusText(resPost.StatusCode))
		t.Fail()
	}

	if string(bodyPost) != "POST" {
		t.Logf("unexpected response: %s", string(bodyPost))
		t.Fail()
	}
}

func TestGroupPrewriter(t *testing.T) {
	t.Cleanup(resetConfig)
	Config.SetGlobalLogger(func(ld *LoggerData) { fmt.Println(ld.Message()) })
	Config.SetPrewriter(func(r *http.Request, hr *HandlerResponse) *HandlerResponse {
		return Response(hr.Data).WithMessage("test prewriter")
	})
	Config.Lock()

	g := New()
	g.Handle(http.MethodGet, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response(nil), nil
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	log := captureOutput(func() { g.ServeHTTP(rr, req) })

	if log != "test prewriter\n" {
		t.Logf("unexpected message: %s", log)
		t.Fail()
	}
}
