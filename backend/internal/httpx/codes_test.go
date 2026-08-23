package httpx

import "testing"

func TestLookupCode(t *testing.T) {
	if LookupCode("E_INVALID_TRANSITION") != 409 {
		t.Fatal("transition")
	}
	if LookupCode("E_NOPE") != 500 {
		t.Fatal("unknown")
	}
	if len(CodeCatalog) < 9 {
		t.Fatal("catalog")
	}
}
