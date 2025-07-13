package rgroup

import (
	"errors"
	"net/http"
	"testing"
)

func TestError(t *testing.T) {
	e := Error(http.StatusInternalServerError)
	if !errorCompare(HandlerError{
		LogMessage: "",
		HTTPStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("Error() failed")
		t.Fail()
	}

	_ = e.WithMessage("test error: %s", "test message")
	if !errorCompare(HandlerError{
		LogMessage: "test error: test message",
		HTTPStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithMessage failed")
		t.Fail()
	}

	_ = e.WithResponse("test error: %s", "test response")
	if !errorCompare(HandlerError{
		LogMessage: "test error: test message",
		Response:   "test error: test response",
		HTTPStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithResponse failed")
		t.Fail()
	}

	Config.SetForwardLogMessage(true)
	env := e.ToEnvelope()
	switch {
	case env.Status.Message == nil:
		t.Logf("expected env.Status.Message not nil")
		t.Fail()
	case *env.Status.Error != e.Error():
		t.Logf("unexpected error: %s", *env.Status.Error)
		t.Fail()
	case *env.Status.Message != e.Response:
		t.Logf("unexpected message: %s", *env.Status.Message)
		t.Fail()
	case env.Status.HTTPStatus != http.StatusInternalServerError:
		t.Logf("unexpected status code: %d", env.Status.HTTPStatus)
		t.Fail()
	}

	etmp := errors.New("error")
	e2 := Error(http.StatusInternalServerError).Wrap(etmp)
	if e2.Error() != "error" {
		t.Logf("unexpected error message: \"%s\"", e2)
		t.Fail()
	}
	e2.WithMessage("test")
	if e2.Error() != "test: error" {
		t.Logf("unexpected error message: \"%s\"", e2)
		t.Fail()
	}

	if e2.Unwrap() != etmp {
		t.Logf("unexpected error unwrap: \"%s\"", e2.Unwrap())
		t.Fail()
	}

	Config.Reset()

}

func errorCompare(e1 HandlerError, e2 HandlerError) bool {
	return e1.Response == e2.Response &&
		e1.HTTPStatus == e2.HTTPStatus &&
		e1.LogMessage == e2.LogMessage
}
