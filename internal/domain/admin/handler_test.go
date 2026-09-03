package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	service := NewService()

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestHandler_Dashboard_Success(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/dashboard",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Dashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var responseBody DashboardMessage

	err := json.NewDecoder(rec.Body).Decode(
		&responseBody,
	)
	if err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if responseBody.Message != "Admin dashboard ready" {
		t.Fatalf(
			"expected dashboard message %q, got %q",
			"Admin dashboard ready",
			responseBody.Message,
		)
	}
}

func TestHandler_Dashboard_ContentType(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/dashboard",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Dashboard(rec, req)

	contentType := rec.Header().Get(
		"Content-Type",
	)

	if contentType != "application/json" {
		t.Fatalf(
			"expected application/json content type, got %q",
			contentType,
		)
	}
}
