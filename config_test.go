package rgroup

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGlobalLogger(t *testing.T) {
	t.Run("nil logger", func(t *testing.T) {
		Config.SetGlobalLogger(nil)

		h := NewWithHandlers(HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return nil, nil
		}}).Make()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := captureOutput(func() { h(rr, req) })
		if res != "" {
			t.Logf("unexpected log: %s", res)
			t.Fail()
		}

		Config.Reset()
	})

	t.Run("global", func(t *testing.T) {
		Config.SetGlobalLogger(func(req *LoggerData) {
			fmt.Println("global postprocessor")
		})

		h := NewWithHandlers(HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return nil, nil
		}}).Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := captureOutput(func() { h(rr, req) })
		if res != "global postprocessor\n" {
			t.Logf("unexpected log: %s", res)
			t.Fail()
		}

		Config.Reset()
	})

}

func TestLogOptions(t *testing.T) {
	Config.SetLogOptionsRequests(false)
	if config.logOptions {
		t.Log("expected Config.logOptions = false")
		t.Fail()
	}
}

func TestSetPrewriter(t *testing.T) {
	Config.SetPrewriter(func(r *http.Request, hr *HandlerResponse) *HandlerResponse {
		return Response(hr.Data).WithHTTPStatus(http.StatusAccepted)
	})

	if config.prewriter == nil {
		t.Logf("expected not nil prewriter")
		t.Fail()
	} else {
		r := config.prewriter(nil, Response("test"))
		if r.Data.(string) != "test" || r.HTTPStatus != http.StatusAccepted {
			t.Logf("unexpected reponse: %v", r)
			t.Fail()
		}
	}

}

func TestForwardErrorLog(t *testing.T) {
	if config.forwardErrorLog == true {
		t.Log("expected forwardErrorLog == false")
		t.Fail()
	}

	Config.SetForwardErrorLog(true)
	if config.forwardErrorLog != true {
		t.Log("expected forwardErrorLog == true")
		t.Fail()
	}

	Config.Reset()

}

func TestEnvelopeConfig(t *testing.T) {
	if config.envelope.enabled {
		t.Log("expected Config.envelopeResponse = nil")
		t.Fail()
	}

	Config.Envelope.Enable()
	if !config.envelope.enabled {
		t.Log("expected Config.envelopeResponse not nil")
		t.Fail()
	}

	Config.Envelope.SetForwardHTTPStatus(true)
	if !config.envelope.enabled {
		t.Log("expected Config.envelopeResponse not nil")
		t.Fail()
		if !config.envelope.forwardHTTPStatus {
			t.Log("expected Config.envelopeResponse.forwardHTTPStatus = true")
			t.Fail()
		}
	}

	Config.Envelope.SetForwardLogMessage(true)
	if !config.envelope.forwardLogMessage {
		t.Log("expected Config.envelopeResponse.forwardLogMessage = true")
		t.Fail()
	}

	Config.Envelope.Disable()
	if config.envelope.enabled {
		t.Log("expected Config.Envelope.enabled=false")
		t.Fail()
	}

	Config.Reset()
}

func TestLockOnMake(t *testing.T) {
	t.Run("lock", func(t *testing.T) {
		g := New()
		g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("test"), nil
		})
		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		b := rr.Body.String()
		if b != "test" {
			t.Logf("unexpected response: %s", b)
			t.Fail()
		}

		g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("changed"), nil
		})

		h2 := g.Make()

		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)

		h2(rr, req)

		b2 := rr.Body.String()
		if b2 != "test" {
			t.Logf("unexpected response: %s", b2)
			t.Fail()
		}
	})

	t.Run("unlock", func(t *testing.T) {
		Config.LockOnMake(false)

		g := New()
		g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("test"), nil
		})
		h := g.Make()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h(rr, req)

		b := rr.Body.String()
		if b != "test" {
			t.Logf("unexpected response: %s", b)
			t.Fail()
		}

		g.Get(func(w http.ResponseWriter, req *http.Request) (*HandlerResponse, error) {
			return Response("changed"), nil
		})

		h2 := g.Make()

		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/", nil)

		h2(rr, req)

		b2 := rr.Body.String()
		if b2 != "changed" {
			t.Logf("unexpected response: %s", b2)
			t.Fail()
		}
		config.lockOnMake = true
	})
}
