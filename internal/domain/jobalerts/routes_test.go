package jobalerts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func jobAlertsAuthMiddleware(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: userID,
					Role:   role,
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func TestRegisterRoutes_Create(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			context.Context,
			*JobAlert,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		jobAlertsAuthMiddleware(5, "user"),
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

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMine(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]JobAlert, error) {
			return []JobAlert{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		jobAlertsAuthMiddleware(5, "user"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetByID(t *testing.T) {
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
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		jobAlertsAuthMiddleware(5, "user"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-alerts/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_Update(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			getCalls++

			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		updateFn: func(
			context.Context,
			*JobAlert,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		jobAlertsAuthMiddleware(5, "user"),
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

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_Delete(t *testing.T) {
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
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		jobAlertsAuthMiddleware(5, "user"),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/job-alerts/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
