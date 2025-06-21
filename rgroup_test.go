package rgroup_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mtsiakkas/go-rgroup"
)

func TestResponse(t *testing.T) {
	r := rgroup.Response(nil)
	if r.Data != nil {
		t.Log("r.Data not nil")
		t.Fail()
	}

	if r.LogMessage != "" {
		t.Logf("r.LogMessage not empty: %s", r.LogMessage)
		t.Fail()
	}

	r.WithMessage("test %s", "message")

	if r.LogMessage != "test message" {
		t.Logf("unexpected r.LogMessage: expected \"test message\" got \"%s\"", r.LogMessage)
		t.Fail()
	}

	if r.HttpStatus != http.StatusOK {
		t.Logf("unexpected r.HttpStatus value: expected \"%d\" got \"%d\"", http.StatusOK, r.HttpStatus)
		t.Fail()
	}

	r.WithHttpStatus(http.StatusAccepted)

	if r.HttpStatus != http.StatusAccepted {
		t.Logf("unexpected r.HttpStatus value: expected \"%d\" got \"%d\"", http.StatusAccepted, r.HttpStatus)
		t.Fail()
	}
}

func TestHandler(t *testing.T) {
	t.Run("new with handlers", func(t *testing.T) {
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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
				return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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
				return rgroup.Response("GET").WithHttpStatus(http.StatusAccepted), nil
			})

		_ = g.Post(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("POST").WithHttpStatus(http.StatusAccepted), nil
			})

		_ = g.Patch(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("PATCH").WithHttpStatus(http.StatusAccepted), nil
			})

		_ = g.Put(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("PUT").WithHttpStatus(http.StatusAccepted), nil
			})

		_ = g.Delete(
			func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("DELETE").WithHttpStatus(http.StatusAccepted), nil
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
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				o := struct {
					Data string `json:"data"`
				}{
					Data: "test",
				}
				return rgroup.Response(o).WithHttpStatus(http.StatusCreated), nil
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
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response(b).WithHttpStatus(http.StatusCreated), nil
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

	t.Run("overwrite options", func(t *testing.T) {
		t.Run("panic", func(t *testing.T) {
			defer func() { _ = recover() }()

			_ = rgroup.NewWithHandlers(map[string]rgroup.Handler{
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

			rgroup.OnOptionsHandler(rgroup.OptionsHandlerOverwrite)
			h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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

		t.Run("duplicate handler", func(t *testing.T) {

			t.Run("panic", func(t *testing.T) {
				defer func() { _ = recover() }()

				g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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
				rgroup.OnDuplicateMethod(rgroup.DuplicateMethodError)
				g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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
				rgroup.OnDuplicateMethod(rgroup.DuplicateMethodIgnore)
				g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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
					t.Logf("DuplicateHandlerIgnore unexpected output: got \"%s\"", string(b))
					t.Fail()
				}

			})

			t.Run("overwrite", func(t *testing.T) {
				rgroup.OnDuplicateMethod(rgroup.DuplicateMethodOverwrite)
				g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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
					t.Logf("DuplicateHandlerOverwrite unexpected output: got \"%s\"", string(b))
					t.Fail()
				}
			})

		})

		t.Run("ignore", func(t *testing.T) {

			rgroup.OnOptionsHandler(rgroup.OptionsHandlerIgnore)
			h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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

	t.Run("fail", func(t *testing.T) {
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
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

func TestOptions(t *testing.T) {
	h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
		"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
		},
		"POST": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
			return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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

func TestError(t *testing.T) {
	e := rgroup.Error(http.StatusInternalServerError)
	if !reflect.DeepEqual(rgroup.HandlerError{
		LogMessage: "",
		ErrorCode:  "",
		HttpStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("rgroup.Error() failed")
		t.Fail()
	}

	_ = e.WithMessage("test error: %s", "test message")
	if !reflect.DeepEqual(rgroup.HandlerError{
		LogMessage: "test error: test message",
		ErrorCode:  "",
		HttpStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithMessage failed")
		t.Fail()
	}

	_ = e.WithResponse("test error: %s", "test response")
	if !reflect.DeepEqual(rgroup.HandlerError{
		LogMessage: "test error: test message",
		Response:   "test error: test response",
		ErrorCode:  "",
		HttpStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithResponse failed")
		t.Fail()
	}

	if e.Error() != "test error: test message" {
		t.Log("HandlerError.Error() failed")
		t.Fail()
	}
	_ = e.WithCode("TEST_ERR")
	if !reflect.DeepEqual(rgroup.HandlerError{
		LogMessage: "test error: test message",
		Response:   "test error: test response",
		ErrorCode:  "TEST_ERR",
		HttpStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithCode failed")
		t.Fail()
	}
}

func TestPostprocessor(t *testing.T) {
	t.Run("print", func(t *testing.T) {
		h := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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
		rgroup.SetGlobalPostprocessor(print)

		g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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

		g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				time.Sleep(1 * time.Second)
				return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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
			c := r.Context.Value(ContextKey("test")).(string)
			fmt.Println("request complete: " + c)
		}

		g := rgroup.NewWithHandlers(map[string]rgroup.Handler{
			"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
				return rgroup.Response("success").WithHttpStatus(http.StatusAccepted).WithMessage("test message"), nil
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

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, _ = io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	_ = writer.Close()
	return <-out
}
