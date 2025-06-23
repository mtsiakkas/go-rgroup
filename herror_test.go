package rgroup_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
)

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
