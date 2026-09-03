package analytics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func analyticsAuthenticatedRequest(
	req *http.Request,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 1,
			Role:   "admin",
		},
	)

	return req.WithContext(ctx)
}

func TestNewHandler(t *testing.T) {
	service := NewService(&mockRepository{})

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestHandler_Dashboard_Success(t *testing.T) {
	repo := dashboardBaseMock()

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/analytics/dashboard",
		nil,
	)

	req = analyticsAuthenticatedRequest(req)

	rec := httptest.NewRecorder()

	handler.Dashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Dashboard_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/analytics/dashboard",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Dashboard(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Dashboard_ServiceError(
	t *testing.T,
) {
	expectedErr := errors.New("analytics failed")

	repo := &mockRepository{
		totalUsersFn: func(
			context.Context,
		) (int, error) {
			return 0, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/analytics/dashboard",
		nil,
	)

	req = analyticsAuthenticatedRequest(req)

	rec := httptest.NewRecorder()

	handler.Dashboard(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}
