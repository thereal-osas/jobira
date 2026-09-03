package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler()

	if handler == nil {
		t.Fatal("expected handler")
	}
}

func TestHandler_HealthCheck(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.HealthCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var body HealthResponse

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if body.Status != "ok" {
		t.Fatalf(
			"expected status ok, got %q",
			body.Status,
		)
	}
}
