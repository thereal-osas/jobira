package adminmoderation

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func adminModerationRequestWithIdentity(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
			Role:   role,
		},
	)

	return req.WithContext(ctx)
}

func adminModerationRequestWithParam(
	req *http.Request,
	name string,
	value string,
) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(name, value)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	)

	return req.WithContext(ctx)
}

func TestHandler_ListReports_Success(t *testing.T) {
	repo := &mockRepository{
		listReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return []CleanerReportAdminView{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "open",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListReports(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_ListReports_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list reports failed",
	)

	repo := &mockRepository{
		listReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListReports(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListOpenReports_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		listOpenReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return []CleanerReportAdminView{
				{
					ID:     1,
					Status: "open",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports/open",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListOpenReports(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListOpenReports_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list open reports failed",
	)

	repo := &mockRepository{
		listOpenReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports/open",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListOpenReports(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListBlockCleaners_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		listBlockCleanersFn: func(
			context.Context,
		) ([]BlockedCleanerAdminView, error) {
			return []BlockedCleanerAdminView{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/blocked-cleaners",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListBlockCleaners(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListBlockCleaners_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list blocked cleaners failed",
	)

	repo := &mockRepository{
		listBlockCleanersFn: func(
			context.Context,
		) ([]BlockedCleanerAdminView, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/blocked-cleaners",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListBlockCleaners(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		updateReportStatusFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		bytes.NewBufferString(`{
			"status":"resolved",
			"admin_notes":"Reviewed"
		}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"12",
	)

	req = adminModerationRequestWithIdentity(
		req,
		9,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_UpdatedReportStatus_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		bytes.NewBufferString(`{
			"status":"resolved"
		}`),
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_InvalidReportID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/invalid",
		bytes.NewBufferString(`{
			"status":"resolved"
		}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"invalid",
	)

	req = adminModerationRequestWithIdentity(
		req,
		9,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_InvalidBody(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		bytes.NewBufferString(`{invalid-json}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"12",
	)

	req = adminModerationRequestWithIdentity(
		req,
		9,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_InvalidInput(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		bytes.NewBufferString(`{
			"status":"resolved"
		}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"12",
	)

	req = adminModerationRequestWithIdentity(
		req,
		0,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_InvalidStatus(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		bytes.NewBufferString(`{
			"status":"pending"
		}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"12",
	)

	req = adminModerationRequestWithIdentity(
		req,
		9,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_NotFound(
	t *testing.T,
) {
	repo := &mockRepository{
		updateReportStatusFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
		) error {
			return ErrReportNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/999",
		bytes.NewBufferString(`{
			"status":"resolved"
		}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"999",
	)

	req = adminModerationRequestWithIdentity(
		req,
		9,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestHandler_UpdatedReportStatus_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"update failed",
	)

	repo := &mockRepository{
		updateReportStatusFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
		) error {
			return expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		bytes.NewBufferString(`{
			"status":"resolved"
		}`),
	)

	req = adminModerationRequestWithParam(
		req,
		"reportID",
		"12",
	)

	req = adminModerationRequestWithIdentity(
		req,
		9,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.UpdatedReportStatus(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
