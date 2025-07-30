package rgroup

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestMiddleware(t *testing.T) {

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

	hm := g.AddMiddleware(m).Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	hm(rr, req)

	if rr.Body.String() != "test: middleware" {
		t.Logf("unexpected response: %s", rr.Body.String())
		t.Fail()
	}
}

func TestAddHandlers(t *testing.T) {
	g := New()
	g.AddHandler("BATCH", func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("BATCH"), nil
	})
	g.Post(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) { return Response("POST"), nil })
	g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) { return Response("GET"), nil })
	g.Put(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) { return Response("PUT"), nil })
	g.Patch(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("PATCH"), nil
	})
	g.Delete(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("DELETE"), nil
	})

	h := g.Make()

	for _, m := range []string{"BATCH", "POST", "GET", "PUT", "PATCH", "DELETE"} {
		req := httptest.NewRequest(m, "/", nil)
		rr := httptest.NewRecorder()

		h(rr, req)
		if rr.Body.String() != m {
			t.Logf("unexpected response: %s", m)
			t.Fail()
		}
	}
}

func TestOptions(t *testing.T) {
	Config.lockOnMake = false
	defer func() { Config.lockOnMake = true }()

	g := New()
	g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) { return Response("GET"), nil })
	g.Post(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) { return Response("POST"), nil })

	opts := g.MethodsAllowed()
	if len(opts) != 3 {
		t.Logf("unexpected opts: %s", opts)
		t.Fail()
	}

	for _, m := range []string{http.MethodGet, http.MethodOptions, http.MethodPost} {
		if !slices.Contains(opts, m) {
			t.Logf("unexpected opts: %s", opts)
			t.Fail()
		}
	}

	h := g.Make()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	h(rr, req)

	if rr.Body.String() != "" {
		t.Logf("unexpected options response: %s", rr.Body.String())
		t.Fail()
	}

	opts = strings.Split(rr.Header().Get("Allow"), ",")
	if len(opts) != 3 {
		t.Logf("unexpected options header: %s", rr.Header().Get("Allow"))
		t.Fail()
	}

	for _, m := range []string{http.MethodGet, http.MethodOptions, http.MethodPost} {
		if !slices.Contains(opts, m) {
			t.Logf("unexpected opts: %s", opts)
			t.Fail()
		}
	}

	g.AddHandler(http.MethodOptions, func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("OPTIONS"), nil
	})

	rr.Body.Reset()
	h(rr, req)

	if rr.Body.String() != "OPTIONS" {
		t.Logf("unexpected options response: %s", rr.Body.String())
		t.Fail()
	}

	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	rr = httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
		t.Fail()
	}

	Config.SetLogOptionsRequests(false)
	res := captureOutput(func() { h(httptest.NewRecorder(), httptest.NewRequest(http.MethodOptions, "/", nil)) })
	if res != "" {
		t.Logf("unexpected log output: %s", res)
		t.Fail()
	}
}

func TestEmptyGroup(t *testing.T) {
	Config.lockOnMake = false
	defer func() { Config.lockOnMake = true }()

	g := HandlerGroup{}
	h := g.Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
		t.Fail()
	}

	opts := g.MethodsAllowed()
	if len(opts) != 1 || opts[0] != "OPTIONS" {
		t.Logf("unexpected opts: %s", opts)
		t.Fail()
	}

	g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) { return Response("test"), nil })
	h = g.Make()
	h(rr, req)

	if rr.Code == http.StatusMethodNotAllowed {
		t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
		t.Fail()
	}

	opts = g.MethodsAllowed()
	if len(opts) != 2 || !slices.Contains(opts, "OPTIONS") || !slices.Contains(opts, "GET") {
		t.Logf("unexpected opts: %s", opts)
		t.Fail()
	}
}

func TestGroupLogger(t *testing.T) {
	g := New()
	g.SetLogger(func(ld *LoggerData) {
		fmt.Printf("LOGGER: %s", ld.Message())
	})

	g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response(nil).WithMessage("test logger"), nil
	})

	h := g.Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	log := captureOutput(func() { h(rr, req) })
	if log != "LOGGER: test logger" {
		t.Logf("unexpected log output: %s", log)
		t.Fail()
	}
}

func TestGroupErrorResponse(t *testing.T) {
	t.Run("rgroup.Error", func(t *testing.T) {
		g := New()
		g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return nil, Error(http.StatusNotAcceptable)
		})
		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		if rr.Code != http.StatusNotAcceptable {
			t.Logf("unexpected status: %d (%s)", rr.Code, http.StatusText(rr.Code))
			t.Fail()
		}

	})

	t.Run("error", func(t *testing.T) {
		g := New()
		g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return nil, errors.New("test error")
		})
		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		log := captureErrorLog(func() { h(rr, req) })

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
	g := New()
	g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response("GET"), nil
	})
	g.Post(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
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
	Config.SetGlobalLogger(func(ld *LoggerData) { fmt.Println(ld.Message()) })
	Config.SetPrewriter(func(r *http.Request, hr *HandlerResponse) *HandlerResponse {
		return Response(hr.Data).WithMessage("test prewriter")
	})

	g := New()
	g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
		return Response(nil), nil
	})

	h := g.Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	log := captureOutput(func() { h(rr, req) })

	if log != "test prewriter\n" {
		t.Logf("unexpected message: %s", log)
		t.Fail()
	}
}
