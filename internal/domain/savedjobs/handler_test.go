package savedjobs

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
		http.MethodDelete,
		"/saved-jobs/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)

	id, err := parseIDParam(req, "jobID")
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
		http.MethodDelete,
		"/saved-jobs/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"nope",
	)

	id, err := parseIDParam(req, "jobID")

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
		http.MethodDelete,
		"/saved-jobs/",
		nil,
	)

	id, err := parseIDParam(req, "jobID")

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

func TestHandler_Save_Success(t *testing.T) {
	repo := &mockRepository{
		saveFn: func(
			ctx context.Context,
			savedJob *SavedJob,
		) error {
			if savedJob.UserID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					savedJob.UserID,
				)
			}

			if savedJob.JobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					savedJob.JobID,
				)
			}

			savedJob.ID = 20

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{
			"job_id":12
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

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
		`"job_id":12`,
	) {
		t.Fatalf(
			"expected job ID in response: %s",
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
}

func TestHandler_Save_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Save_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{"job_id":`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{
			"job_id":0
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_ZeroUserID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{
			"job_id":12
		}`),
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_AlreadySaved(t *testing.T) {
	repo := &mockRepository{
		saveFn: func(
			context.Context,
			*SavedJob,
		) error {
			return ErrAlreadySved
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{
			"job_id":12
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_InternalServerError(t *testing.T) {
	expectedErr := errors.New("save failed")

	repo := &mockRepository{
		saveFn: func(
			context.Context,
			*SavedJob,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/saved-jobs/",
		strings.NewReader(`{
			"job_id":12
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

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
		) ([]SavedJob, error) {
			return []SavedJob{
				{
					ID:     20,
					UserID: 5,
					JobID:  12,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/saved-jobs/me",
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
		`"job_id":12`,
	) {
		t.Fatalf(
			"expected saved job in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]SavedJob, error) {
			return []SavedJob{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/saved-jobs/me",
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
		"/saved-jobs/me",
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
		"/saved-jobs/me",
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
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]SavedJob, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/saved-jobs/me",
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

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			ctx context.Context,
			userID uint,
			jobID uint,
		) error {
			if userID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					userID,
				)
			}

			if jobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					jobID,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/saved-jobs/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
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
		`"message":"saved job removed"`,
	) {
		t.Fatalf(
			"expected success message: %s",
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
		"/saved-jobs/12",
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

func TestHandler_Delete_InvalidJobID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/saved-jobs/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
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
		"/saved-jobs/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
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
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrSavedJobNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/saved-jobs/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
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

func TestHandler_Delete_InternalServerError(t *testing.T) {
	expectedErr := errors.New("delete failed")

	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/saved-jobs/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"jobID",
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
