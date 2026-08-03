package availability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestRegisterRoutes_PublicCleanerAvailability(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]CleanerAvailability, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return []CleanerAvailability{
				{
					ID:        1,
					CleanerID: 8,
					Status:    "available",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)
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
		"/availability/cleaners/8",
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

func TestRegisterRoutes_AuthenticatedListMine(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]CleanerAvailability, error) {
			return []CleanerAvailability{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: 8,
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/me",
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

func TestRegisterRoutes_AuthenticatedListBlocks(t *testing.T) {
	repo := &mockRepository{
		listBlocksByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]AvailabilityBlock, error) {
			return []AvailabilityBlock{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: 8,
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/blocks/me",
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
