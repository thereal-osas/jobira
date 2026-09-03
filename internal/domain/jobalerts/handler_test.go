package jobalerts

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newHandlerForTest(repo Repository) *Handler {
	service := NewService(repo)

	return NewHandler(service)
}

func requestWithUser(
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

func requestWithURLParam(
	req *http.Request,
	name string,
	value string,
) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(name, value)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
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

func TestParseIDParam_Success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)

	id, err := parseIDParam(req, "alertID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			id,
		)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"nope",
	)

	id, err := parseIDParam(req, "alertID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf(
			"expected ID 0, got %d",
			id,
		)
	}
}

func TestParseIDParam_MissingID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/",
		nil,
	)

	id, err := parseIDParam(req, "alertID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf(
			"expected ID 0, got %d",
			id,
		)
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			alert *JobAlert,
		) error {
			if alert.UserID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					alert.UserID,
				)
			}

			if alert.Location != "East London" {
				t.Fatalf(
					"expected East London, got %q",
					alert.Location,
				)
			}

			if alert.JobType != "domestic" {
				t.Fatalf(
					"expected domestic, got %q",
					alert.JobType,
				)
			}

			if alert.MinimumBudget != 80 {
				t.Fatalf(
					"expected minimum budget 80, got %d",
					alert.MinimumBudget,
				)
			}

			alert.ID = 12

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-alerts/",
		strings.NewReader(`{
			"location":"East London",
			"job_type":"domestic",
			"minimum_budget":80
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"user_id":5`,
	) {
		t.Fatalf(
			"expected user ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"location":"East London"`,
	) {
		t.Fatalf(
			"expected location in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"job_type":"domestic"`,
	) {
		t.Fatalf(
			"expected job type in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-alerts/",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-alerts/",
		strings.NewReader(`{"location":`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-alerts/",
		strings.NewReader(`{
			"location":"",
			"job_type":"domestic",
			"minimum_budget":80
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_ZeroUserID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-alerts/",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80
		}`),
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	expectedErr := errors.New("create alert failed")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*JobAlert,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-alerts/",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

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
	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			id uint,
		) (*JobAlert, error) {
			if id != 12 {
				t.Fatalf(
					"expected alert ID 12, got %d",
					id,
				)
			}

			return &JobAlert{
				ID:            12,
				UserID:        5,
				Location:      "London",
				JobType:       "domestic",
				MinimumBudget: 80,
				IsActive:      true,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"id":12`,
	) {
		t.Fatalf(
			"expected alert ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_AdminSuccess(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"admin",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetByID_InvalidAlertID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, ErrAlertNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_InternalServerError(t *testing.T) {
	expectedErr := errors.New("get alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

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
	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]JobAlert, error) {
			return []JobAlert{
				{
					ID:       1,
					UserID:   5,
					Location: "London",
					JobType:  "domestic",
					IsActive: true,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
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

	if !strings.Contains(
		recorder.Body.String(),
		`"location":"London"`,
	) {
		t.Fatalf(
			"expected alert in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]JobAlert, error) {
			return []JobAlert{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
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
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/me",
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
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/me",
		nil,
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list alerts failed")

	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]JobAlert, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			getCalls++

			if getCalls == 1 {
				return &JobAlert{
					ID:     12,
					UserID: 5,
				}, nil
			}

			return &JobAlert{
				ID:            12,
				UserID:        5,
				Location:      "East London",
				JobType:       "airbnb",
				MinimumBudget: 120,
				IsActive:      false,
			}, nil
		},
		updateFn: func(
			ctx context.Context,
			alert *JobAlert,
		) error {
			if alert.Location != "East London" {
				t.Fatalf(
					"expected East London, got %q",
					alert.Location,
				)
			}

			if alert.JobType != "airbnb" {
				t.Fatalf(
					"expected airbnb, got %q",
					alert.JobType,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"East London",
			"job_type":"airbnb",
			"minimum_budget":120,
			"is_active":false
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"location":"East London"`,
	) {
		t.Fatalf(
			"expected updated alert in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_AdminSuccess(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			getCalls++

			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80,
			"is_active":true
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"admin",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Update_InvalidAlertID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/nope",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{"location":`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InvalidInput(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"",
			"job_type":"domestic",
			"minimum_budget":80,
			"is_active":true
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_ZeroUserID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80,
			"is_active":true
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, ErrAlertNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80,
			"is_active":true
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80,
			"is_active":true
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InternalServerError(t *testing.T) {
	expectedErr := errors.New("update alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		updateFn: func(
			context.Context,
			*JobAlert,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/job-alerts/12",
		strings.NewReader(`{
			"location":"London",
			"job_type":"domestic",
			"minimum_budget":80,
			"is_active":true
		}`),
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		deleteFn: func(
			ctx context.Context,
			id uint,
		) error {
			if id != 12 {
				t.Fatalf(
					"expected alert ID 12, got %d",
					id,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"message":"job alert deleted"`,
	) {
		t.Fatalf(
			"expected success message: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_AdminSuccess(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"admin",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Delete_InvalidAlertID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, ErrAlertNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_InternalServerError(t *testing.T) {
	expectedErr := errors.New("delete alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		deleteFn: func(
			context.Context,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"alertID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
