package httpx_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gopomodoro/internal/httpx"
)

func TestFailPreservesWrappedAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	httpx.Fail(ctx, fmt.Errorf("load pomodoro session: %w", httpx.ErrNotFound))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
	var response httpx.Envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.OK || response.Error == nil || response.Error.Code != httpx.ErrNotFound.Code {
		t.Fatalf("response = %+v, want wrapped %s error", response, httpx.ErrNotFound.Code)
	}
}
