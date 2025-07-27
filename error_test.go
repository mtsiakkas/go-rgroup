package rgroup

import (
	"errors"
	"fmt"
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
	case *env.Status.Error != e.Response:
		t.Logf("unexpected error: %s", *env.Status.Error)
		t.Fail()
	case *env.Status.Message != e.Error():
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

func TestErrorEnvelope(t *testing.T) {
	err := HandlerError{
		HTTPStatus: http.StatusNotAcceptable,
	}

	target := Envelope{
		Data: nil,
		Status: EnvelopeStatus{
			HTTPStatus: http.StatusNotAcceptable,
			Message:    nil,
			Error:      toPtr(http.StatusText(http.StatusNotAcceptable)),
		},
	}
	envCompare(t, err.ToEnvelope(), &target)

	err.Response = "test response"
	target.Status.Error = toPtr("test response")
	envCompare(t, err.ToEnvelope(), &target)

	Config.SetForwardLogMessage(true)
	target.Status.Message = toPtr(err.Error())
	envCompare(t, err.ToEnvelope(), &target)

	err.err = fmt.Errorf("test error")
	target.Status.Message = toPtr(err.Error())
	envCompare(t, err.ToEnvelope(), &target)

	err.HTTPStatus = 333
	err.Response = ""
	target.Status.Error = toPtr("unkown error")
	target.Status.HTTPStatus = 333
	envCompare(t, err.ToEnvelope(), &target)

	Config.Reset()
}

func envCompare(t *testing.T, env *Envelope, target *Envelope) {
	switch {
	case env.Status.HTTPStatus != target.Status.HTTPStatus:
		t.Logf("unexpected status: %d", env.Status.HTTPStatus)
		t.Fail()
	case env.Data != target.Data:
		t.Logf("unexpected data: %s", env.Data)
		t.Fail()
	case !strPtrCompare(env.Status.Message, target.Status.Message):
		t.Log("message not match")
		if target.Status.Message != nil {
			t.Logf("target: %s", *target.Status.Message)
		} else {
			t.Log("target: nil")
		}
		if env.Status.Message != nil {
			t.Logf("got: %s", *env.Status.Message)
		} else {
			t.Log("got: nil")
		}
		t.Fail()
	case !strPtrCompare(env.Status.Error, target.Status.Error):
		t.Log("error not match")
		if target.Status.Error != nil {
			t.Logf("target: %s", *target.Status.Error)
		} else {
			t.Log("target: nil")
		}
		if env.Status.Error != nil {
			t.Logf("got: %s", *env.Status.Error)
		} else {
			t.Log("got: nil")
		}
	}
}

func strPtrCompare(s1 *string, s2 *string) bool {
	switch {
	case s1 == nil && s2 == nil:
		return true
	case s1 == nil && s2 != nil || s1 != nil && s2 == nil:
		return false
	case *s1 != *s2:
		return false
	case *s1 == *s2:
		return true
	default:
		return false
	}
}

func errorCompare(e1 HandlerError, e2 HandlerError) bool {
	return e1.Response == e2.Response &&
		e1.HTTPStatus == e2.HTTPStatus &&
		e1.LogMessage == e2.LogMessage
}
