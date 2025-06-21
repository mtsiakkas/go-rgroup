package rgroup_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
	"github.com/mtsiakkas/go-rgroup/pkg/config"
	"github.com/mtsiakkas/go-rgroup/pkg/group"
	"github.com/mtsiakkas/go-rgroup/pkg/herror"
	"github.com/mtsiakkas/go-rgroup/pkg/response"
)

func TestError(t *testing.T) {
	e := rgroup.Error(http.StatusInsufficientStorage).WithCode("test code").WithResponse("test response").WithMessage("test message")

	if !reflect.DeepEqual(e, &herror.HandlerError{
		LogMessage: "test message",
		Response:   "test response",
		ErrorCode:  "test code",
		HttpStatus: http.StatusInsufficientStorage,
	}) {
		t.Fail()
	}
}

func TestResponse(t *testing.T) {
	r := rgroup.Response("test response").WithMessage("test message").WithHttpStatus(http.StatusAccepted)

	if !reflect.DeepEqual(r, &response.HandlerResponse{
		LogMessage: "test message",
		Data:       "test response",
		HttpStatus: http.StatusAccepted,
	}) {
		t.Fail()
	}
}

func TestHandler(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		h := rgroup.New().Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Logf("unexpected status: %s", http.StatusText(rr.Code))
			t.Fail()
		}
	})

	t.Run("new with handlers", func(t *testing.T) {

		t.Run("success", func(t *testing.T) {
			g := rgroup.NewWithHandlers(map[string]group.Handler{
				"GET": func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
					return rgroup.Response("test").WithHttpStatus(http.StatusAccepted), nil
				},
			})

			if g == nil {
				t.Log("nil group ptr")
				t.FailNow()
			}

			h := g.Make()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			h(rr, req)

			if rr.Code != http.StatusAccepted {
				t.Logf("unexpected status: %s", http.StatusText(rr.Code))
				t.Fail()
			}

			b, err := io.ReadAll(rr.Result().Body)
			if err != nil {
				t.Logf("failed to read response body: %s", err)
				t.FailNow()
			}

			if string(b) != "test" {
				t.Logf("unexpected response: %s", string(b))
				t.Fail()
			}
		})

		t.Run("options", func(t *testing.T) {
			t.Run("panic", func(t *testing.T) {
				defer func() { _ = recover() }()
				rgroup.NewWithHandlers(map[string]group.Handler{
					http.MethodOptions: func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
						return nil, nil
					},
				})

				t.Log("should panic")
				t.Fail()
			})
			t.Run("overwrite", func(t *testing.T) {
				if err := config.OnOptionsHandler(config.OptionsHandlerOverwrite); err != nil {
					t.Logf("unexpected error: %s", err)
					t.FailNow()
				}
				h := rgroup.NewWithHandlers(map[string]group.Handler{
					http.MethodOptions: func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
						return rgroup.Response("test overwrite"), nil
					},
				}).Make()

				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodOptions, "/", nil)

				h(rr, req)

				b, err := io.ReadAll(rr.Result().Body)
				if err != nil {
					t.Logf("failed to read response body: %s", err)
					t.FailNow()
				}

				if string(b) != "test overwrite" {
					t.Logf("unexpected response: %s", string(b))
					t.Fail()
				}
			})
			t.Run("ignore", func(t *testing.T) {
				if err := config.OnOptionsHandler(config.OptionsHandlerIgnore); err != nil {
					t.Logf("unexpected error: %s", err)
					t.FailNow()
				}
				h := rgroup.NewWithHandlers(map[string]group.Handler{
					http.MethodGet: func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
						return nil, nil
					},
					http.MethodOptions: func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
						return rgroup.Response("test ignore"), nil
					},
				}).Make()

				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodOptions, "/", nil)

				h(rr, req)

				b, err := io.ReadAll(rr.Result().Body)
				if err != nil {
					t.Logf("failed to read response body: %s", err)
					t.FailNow()
				}

				if string(b) == "test ignore" {
					t.Logf("unexpected response: %s", string(b))
					t.Fail()
				}
			})
		})
	})
}
