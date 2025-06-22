//go:build test

package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtsiakkas/go-rgroup/pkg/request"
)

func TestRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?t=1", nil)

	l := request.FromRequest(req)

	if l.Path != "/test" || l.Method != http.MethodGet || l.Status != http.StatusOK {
		t.Logf("unexpected data: %v", l)
		t.Fail()
	}

	t1 := l.Time()
	if t1 <= 0 {
		t.Logf("unexpected request time: %d", t1)
		t.Fail()
	}

	t2 := l.Time()

	if t1 != t2 {
		t.Logf("unexpected request.Time(): %d, %d", t1, t2)
		t.Fail()
	}

}
