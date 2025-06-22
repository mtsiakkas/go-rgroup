//go:build !test

package rgroup_test

import (
	"testing"
)

func TestTags(t *testing.T) {
	t.Fatalf("test should be run with the \"test\" build tag")
}
