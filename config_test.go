package rgroup_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
)

func TestGlobalPostprocessor(t *testing.T) {
	rgroup.Config.SetGlobalLogger(func(req *rgroup.LoggerData) {
		fmt.Println("global postprocessor")
	})

	h := rgroup.NewWithHandlers(rgroup.HandlerMap{"GET": func(w http.ResponseWriter, req *http.Request) (*rgroup.HandlerResponse, error) {
		return nil, nil
	}}).Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := captureOutput(func() { h(rr, req) })
	if res != "global postprocessor\n" {
		t.Logf("unexpected log: %s", res)
		t.Fail()
	}

	p := rgroup.Config.GetGlobalLogger()
	if p == nil {
		t.Log("expected not nil global postprocessor")
		t.Fail()
	}

	rgroup.Config.SetGlobalLogger(nil)
}
