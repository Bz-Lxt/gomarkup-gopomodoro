package validate_test

import (
	"testing"

	"gopomodoro/internal/validate"
)

func TestRequiredStringAcceptsValidValue(t *testing.T) {
	got, err := validate.RequiredString("name", "  Sprint Alpha  ", 1, 80)
	if err != nil {
		t.Fatalf("valid project name was rejected: %v", err)
	}
	if got != "Sprint Alpha" {
		t.Fatalf("normalized project name = %q, want %q", got, "Sprint Alpha")
	}
}
