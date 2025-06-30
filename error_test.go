package rgroup_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
)

func TestError(t *testing.T) {
	e := rgroup.Error(http.StatusInternalServerError)
	if !errorCompare(rgroup.HandlerError{
		LogMessage: "",
		HTTPStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("rgroup.Error() failed")
		t.Fail()
	}

	_ = e.WithMessage("test error: %s", "test message")
	if !errorCompare(rgroup.HandlerError{
		LogMessage: "test error: test message",
		HTTPStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithMessage failed")
		t.Fail()
	}

	_ = e.WithResponse("test error: %s", "test response")
	if !errorCompare(rgroup.HandlerError{
		LogMessage: "test error: test message",
		Response:   "test error: test response",
		HTTPStatus: http.StatusInternalServerError,
	}, *e) {
		t.Log("WithResponse failed")
		t.Fail()
	}

	etmp := errors.New("error")
	e2 := rgroup.Error(http.StatusInternalServerError).WithMessage("test").Wrap(etmp)
	if e2.Error() != "test: error" {
		t.Logf("unexpected error message: \"%s\"", e2)
		t.Fail()
	}

	if e2.Unwrap() != etmp {
		t.Logf("unexpected error unwrap: \"%s\"", e2.Unwrap())
		t.Fail()
	}
}

func errorCompare(e1 rgroup.HandlerError, e2 rgroup.HandlerError) bool {
	return e1.Response == e2.Response &&
		e1.HTTPStatus == e2.HTTPStatus &&
		e1.LogMessage == e2.LogMessage
}
