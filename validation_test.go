package testkit_test

import (
	"testing"

	. "github.com/dogmatiq/testkit"
)

func TestValidationScopeCreation(t *testing.T) {
	if CommandValidationScope() == nil {
		t.Errorf("CommandValidationScope() returned nil")
	}

	if EventValidationScope() == nil {
		t.Errorf("EventValidationScope() returned nil")
	}

	if TimeoutValidationScope() == nil {
		t.Errorf("TimeoutValidationScope() returned nil")
	}
}
