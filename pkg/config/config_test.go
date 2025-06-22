//go:build test
package config_test

import (
	"testing"

	"github.com/mtsiakkas/go-rgroup/pkg/config"
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
