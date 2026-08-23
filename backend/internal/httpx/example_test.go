package httpx_test

import (
	"encoding/json"
	"testing"

	"gopomodoro/internal/httpx"
)

func TestEnvelopeShape(t *testing.T) {
	ok := httpx.Envelope{OK: true, Data: map[string]any{"live": 0}}
	b, err := json.Marshal(ok)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" {
		t.Fatal("empty")
	}
	fail := httpx.Envelope{OK: false, Error: &httpx.ErrorPayload{Code: "E_INVALID_TRANSITION", Message: "非法的番茄钟状态迁移"}}
	b, _ = json.Marshal(fail)
	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil || parsed["ok"] != false {
		t.Fatal(parsed)
	}
}

func TestAllCatalogJSON(t *testing.T) {
	for _, c := range httpx.CodeCatalog {
		if httpx.LookupCode(c.Code) != c.Status {
			t.Fatalf("%s", c.Code)
		}
	}
}
