package cleanerreports

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

func cleanerReportsRequestWithIdentity(
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

func cleanerReportsRequestWithParam(
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

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			report *CleanerReport,
		) error {
			report.ID = 20
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"reason":"No show",
		"details":"Cleaner did not attend."
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/8",
		bytes.NewReader(body),
	)

	req = cleanerReportsRequestWithParam(
		req,
		"cleanerID",
		"8",
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
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
		"/cleaner-reports/8",
		bytes.NewBufferString(`{
			"reason":"No show"
		}`),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_Create_InvalidCleanerID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/invalid",
		bytes.NewBufferString(`{
			"reason":"No show"
		}`),
	)

	req = cleanerReportsRequestWithParam(
		req,
		"cleanerID",
		"invalid",
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Create_InvalidRequestBody(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/8",
		bytes.NewBufferString(`{invalid-json}`),
	)

	req = cleanerReportsRequestWithParam(
		req,
		"cleanerID",
		"8",
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Create_InvalidInput(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/8",
		bytes.NewBufferString(`{
			"reason":"   "
		}`),
	)

	req = cleanerReportsRequestWithParam(
		req,
		"cleanerID",
		"8",
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Create_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"create report failed",
	)

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*CleanerReport,
		) error {
			return expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/8",
		bytes.NewBufferString(`{
			"reason":"No show"
		}`),
	)

	req = cleanerReportsRequestWithParam(
		req,
		"cleanerID",
		"8",
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]CleanerReport, error) {
			return []CleanerReport{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Reason:    "No show",
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
		"/cleaner-reports/me",
		nil,
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_ListMine_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-reports/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_InvalidInput(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-reports/me",
		nil,
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		0,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list reports failed",
	)

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]CleanerReport, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-reports/me",
		nil,
	)

	req = cleanerReportsRequestWithIdentity(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ListMine(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
