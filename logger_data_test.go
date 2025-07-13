package rgroup

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerData(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?t=1", nil)

	l := fromRequest(*req)

	if l.Path() != "/test" || l.Status() != http.StatusOK || l.Message() != "" {
		t.Logf("unexpected data: %v", l)
		t.Fail()
	}

	t1 := l.Duration()
	if t1 <= 0 {
		t.Logf("unexpected request time: %d", t1)
		t.Fail()
	}

	t2 := l.Duration()

	if t1 != t2 {
		t.Logf("unexpected request.Time(): %d, %d", t1, t2)
		t.Fail()
	}

	l.Response = Response("test message").WithMessage("test message").WithHTTPStatus(http.StatusAccepted)
	if l.Status() != http.StatusAccepted {
		t.Logf("unexpected status: %d", l.Status())
		t.Fail()
	}

	if l.Message() != "test message" {
		t.Logf("unexpected message: %s", l.Message())
		t.Fail()
	}

	l.Error = Error(http.StatusNotAcceptable).WithMessage("test error")
	if l.Status() != http.StatusNotAcceptable {
		t.Logf("unexpected status: %d", l.Status())
		t.Fail()
	}

	if l.Message() != "test error" {
		t.Logf("unexpected message: %s", l.Message())
		t.Fail()
	}

}
