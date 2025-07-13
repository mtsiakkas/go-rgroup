package rgroup

import (
	"net/http"
	"testing"
)

func TestResponse(t *testing.T) {
	r := Response(nil)
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

	if r.HTTPStatus != http.StatusOK {
		t.Logf("unexpected r.HttpStatus value: expected \"%d\" got \"%d\"", http.StatusOK, r.HTTPStatus)
		t.Fail()
	}

	r.WithHTTPStatus(http.StatusAccepted)

	if r.HTTPStatus != http.StatusAccepted {
		t.Logf("unexpected r.HttpStatus value: expected \"%d\" got \"%d\"", http.StatusAccepted, r.HTTPStatus)
		t.Fail()
	}

	Config.SetEnvelopeResponse(true)
	Config.SetForwardLogMessage(true)

	env := r.ToEnvelope()
	switch {
	case env.Data != nil:
		t.Log("expected nil data")
		t.Fail()
	case env.Status.HTTPStatus != http.StatusAccepted:
		t.Logf("unexpected status: %d (%s)", env.Status.HTTPStatus, http.StatusText(env.Status.HTTPStatus))
		t.Fail()
	case *env.Status.Message != r.LogMessage:
		t.Logf("unexpected status message: %s", *env.Status.Message)
		t.Fail()
	case env.Status.Error != nil:
		t.Logf("unexpected error: %s", *env.Status.Error)
		t.Fail()
	}

	Config.Reset()
}
