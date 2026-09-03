package reports

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func reportsRequestWithUser(
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

func reportsRequestWithParam(
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

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			report *Report,
		) error {
			report.ID = 10
			return nil
		},
	}

	handler := NewHandler(NewService(repo))

	body := `{
		"reported_user_id": 20,
		"report_type": "user",
		"reason": "abusive behaviour",
		"details": "details"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		bytes.NewBufferString(body),
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"id":10`,
	) {
		t.Fatalf(
			"expected report id in response: %s",
			rec.Body.String(),
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		bytes.NewBufferString(`{}`),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		bytes.NewBufferString(`{invalid`),
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		bytes.NewBufferString(`{
			"report_type": "",
			"reason": ""
		}`),
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidReportType(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		bytes.NewBufferString(`{
			"report_type": "invalid",
			"reason": "reason"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Create_ServiceError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*Report,
		) error {
			return expectedErr
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		bytes.NewBufferString(`{
			"report_type": "user",
			"reason": "reason"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_GetByID_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				ReporterID: 5,
				Status:     "open",
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/10",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetByID_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/10",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/abc",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"abc",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return nil, ErrReportNotFound
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/999",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestHandler_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				ReporterID: 50,
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/10",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/10",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByReporterIDFn: func(
			context.Context,
			uint,
		) ([]Report, error) {
			return []Report{
				{ID: 1, ReporterID: 5},
				{ID: 2, ReporterID: 5},
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/me",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_ServiceError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		listByReporterIDFn: func(
			context.Context,
			uint,
		) ([]Report, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/me",
		nil,
	)

	req = reportsRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListAll_Success(t *testing.T) {
	repo := &mockRepository{
		listAllFn: func(
			context.Context,
		) ([]Report, error) {
			return []Report{
				{ID: 1},
				{ID: 2},
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListAll_Error(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		listAllFn: func(
			context.Context,
		) ([]Report, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListAll(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListOpen_Success(t *testing.T) {
	repo := &mockRepository{
		listOpenFn: func(
			context.Context,
		) ([]Report, error) {
			return []Report{
				{
					ID:     1,
					Status: "open",
				},
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/open",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListOpen(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListOpen_Error(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		listOpenFn: func(
			context.Context,
		) ([]Report, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/open",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListOpen(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_Review_Success(t *testing.T) {
	repo := &mockRepository{
		reviewFn: func(
			context.Context,
			uint,
			string,
			string,
			uint,
		) error {
			return nil
		},
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				Status:     "resolved",
				AdminNotes: "reviewed",
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/10",
		bytes.NewBufferString(`{
			"status": "resolved",
			"admin_notes": "reviewed"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		99,
		"admin",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Review_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/10",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_Review_InvalidID(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/abc",
		bytes.NewBufferString(`{
			"status": "resolved"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		99,
		"admin",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"abc",
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Review_InvalidBody(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/10",
		bytes.NewBufferString(`{invalid`),
	)

	req = reportsRequestWithUser(
		req,
		99,
		"admin",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Review_InvalidStatus(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/10",
		bytes.NewBufferString(`{
			"status": "invalid"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		99,
		"admin",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Review_NotFound(t *testing.T) {
	repo := &mockRepository{
		reviewFn: func(
			context.Context,
			uint,
			string,
			string,
			uint,
		) error {
			return ErrReportNotFound
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/999",
		bytes.NewBufferString(`{
			"status": "resolved"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		99,
		"admin",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestHandler_Review_ServiceError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		reviewFn: func(
			context.Context,
			uint,
			string,
			string,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := NewHandler(NewService(repo))

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/10",
		bytes.NewBufferString(`{
			"status": "resolved"
		}`),
	)

	req = reportsRequestWithUser(
		req,
		99,
		"admin",
	)

	req = reportsRequestWithParam(
		req,
		"reportID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.Review(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
