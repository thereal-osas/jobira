package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	subscriptionaccess "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func authenticatedRequest(
	method string,
	target string,
	body string,
	user identity.UserIdentity,
) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	ctx := identity.WithUser(req.Context(), user)

	return req.WithContext(ctx)
}

func requestWithURLParam(
	method string,
	target string,
	body string,
	paramName string,
	paramValue string,
) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(paramName, paramValue)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	)

	return req.WithContext(ctx)
}

func authenticatedRequestWithURLParam(
	method string,
	target string,
	body string,
	paramName string,
	paramValue string,
	user identity.UserIdentity,
) *http.Request {
	req := requestWithURLParam(
		method,
		target,
		body,
		paramName,
		paramValue,
	)

	ctx := identity.WithUser(req.Context(), user)

	return req.WithContext(ctx)
}

func TestNewHandler(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service to be assigned")
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			job.ID = 20
			return nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequest(
		http.MethodPost,
		"/jobs",
		`{
			"title":"Domestic cleaner needed",
			"description":"Clean a two-bedroom flat",
			"location":"East London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":80
		}`,
		identity.UserIdentity{
			UserID: 5,
			Email:  "client@example.com",
			Role:   "client",
		},
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

	var job Job

	if err := json.NewDecoder(rec.Body).Decode(&job); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if job.ID != 20 {
		t.Fatalf("expected job ID 20, got %d", job.ID)
	}

	if job.ClientID != 5 {
		t.Fatalf("expected client ID 5, got %d", job.ClientID)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/jobs",
		strings.NewReader(`{}`),
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

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := authenticatedRequest(
		http.MethodPost,
		"/jobs",
		`{"title":`,
		identity.UserIdentity{
			UserID: 5,
			Role:   "client",
		},
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

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := authenticatedRequest(
		http.MethodPost,
		"/jobs",
		`{
			"title":"",
			"description":"Clean flat",
			"location":"London",
			"job_type":"domestic"
		}`,
		identity.UserIdentity{
			UserID: 5,
			Role:   "client",
		},
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Create_DailyLimit(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			return subscriptionaccess.ErrDailyJobPostLimit
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequest(
		http.MethodPost,
		"/jobs",
		`{
			"title":"Cleaner needed",
			"description":"Clean flat",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":60
		}`,
		identity.UserIdentity{
			UserID: 5,
			Role:   "client",
		},
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			return expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequest(
		http.MethodPost,
		"/jobs",
		`{
			"title":"Cleaner needed",
			"description":"Clean flat",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":60
		}`,
		identity.UserIdentity{
			UserID: 5,
			Role:   "client",
		},
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetByID_Success(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
				Title:    "Cleaner needed",
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := requestWithURLParam(
		http.MethodGet,
		"/jobs/12",
		"",
		"id",
		"12",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	var job Job

	if err := json.NewDecoder(rec.Body).Decode(&job); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if job.ID != 12 {
		t.Fatalf("expected job ID 12, got %d", job.ID)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := requestWithURLParam(
		http.MethodGet,
		"/jobs/not-a-number",
		"",
		"id",
		"not-a-number",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_GetByID_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := requestWithURLParam(
		http.MethodGet,
		"/jobs/0",
		"",
		"id",
		"0",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetByID_InternalServerError(t *testing.T) {
	expected := errors.New("query failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return nil, expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := requestWithURLParam(
		http.MethodGet,
		"/jobs/12",
		"",
		"id",
		"12",
	)

	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_List_Success(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(ctx context.Context) ([]Job, error) {
			return []Job{
				{ID: 1, Title: "First job"},
				{ID: 2, Title: "Second job"},
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	var jobs []Job

	if err := json.NewDecoder(rec.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestHandler_List_InternalServerError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		ListFunc: func(ctx context.Context) ([]Job, error) {
			return nil, expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &MockRepository{
		ListByClientIDFunc: func(
			ctx context.Context,
			clientID uint,
		) ([]Job, error) {
			if clientID != 5 {
				t.Fatalf("expected client ID 5, got %d", clientID)
			}

			return []Job{
				{ID: 1, ClientID: clientID},
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequest(
		http.MethodGet,
		"/jobs/mine",
		"",
		identity.UserIdentity{
			UserID: 5,
			Role:   "client",
		},
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

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := httptest.NewRequest(http.MethodGet, "/jobs/mine", nil)
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

func TestHandler_ListMine_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(&MockRepository{}, nil, nil),
	)

	req := authenticatedRequest(
		http.MethodGet,
		"/jobs/mine",
		"",
		identity.UserIdentity{
			UserID: 0,
			Role:   "client",
		},
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

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		ListByClientIDFunc: func(
			ctx context.Context,
			clientID uint,
		) ([]Job, error) {
			return nil, expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequest(
		http.MethodGet,
		"/jobs/mine",
		"",
		identity.UserIdentity{
			UserID: 5,
			Role:   "client",
		},
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

func TestHandler_Update_Success(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
				Status:   "open",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, job *Job) error {
			return nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/12",
		`{
			"title":"Updated",
			"description":"Updated description",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":120,
			"status":"open"
		}`,
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestHandler_Update_Unauthorized(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil))

	req := httptest.NewRequest(http.MethodPut, "/jobs/1", nil)
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/x",
		`{}`,
		"id",
		"x",
		identity.UserIdentity{UserID: 5},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/12",
		`{"title":`,
		"id",
		"12",
		identity.UserIdentity{UserID: 5},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint) error {
			return nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodDelete,
		"/jobs/12",
		"",
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestHandler_Delete_Unauthorized(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil))

	req := httptest.NewRequest(http.MethodDelete, "/jobs/1", nil)
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestHandler_Search_Success(t *testing.T) {
	repo := &MockRepository{
		SearchFunc: func(ctx context.Context, req SearchJobRequest) ([]Job, error) {
			if req.MinBudget != 100 {
				t.Fatalf("expected min budget 100 got %d", req.MinBudget)
			}

			return []Job{
				{ID: 1},
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/jobs/search?location=London&job_type=domestic&listing_type=shift&min_budget=100",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d got %d (%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestHandler_Search_InvalidBudgetDefaultsToZero(t *testing.T) {
	repo := &MockRepository{
		SearchFunc: func(ctx context.Context, req SearchJobRequest) ([]Job, error) {
			if req.MinBudget != 0 {
				t.Fatalf("expected zero budget got %d", req.MinBudget)
			}
			return []Job{}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/jobs/search?min_budget=abc",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, rec.Code)
	}
}

func TestParseIDParam(t *testing.T) {
	req := requestWithURLParam(
		http.MethodGet,
		"/jobs/55",
		"",
		"id",
		"55",
	)

	id, err := parseIDParam(req, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 55 {
		t.Fatalf("expected 55 got %d", id)
	}
}

func TestParseIDParam_Invalid(t *testing.T) {
	req := requestWithURLParam(
		http.MethodGet,
		"/jobs/x",
		"",
		"id",
		"x",
	)

	_, err := parseIDParam(req, "id")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandler_Update_InvalidInput(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
				Status:   "open",
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/12",
		`{
			"title":"",
			"description":"Updated description",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":100,
			"status":"open"
		}`,
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return nil, ErrJobNotFound
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/99",
		`{
			"title":"Updated title",
			"description":"Updated description",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":100,
			"status":"open"
		}`,
		"id",
		"99",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Update_Forbidden(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 50,
				Status:   "open",
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/12",
		`{
			"title":"Updated title",
			"description":"Updated description",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":100,
			"status":"open"
		}`,
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Update_InternalServerError(t *testing.T) {
	expected := errors.New("update failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
				Status:   "open",
			}, nil
		},
		UpdateFunc: func(ctx context.Context, job *Job) error {
			return expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodPut,
		"/jobs/12",
		`{
			"title":"Updated title",
			"description":"Updated description",
			"location":"London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":100,
			"status":"open"
		}`,
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodDelete,
		"/jobs/invalid",
		"",
		"id",
		"invalid",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Delete_InvalidInput(t *testing.T) {
	handler := NewHandler(NewService(&MockRepository{}, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodDelete,
		"/jobs/0",
		"",
		"id",
		"0",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return nil, ErrJobNotFound
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodDelete,
		"/jobs/99",
		"",
		"id",
		"99",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Delete_Forbidden(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 50,
			}, nil
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodDelete,
		"/jobs/12",
		"",
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Delete_InternalServerError(t *testing.T) {
	expected := errors.New("delete failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint) error {
			return expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := authenticatedRequestWithURLParam(
		http.MethodDelete,
		"/jobs/12",
		"",
		"id",
		"12",
		identity.UserIdentity{
			UserID: 5,
			Role:   "user",
		},
	)

	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Search_InternalServerError(t *testing.T) {
	expected := errors.New("search failed")

	repo := &MockRepository{
		SearchFunc: func(
			ctx context.Context,
			req SearchJobRequest,
		) ([]Job, error) {
			return nil, expected
		},
	}

	handler := NewHandler(NewService(repo, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/jobs/search?location=London",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}
