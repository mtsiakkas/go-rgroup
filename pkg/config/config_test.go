//go:build test

package config_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtsiakkas/go-rgroup"
	testing_helpers "github.com/mtsiakkas/go-rgroup/internal/testing"
	"github.com/mtsiakkas/go-rgroup/pkg/config"
	"github.com/mtsiakkas/go-rgroup/pkg/group"
	"github.com/mtsiakkas/go-rgroup/pkg/request"
	"github.com/mtsiakkas/go-rgroup/pkg/response"
)

func TestDuplicate(t *testing.T) {
	t.Run("set - unknown option", func(t *testing.T) {
		if err := config.OnDuplicateMethod(config.DuplicateMethodBehaviour(4)); err == nil {
			t.Log("expected error")
			t.Fail()
		}
	})

	t.Run("set - success", func(t *testing.T) {
		if err := config.OnDuplicateMethod(config.DuplicateMethodError); err != nil {
			t.Log("expected error")
			t.Fail()
		}
	})

	t.Run("validate", func(t *testing.T) {
		if !config.DuplicateMethodError.Validate() {
			t.Logf("%s not validated", config.DuplicateMethodError)
			t.Fail()
		}
	})

	t.Run("stringer", func(t *testing.T) {
		if config.DuplicateMethodError.String() != "error" {
			t.Logf("unexpected .String(): %s", config.DuplicateMethodError.String())
			t.Fail()
		}
	})

	t.Run("get", func(t *testing.T) {
		if config.GetDuplicateMethod() != config.DuplicateMethodError {
			t.Logf("got %s", config.GetDuplicateMethod())
			t.Fail()
		}
	})

}

func TestOptions(t *testing.T) {

	t.Run("options - unknown option", func(t *testing.T) {
		if err := config.OnOptionsHandler(config.OptionsHandlerBehaviour(4)); err == nil {
			t.Log("expected error", err)
			t.Fail()
		}
	})

	t.Run("set - success", func(t *testing.T) {
		if err := config.OnOptionsHandler(config.OptionsHandlerIgnore); err != nil {
			t.Log("expected error")
			t.Fail()
		}
	})

	t.Run("validate", func(t *testing.T) {
		if !config.OptionsHandlerIgnore.Validate() {
			t.Logf("%s not validated", config.OptionsHandlerIgnore)
			t.Fail()
		}
	})

	t.Run("stringer", func(t *testing.T) {
		if config.OptionsHandlerIgnore.String() != "ignore" {
			t.Logf("unexpected .String(): %s", config.OptionsHandlerIgnore.String())
			t.Fail()
		}
	})

	t.Run("get", func(t *testing.T) {
		if config.GetOnOptionsHandler() != config.OptionsHandlerIgnore {
			t.Fail()
		}
	})
}

func TestPostprocessor(t *testing.T) {
	config.SetGlobalPostprocessor(func(ctx context.Context, req *request.RequestData) {
		fmt.Println("global postprocessor")
	})

	h := rgroup.NewWithHandlers(map[string]group.Handler{"GET": func(w http.ResponseWriter, req *http.Request) (*response.HandlerResponse, error) {
		return nil, nil
	}}).Make()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := testing_helpers.CaptureOutput(func() { h(rr, req) })
	if res != "global postprocessor\n" {
		t.Logf("unexpected log: %s", res)
		t.Fail()
	}

	p := config.GetGlobalPostprocessor()
	if p == nil {
		t.Log("expected not nil global postprocessor")
		t.Fail()
	}
}
