package rgroup_test

import (
	"net/http"
	"testing"

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

	if r.HTTPStatus != http.StatusOK {
		t.Logf("unexpected r.HttpStatus value: expected \"%d\" got \"%d\"", http.StatusOK, r.HTTPStatus)
		t.Fail()
	}

	r.WithHTTPStatus(http.StatusAccepted)

	if r.HTTPStatus != http.StatusAccepted {
		t.Logf("unexpected r.HttpStatus value: expected \"%d\" got \"%d\"", http.StatusAccepted, r.HTTPStatus)
		t.Fail()
	}
}
