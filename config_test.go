package rgroup_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
)

func TestGlobalDuplicateHandler(t *testing.T) {
	t.Run("set - unknown option", func(t *testing.T) {
		if err := rgroup.OnDuplicateMethod(rgroup.DuplicateMethodBehaviour(4)); err == nil {
			t.Log("expected error")
			t.Fail()
		}
	})

	t.Run("set - success", func(t *testing.T) {
		if err := rgroup.OnDuplicateMethod(rgroup.DuplicateMethodError); err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}
	})

	t.Run("validate", func(t *testing.T) {
		if !rgroup.DuplicateMethodError.Validate() {
			t.Logf("%s not validated", rgroup.DuplicateMethodError)
			t.Fail()
		}
	})

	t.Run("stringer", func(t *testing.T) {
		if rgroup.DuplicateMethodError.String() != "error" {
			t.Logf("unexpected .String(): %s", rgroup.DuplicateMethodError.String())
			t.Fail()
		}
	})

	t.Run("get", func(t *testing.T) {
		if rgroup.GetDuplicateMethod() != rgroup.DuplicateMethodError {
			t.Logf("got %s", rgroup.GetDuplicateMethod())
			t.Fail()
		}
	})

	rgroup.OnDuplicateMethod(rgroup.DuplicateMethodPanic)

}

func TestGlobalOptionsHandler(t *testing.T) {

	t.Run("options - unknown option", func(t *testing.T) {
		if err := rgroup.OnOptionsHandler(rgroup.OptionsHandlerBehaviour(4)); err == nil {
			t.Log("expected error", err)
			t.Fail()
		}
	})

	t.Run("set - success", func(t *testing.T) {
		if err := rgroup.OnOptionsHandler(rgroup.OptionsHandlerIgnore); err != nil {
			t.Logf("unexpected error: %s", err)
			t.Fail()
		}
	})

	t.Run("validate", func(t *testing.T) {
		if !rgroup.OptionsHandlerIgnore.Validate() {
			t.Logf("%s not validated", rgroup.OptionsHandlerIgnore)
			t.Fail()
		}
	})

	t.Run("stringer", func(t *testing.T) {
		if rgroup.OptionsHandlerIgnore.String() != "ignore" {
			t.Logf("unexpected .String(): %s", rgroup.OptionsHandlerIgnore.String())
			t.Fail()
		}
	})

	t.Run("get", func(t *testing.T) {
		if rgroup.GetOnOptionsHandler() != rgroup.OptionsHandlerIgnore {
			t.Fail()
		}
	})

	rgroup.OnOptionsHandler(rgroup.OptionsHandlerPanic)
}

func TestGlobalPostprocessor(t *testing.T) {
	rgroup.SetGlobalPostprocessor(func(ctx context.Context, req *rgroup.RequestData) {
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

	p := rgroup.GetGlobalPostprocessor()
	if p == nil {
		t.Log("expected not nil global postprocessor")
		t.Fail()
	}

	rgroup.SetGlobalPostprocessor(nil)
}
