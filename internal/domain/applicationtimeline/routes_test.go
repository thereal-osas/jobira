package applicationtimeline

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestRegisterRoutes_GetTimeline(t *testing.T) {
	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
		getCurrentStatusFn: func(
			ctx context.Context,
			applicationID uint,
		) (string, error) {
			return "pending", nil
		},
		listByApplicationIDFn: func(
			ctx context.Context,
			applicationID uint,
		) ([]Event, error) {
			return []Event{
				{
					ID:            1,
					ApplicationID: applicationID,
					Status:        "pending",
				},
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				currentUser := identity.UserIdentity{
					UserID: 20,
					Role:   "cleaner",
				}

				ctx := identity.WithUser(
					r.Context(),
					currentUser,
				)

				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_RequiresAuthentication(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
			},
		)
	}

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_WrongMethod(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return next
	}

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/application-timeline/applications/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_NotFound(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return next
	}

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/unknown/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
