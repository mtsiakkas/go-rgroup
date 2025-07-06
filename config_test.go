package rgroup_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
)

func TestGlobalOverwriteHandler(t *testing.T) {
	t.Run("set - unknown option", func(t *testing.T) {
		if err := rgroup.Config.SetOverwriteMethodBehaviour(rgroup.OverwriteMethodBehaviour(4)); err == nil {
			t.Log("expected error")
			t.Fail()
		}
	})

	t.Run("set - success", func(t *testing.T) {
		if err := rgroup.Config.SetOverwriteMethodBehaviour(rgroup.OverwriteMethodError); err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}
	})

	t.Run("validate", func(t *testing.T) {
		if !rgroup.OverwriteMethodError.Validate() {
			t.Logf("%s not validated", rgroup.OverwriteMethodError)
			t.Fail()
		}
	})

	t.Run("stringer", func(t *testing.T) {
		if rgroup.OverwriteMethodError.String() != "error" {
			t.Logf("unexpected .String(): %s", rgroup.OverwriteMethodError.String())
			t.Fail()
		}
	})

	t.Run("get", func(t *testing.T) {
		if rgroup.Config.GetOverwriteMethod() != rgroup.OverwriteMethodError {
			t.Logf("got %s", rgroup.Config.GetOverwriteMethod())
			t.Fail()
		}
	})

	_ = rgroup.Config.SetOverwriteMethodBehaviour(rgroup.OverwriteMethodPanic)

}

func TestGlobalOptionsHandler(t *testing.T) {

	t.Run("options - unknown option", func(t *testing.T) {
		if err := rgroup.Config.SetOverwriteOptionsHandlerBehaviour(rgroup.OverwriteOptionsHandlerBehaviour(4)); err == nil {
			t.Log("expected error", err)
			t.Fail()
		}
	})

	t.Run("set - success", func(t *testing.T) {
		if err := rgroup.Config.SetOverwriteOptionsHandlerBehaviour(rgroup.OverwriteOptionsHandlerIgnore); err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}
	})

	t.Run("validate", func(t *testing.T) {
		if !rgroup.OverwriteOptionsHandlerIgnore.Validate() {
			t.Logf("%s not validated", rgroup.OverwriteOptionsHandlerIgnore)
			t.Fail()
		}
	})

	t.Run("stringer", func(t *testing.T) {
		if rgroup.OverwriteOptionsHandlerIgnore.String() != "ignore" {
			t.Logf("unexpected .String(): %s", rgroup.OverwriteOptionsHandlerIgnore.String())
			t.Fail()
		}
	})

	t.Run("get", func(t *testing.T) {
		if rgroup.Config.GetOverwriteOptionsHandlerBehaviour() != rgroup.OverwriteOptionsHandlerIgnore {
			t.Fail()
		}
	})

	_ = rgroup.Config.SetOverwriteOptionsHandlerBehaviour(rgroup.OverwriteOptionsHandlerPanic)
}

func TestGlobalPostprocessor(t *testing.T) {
	rgroup.Config.SetGlobalLogger(func(req *rgroup.RequestData) {
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
