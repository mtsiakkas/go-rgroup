package rgroup_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mtsiakkas/go-rgroup"
)

func TestHandler(t *testing.T) {

	t.Run("new with handlers", func(t *testing.T) {
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		res, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.FailNow()
		}
		if !reflect.DeepEqual(res, []byte("success")) {
			t.Logf("unexpected response: expected \"success\" got \"%s\"", res)
			t.Fail()
		}
		if rr.Code != http.StatusAccepted {
			t.Logf("unexpected status code: expected 202 got \"%d\"", rr.Code)
			t.Fail()
		}
	})

	t.Run("add handler", func(t *testing.T) {
		g := rgroup.New()

		_ = g.AddHandler(
			"GET", func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
			})

		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		res, err := io.ReadAll(rr.Body)

		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.FailNow()
		}
		if !reflect.DeepEqual(res, []byte("success")) {
			t.Logf("unexpected response: expected \"success\" got \"%s\"", res)
			t.Fail()
		}
		if rr.Code != http.StatusAccepted {
			t.Logf("unexpected status code: expected 202 got \"%d\"", rr.Code)
			t.Fail()
		}
	})

	t.Run("direct add handler", func(t *testing.T) {
		g := rgroup.New()

		_ = g.Get(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("GET").WithHTTPStatus(http.StatusAccepted), nil
			})

		_ = g.Post(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("POST").WithHTTPStatus(http.StatusAccepted), nil
			})

		_ = g.Patch(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("PATCH").WithHTTPStatus(http.StatusAccepted), nil
			})

		_ = g.Put(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("PUT").WithHTTPStatus(http.StatusAccepted), nil
			})

		_ = g.Delete(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("DELETE").WithHTTPStatus(http.StatusAccepted), nil
			})

		h := g.Make()

		for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete} {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(m, "/", nil)

			h(rr, req)

			res, err := io.ReadAll(rr.Body)

			if err != nil {
				t.Logf("unexpected error: %s", err)
				t.FailNow()
			}
			if !reflect.DeepEqual(res, []byte(m)) {
				t.Logf("unexpected response: expected \"%s\" got \"%s\"", m, res)
				t.Fail()
			}
			if rr.Code != http.StatusAccepted {
				t.Logf("unexpected status code: expected 202 got \"%d\"", rr.Code)
				t.Fail()
			}
		}
	})

	t.Run("struct response", func(t *testing.T) {
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				o := struct {
					Data string `json:"data"`
				}{
					Data: "test",
				}
				return rgroup.Response(o).WithHTTPStatus(http.StatusCreated), nil
			},
		}).Make()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		if rr.Code != http.StatusCreated {
			t.Logf("unexpected status: expected %d got %d", http.StatusCreated, rr.Code)
			t.Fail()
		}

		b, err := io.ReadAll(rr.Result().Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if string(b) != "{\"data\":\"test\"}" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}

	})

	t.Run("with bytes", func(t *testing.T) {
		b := []byte("test")
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response(b).WithHTTPStatus(http.StatusCreated), nil
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		res, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.FailNow()
		}

		if !reflect.DeepEqual(res, []byte("test")) {
			t.Logf("unexpected response: expected \"test\" got \"%s\"", res)
			t.Fail()
		}
		if rr.Code != http.StatusCreated {
			t.Logf("unexpected status code: expected 201 got \"%d\"", rr.Code)
			t.Fail()
		}
	})

	t.Run("error", func(t *testing.T) {
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return nil, rgroup.Error(http.StatusForbidden)
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		res, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.FailNow()
		}
		if !slices.Equal(res, []byte("\n")) {
			t.Logf("unexpected response: %s", res)
			t.Fail()
		}
		if rr.Code != http.StatusForbidden {
			t.Logf("unexpected status: %d", rr.Code)
			t.Fail()
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return nil, rgroup.Error(http.StatusForbidden)
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		h(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Logf("unexpected status code: expected 405 got %d", rr.Code)
			t.Fail()
		}
	})

	t.Run("generic err", func(t *testing.T) {
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return nil, errors.New("test error")
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Logf("unexpected status code: expected 500 got %d", rr.Code)
			t.Fail()
		}
	})
}

func TestGlobalSettings(t *testing.T) {
	t.Run("options handler", func(t *testing.T) {
		t.Run("panic", func(t *testing.T) {
			defer func() { _ = recover() }()

			_ = rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return nil, rgroup.Error(http.StatusForbidden)
				},
				"OPTIONS": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return nil, nil
				},
			}).Make()

			t.Log("should panic")
			t.Fail()
		})

		t.Run("overwrite", func(t *testing.T) {

			if err := rgroup.Config.OnOverwriteOptionsHandler(rgroup.OverwriteOptionsHandlerOverwrite); err != nil {
				t.Logf("unexpected error: %s", err)
				t.FailNow()
			}
			h := rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return nil, rgroup.Error(http.StatusForbidden)
				},
				"OPTIONS": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return rgroup.Response("overwrite"), nil
				},
			}).Make()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodOptions, "/", nil)

			h(rr, req)
			b, err := io.ReadAll(rr.Result().Body)
			if err != nil {
				t.Logf("failed to read response body: %s", err)
				t.Fail()
			}
			if string(b) != "overwrite" {
				t.Logf("OptionsHandlerBehaviour: Overwrite - failed: expected response \"overwrite\" got \"%s\"", string(b))
				t.Fail()
			}
		})

		t.Run("ignore", func(t *testing.T) {

			if err := rgroup.Config.OnOverwriteOptionsHandler(rgroup.OverwriteOptionsHandlerIgnore); err != nil {
				t.Logf("unexpected error: %s", err)
				t.FailNow()
			}
			h := rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return nil, rgroup.Error(http.StatusForbidden)
				},
				"OPTIONS": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return rgroup.Response("failed"), nil
				},
			}).Make()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodOptions, "/", nil)

			h(rr, req)
			b, err := io.ReadAll(rr.Result().Body)
			if err != nil {
				t.Logf("failed to read response body: %s", err)
				t.Fail()
			}
			if string(b) == "failed" {
				t.Logf("OptionsHandlerBehaviour: Ignore - failed: got \"%s\"", string(b))
				t.Fail()
			}
			allow := strings.Split(rr.Header().Get("Allow"), ",")
			if !slices.Contains(allow, http.MethodGet) || !slices.Contains(allow, http.MethodOptions) {
				t.Logf("unexpected allow header: %s", rr.Header().Get("Allow"))
				t.Fail()
			}
		})
	})

	t.Run("duplicate handler", func(t *testing.T) {
		if err := rgroup.Config.OnOverwriteMethod(rgroup.OverwriteMethodPanic); err != nil {
			t.Log(err)
			t.Fail()
		}
		t.Run("panic", func(t *testing.T) {

			defer func() { _ = recover() }()

			g := rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return nil, nil
				},
			})

			_ = g.Get(func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return nil, nil
			})

			t.Log("should panic")
			t.Fail()
		})

		t.Run("error", func(t *testing.T) {
			if err := rgroup.Config.OnOverwriteMethod(rgroup.OverwriteMethodError); err != nil {
				t.Logf("unexpected error: %s", err)
				t.FailNow()
			}
			g := rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return nil, nil
				},
			})

			err := g.Get(func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return nil, nil
			})

			if err == nil {
				t.Log("expected error")
				t.Fail()
			}
		})

		t.Run("ignore", func(t *testing.T) {
			if err := rgroup.Config.OnOverwriteMethod(rgroup.OverwriteMethodIgnore); err != nil {
				t.Logf("unexpected error: %s", err)
				t.FailNow()
			}
			g := rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return rgroup.Response("get1"), nil
				},
			})

			_ = g.Get(func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("get2"), nil
			})

			h := g.Make()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			h(rr, req)
			b, err := io.ReadAll(rr.Result().Body)
			if err != nil {
				t.Logf("failed to read response body: %s", err)
				t.Fail()
			}
			if string(b) != "get1" {
				t.Logf("OverwriteHandlerIgnore unexpected output: got \"%s\"", string(b))
				t.Fail()
			}

		})

		t.Run("overwrite", func(t *testing.T) {
			if err := rgroup.Config.OnOverwriteMethod(rgroup.OverwriteMethodAllow); err != nil {
				t.Logf("unexpected error: %s", err)
				t.FailNow()
			}
			g := rgroup.NewWithHandlers(rgroup.HandlerMap{
				"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
					return rgroup.Response("get1"), nil
				},
			})

			_ = g.Get(func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("get2"), nil
			})

			h := g.Make()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			h(rr, req)
			b, err := io.ReadAll(rr.Result().Body)
			if err != nil {
				t.Logf("failed to read response body: %s", err)
				t.Fail()
			}
			if string(b) != "get2" {
				t.Logf("OverwriteHandlerOverwrite unexpected output: got \"%s\"", string(b))
				t.Fail()
			}
		})

	})
}

func TestOptions(t *testing.T) {
	h := rgroup.NewWithHandlers(rgroup.HandlerMap{
		"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
		},
		"POST": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
		},
	}).Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	h(rr, req)

	allow := strings.Split(rr.Header().Get("Allow"), ",")
	slices.Sort(allow)
	if !slices.Equal([]string{"GET", "OPTIONS", "POST"}, allow) {
		t.Logf("unexpected allow header: %s", allow)
		t.Fail()
	}
}

func TestPostprocessor(t *testing.T) {
	t.Run("print", func(t *testing.T) {
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		out := captureOutput(func() { h(rr, req) })

		if !strings.Contains(out, "GET 202 / [") || !strings.HasSuffix(out, "test message\n") {
			t.Logf("unexpected output: \"%s\"", out)
			t.Fail()
		}
	})

	t.Run("global", func(t *testing.T) {
		print := func(ctx context.Context, r *rgroup.RequestData) {
			fmt.Println("global")
		}
		rgroup.Config.SetGlobalPostprocessor(print)

		g := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
			},
		})

		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		out := captureOutput(func() { h(rr, req) })

		if out != "global\n" {
			t.Logf("unexpected log: %s", out)
			t.Fail()
		}
	})

	t.Run("global + local", func(t *testing.T) {
		print := func(ctx context.Context, r *rgroup.RequestData) {
			fmt.Println("request complete")
		}

		g := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				time.Sleep(1 * time.Second)
				return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
			},
		})

		g.SetPostprocessor(print)

		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		out := captureOutput(func() { h(rr, req) })

		if out != "request complete\n" {
			t.Logf("unexpected log: %s", out)
			t.Fail()
		}
	})

	t.Run("request with context", func(t *testing.T) {

		type ContextKey string
		print := func(ctx context.Context, r *rgroup.RequestData) {
			c := r.Request.Context().Value(ContextKey("test")).(string)
			fmt.Println("request complete: " + c)
		}

		g := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHTTPStatus(http.StatusAccepted).WithMessage("test message"), nil
			},
		})

		g.SetPostprocessor(print)

		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(
			context.WithValue(context.Background(), ContextKey("test"), "context test"),
			http.MethodGet,
			"/",
			nil,
		)

		out := captureOutput(func() { h(rr, req) })

		if out != "request complete: context test\n" {
			t.Logf("unexpected log: %s", out)
			t.Fail()
		}

	})
}

func TestMiddleware(t *testing.T) {
	g := rgroup.NewWithHandlers(rgroup.HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
		return rgroup.Response("test"), nil
	}})

	g.AddMiddleware(func(h rgroup.Handler) rgroup.Handler {
		return func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			res, err := h(w, req)
			if err != nil {
				return nil, err
			}

			if _, ok := res.Data.(string); ok {
				return rgroup.Response(res.Data.(string) + ": middleware 1"), nil
			}
			return nil, nil
		}
	}).AddMiddleware(func(h rgroup.Handler) rgroup.Handler {
		return func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			res, err := h(w, req)
			if err != nil {
				return nil, err
			}

			if _, ok := res.Data.(string); ok {
				return rgroup.Response(res.Data.(string) + ": middleware 2"), nil
			}
			return nil, nil
		}
	})

	h := g.Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h(rr, req)

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Logf("unexpected error: %s", err)
		t.Fail()
	}

	if string(b) != "test: middleware 1: middleware 2" {
		t.Logf("unexpected response: %s", string(b))
		t.Fail()
	}
}

func TestEnvelope(t *testing.T) {
	rgroup.Config.SetEnvelopeResponse(true)
	t.Run("envelope response", func(t *testing.T) {

		h := rgroup.NewWithHandlers(rgroup.HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("test").WithHTTPStatus(http.StatusCreated), nil
		}}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if rr.Code != http.StatusOK {
			t.Logf("unexpected code: %d", rr.Code)
		}
		if string(b) != "{\"data\":\"test\",\"status\":{\"http_status\":201}}" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}

	})

	t.Run("forward status code", func(t *testing.T) {
		rgroup.Config.SetForwardHTTPStatus(true)

		h := rgroup.NewWithHandlers(rgroup.HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("test").WithHTTPStatus(http.StatusCreated), nil
		}}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if rr.Code != http.StatusCreated {
			t.Logf("unexpected code: %d", rr.Code)
		}
		if string(b) != "{\"data\":\"test\",\"status\":{\"http_status\":201}}" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}

	})

	t.Run("with bytes", func(t *testing.T) {

		b := []byte("test")
		h := rgroup.NewWithHandlers(rgroup.HandlerMap{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response(b).WithHTTPStatus(http.StatusCreated), nil
			},
		}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		res, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.FailNow()
		}

		if !reflect.DeepEqual(res, []byte("test")) {
			t.Logf("unexpected response: expected \"test\" got \"%s\"", res)
			t.Fail()
		}
		if rr.Code != http.StatusCreated {
			t.Logf("unexpected status code: expected 201 got \"%d\"", rr.Code)
			t.Fail()
		}
	})

	t.Run("forward message", func(t *testing.T) {
		rgroup.Config.SetForwardLogMessage(true)

		h := rgroup.NewWithHandlers(rgroup.HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("test").WithHTTPStatus(http.StatusCreated).WithMessage("test message"), nil
		}}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if rr.Code != http.StatusOK {
			t.Logf("unexpected code: %d", rr.Code)
		}

		if string(b) != "{\"data\":\"test\",\"status\":{\"http_status\":201,\"message\":\"test message\"}}" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}
	})

	t.Run("error", func(t *testing.T) {
		rgroup.Config.SetEnvelopeResponse(true)

		h := rgroup.NewWithHandlers(rgroup.HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return nil, rgroup.Error(http.StatusForbidden).WithMessage("test error")
		}}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		b, err := io.ReadAll(rr.Body)
		if err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}

		if rr.Code != http.StatusOK {
			t.Logf("unexpected code: %d", rr.Code)
		}

		if string(b) != "{\"status\":{\"http_status\":403,\"error\":\"test error\"}}" {
			t.Logf("unexpected response: %s", string(b))
			t.Fail()
		}
	})

	rgroup.Config.Reset()
}
