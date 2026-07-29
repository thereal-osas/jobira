package applications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestNewHandler(t *testing.T) {
	service := &Service{}

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service to be assigned")
	}
}

func TestParseIDParam_Success(t *testing.T) {
	request, err := http.NewRequest(
		http.MethodGet,
		"/applications/42",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "42")

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	result, err := parseIDParam(request, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 42 {
		t.Fatalf("expected ID 42, got %d", result)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	request, err := http.NewRequest(
		http.MethodGet,
		"/applications/not-a-number",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "not-a-number")

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	result, err := parseIDParam(request, "id")
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if result != 0 {
		t.Fatalf("expected ID 0, got %d", result)
	}
}

func TestParseIDParam_MissingID(t *testing.T) {
	request, err := http.NewRequest(
		http.MethodGet,
		"/applications",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	routeContext := chi.NewRouteContext()

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	result, err := parseIDParam(request, "id")
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if result != 0 {
		t.Fatalf("expected ID 0, got %d", result)
	}
}

func TestParseIDParam_ZeroID(t *testing.T) {
	request, err := http.NewRequest(
		http.MethodGet,
		"/applications/0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "0")

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	result, err := parseIDParam(request, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 0 {
		t.Fatalf("expected ID 0, got %d", result)
	}
}

func newAuthenticatedRequest(
	method string,
	target string,
	body any,
	userID uint,
	role string,
) *http.Request {
	var requestBody *bytes.Buffer

	if body != nil {
		payload, _ := json.Marshal(body)
		requestBody = bytes.NewBuffer(payload)
	} else {
		requestBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, target, requestBody)
	req.Header.Set("Content-Type", "application/json")

	user := identity.UserIdentity{
		UserID: userID,
		Role:   role,
	}

	ctx := identity.WithUser(req.Context(), user)

	return req.WithContext(ctx)
}

func addRouteParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	)

	return req.WithContext(ctx)
}

func TestHandler_Apply_Success(t *testing.T) {
	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		CreateFunc: func(ctx context.Context, application *Application) error {
			application.ID = 1
			return nil
		},
		GetJobTitleFunc: func(ctx context.Context, jobID uint) (string, error) {
			return "Office cleaning", nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/10",
		CreateApplicationRequest{
			CoverMessage: "I am available tomorrow.",
			ProposedRate: 75,
		},
		20,
		"cleaner",
	)
	request = addRouteParam(request, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var result Application

	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.ID != 1 {
		t.Fatalf("expected application ID 1, got %d", result.ID)
	}

	if result.JobID != 10 {
		t.Fatalf("expected job ID 10, got %d", result.JobID)
	}

	if result.CleanerID != 20 {
		t.Fatalf("expected cleaner ID 20, got %d", result.CleanerID)
	}

	if result.Status != "pending" {
		t.Fatalf("expected pending status, got %q", result.Status)
	}
}

func TestHandler_Apply_Unauthorized(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/applications/jobs/10",
		bytes.NewBufferString(`{
			"cover_message": "Available",
			"proposed_rate": 50
		}`),
	)

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Apply_InvalidJobID(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/not-a-number",
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
		20,
		"cleaner",
	)
	request = addRouteParam(request, "jobID", "not-a-number")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Apply_InvalidRequestBody(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/10",
		nil,
		20,
		"cleaner",
	)
	request.Body = http.NoBody
	request = addRouteParam(request, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Apply_InvalidInput(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/10",
		CreateApplicationRequest{
			CoverMessage: "   ",
			ProposedRate: 50,
		},
		20,
		"cleaner",
	)
	request = addRouteParam(request, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Apply_AlreadyApplied(t *testing.T) {
	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		CreateFunc: func(ctx context.Context, application *Application) error {
			return ErrAlreadyApplied
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/10",
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
		20,
		"cleaner",
	)
	request = addRouteParam(request, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Apply_CannotApplyToOwnJob(t *testing.T) {
	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 20, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/10",
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
		20,
		"cleaner",
	)
	request = addRouteParam(request, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Apply_InternalServerError(t *testing.T) {
	expectedError := errors.New("database unavailable")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 0, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	request := newAuthenticatedRequest(
		http.MethodPost,
		"/applications/jobs/10",
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
		20,
		"cleaner",
	)
	request = addRouteParam(request, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.Apply(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_Success(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "pending",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)
	handler := NewHandler(service)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/1",
		nil,
		20,
		"cleaner",
	)
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, recorder.Code)
	}

	var application Application

	if err := json.NewDecoder(recorder.Body).Decode(&application); err != nil {
		t.Fatal(err)
	}

	if application.ID != 1 {
		t.Fatalf("expected application ID 1 got %d", application.ID)
	}
}

func TestHandler_GetByID_Unauthorized(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil, nil, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/applications/1",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil, nil, nil, nil))

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/abc",
		nil,
		20,
		"cleaner",
	)

	req = addRouteParam(req, "id", "abc")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return nil, ErrApplicationNotFound
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/1",
		nil,
		20,
		"cleaner",
	)

	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected %d got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestHandler_GetByID_Forbidden(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/1",
		nil,
		999,
		"cleaner",
	)

	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestHandler_GetByID_InternalServerError(t *testing.T) {
	expected := errors.New("database offline")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return nil, expected
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/1",
		nil,
		20,
		"cleaner",
	)

	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestHandler_ListForJob_Success(t *testing.T) {
	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			return []Application{
				{
					ID:        1,
					JobID:     jobID,
					CleanerID: 20,
					Status:    "pending",
				},
				{
					ID:        2,
					JobID:     jobID,
					CleanerID: 21,
					Status:    "shortlisted",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/jobs/10",
		nil,
		50,
		"client",
	)
	req = addRouteParam(req, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var applications []Application

	if err := json.NewDecoder(recorder.Body).Decode(&applications); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(applications) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(applications))
	}
}

func TestHandler_ListForJob_AdminSuccess(t *testing.T) {
	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			return []Application{
				{
					ID:    1,
					JobID: jobID,
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/jobs/10",
		nil,
		999,
		"admin",
	)
	req = addRouteParam(req, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListForJob_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/applications/jobs/10",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListForJob_InvalidJobID(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/jobs/abc",
		nil,
		50,
		"client",
	)
	req = addRouteParam(req, "jobID", "abc")

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListForJob_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/jobs/0",
		nil,
		50,
		"client",
	)
	req = addRouteParam(req, "jobID", "0")

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListForJob_Forbidden(t *testing.T) {
	listCalled := false

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			listCalled = true
			return nil, nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/jobs/10",
		nil,
		999,
		"client",
	)
	req = addRouteParam(req, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if listCalled {
		t.Fatal("ListByJobID should not be called for unauthorized user")
	}
}

func TestHandler_ListForJob_InternalServerError(t *testing.T) {
	expectedError := errors.New("database unavailable")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 0, expectedError
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/jobs/10",
		nil,
		50,
		"client",
	)
	req = addRouteParam(req, "jobID", "10")

	recorder := httptest.NewRecorder()

	handler.ListForJob(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &MockRepository{
		ListByCleanerIDFunc: func(ctx context.Context, cleanerID uint) ([]Application, error) {
			if cleanerID != 20 {
				t.Fatalf("expected cleaner ID 20, got %d", cleanerID)
			}

			return []Application{
				{
					ID:        1,
					JobID:     10,
					CleanerID: cleanerID,
					Status:    "pending",
				},
				{
					ID:        2,
					JobID:     11,
					CleanerID: cleanerID,
					Status:    "accepted",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/me",
		nil,
		20,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var applications []Application

	if err := json.NewDecoder(recorder.Body).Decode(&applications); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(applications) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(applications))
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/applications/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMine_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/me",
		nil,
		0,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repo := &MockRepository{
		ListByCleanerIDFunc: func(ctx context.Context, cleanerID uint) ([]Application, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodGet,
		"/applications/me",
		nil,
		20,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}

func TestHandler_UpdateStatus_Success(t *testing.T) {
	getCalls := 0

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			getCalls++

			if getCalls == 1 {
				return &Application{
					ID:        id,
					JobID:     10,
					CleanerID: 20,
					Status:    "pending",
				}, nil
			}

			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "accepted",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			if id != 1 {
				t.Fatalf("expected application ID 1, got %d", id)
			}

			if status != "accepted" {
				t.Fatalf("expected accepted status, got %q", status)
			}

			return nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/1/status",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
		50,
		"client",
	)
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var application Application

	if err := json.NewDecoder(recorder.Body).Decode(&application); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if application.Status != "accepted" {
		t.Fatalf("expected accepted status, got %q", application.Status)
	}
}

func TestHandler_UpdateStatus_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/applications/1/status",
		bytes.NewBufferString(`{"status":"accepted"}`),
	)

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateStatus_InvalidID(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/abc/status",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
		50,
		"client",
	)
	req = addRouteParam(req, "id", "abc")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateStatus_InvalidRequestBody(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/1/status",
		nil,
		50,
		"client",
	)
	req.Body = http.NoBody
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateStatus_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/0/status",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
		50,
		"client",
	)
	req = addRouteParam(req, "id", "0")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/1/status",
		UpdateApplicationStatusRequest{
			Status: "unknown",
		},
		50,
		"client",
	)
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateStatus_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return nil, ErrApplicationNotFound
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/1/status",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
		50,
		"client",
	)
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateStatus_Forbidden(t *testing.T) {
	updateCalled := false

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "pending",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			updateCalled = true
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/1/status",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
		999,
		"client",
	)
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if updateCalled {
		t.Fatal("unauthorized user must not update application status")
	}
}

func TestHandler_UpdateStatus_InternalServerError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo, nil, nil, nil, nil, nil),
	)

	req := newAuthenticatedRequest(
		http.MethodPatch,
		"/applications/1/status",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
		50,
		"client",
	)
	req = addRouteParam(req, "id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateStatus(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
