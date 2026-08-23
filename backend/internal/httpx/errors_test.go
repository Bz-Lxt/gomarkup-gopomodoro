package httpx

import "testing"

func TestAppErrorAs(t *testing.T) {
	err := ErrInvalidTransition.WithDetails(map[string]any{"from": "idle"})
	ae, ok := IsAppError(err)
	if !ok || ae.Code != "E_INVALID_TRANSITION" {
		t.Fatal(err)
	}
	if ae.HTTPStatus != 409 {
		t.Fatal(ae.HTTPStatus)
	}
}
