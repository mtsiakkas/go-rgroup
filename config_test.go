package rgroup

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGlobalLogger(t *testing.T) {
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
}

func TestLogOptions(t *testing.T) {
	Config.SetLogOptionsRequests(false)
	if Config.logOptions {
		t.Log("expected Config.logOptions = false")
		t.Fail()
	}
}

func TestSetPrewriter(t *testing.T) {
	Config.SetPrewriter(func(r *http.Request, hr *HandlerResponse) *HandlerResponse {
		return Response(hr.Data).WithHTTPStatus(http.StatusAccepted)
	})

	if Config.prewriter == nil {
		t.Logf("expected not nil prewriter")
		t.Fail()
	} else {
		r := Config.prewriter(nil, Response("test"))
		if r.Data.(string) != "test" || r.HTTPStatus != http.StatusAccepted {
			t.Logf("unexpected reponse: %v", r)
			t.Fail()
		}
	}

}

func TestEnvelopeConfig(t *testing.T) {
	if Config.envelopeResponse != nil {
		t.Log("expected Config.envelopeResponse = nil")
		t.Fail()
	}

	Config.SetEnvelopeResponse(true)
	if Config.envelopeResponse == nil {
		t.Log("expected Config.envelopeResponse not nil")
		t.Fail()
	}

	Config.SetEnvelopeResponse(false)
	if Config.envelopeResponse != nil {
		t.Log("expected Config.envelopeResponse = nil")
		t.Fail()
	}

	Config.SetForwardHTTPStatus(true)
	if Config.envelopeResponse == nil {
		t.Log("expected Config.envelopeResponse not nil")
		t.Fail()
		if !Config.envelopeResponse.forwardHTTPStatus {
			t.Log("expected Config.envelopeResponse.forwardHTTPStatus = true")
			t.Fail()
		}
	}

	Config.SetEnvelopeResponse(false)
	Config.SetForwardLogMessage(true)
	if Config.envelopeResponse == nil {
		t.Log("expected Config.envelopeResponse not nil")
		t.Fail()
		if !Config.envelopeResponse.forwardLogMessage {
			t.Log("expected Config.envelopeResponse.forwardLogMessage = true")
			t.Fail()
		}
	}

	Config.Reset()
}
