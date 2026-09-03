package recentviews

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func recentViewsAuthMiddleware(
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

func TestRegisterRoutes_RecordView(t *testing.T) {
	repo := &mockRepository{
		recordViewFn: func(
			context.Context,
			uint,
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
		recentViewsAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/8",
		nil,
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
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RecentlyViewedCleaner, error) {
			return []RecentlyViewedCleaner{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		recentViewsAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me",
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

func TestRegisterRoutes_ProfileViewAnalytics(
	t *testing.T,
) {
	repo := &mockRepository{
		countByCleanerIDFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 30, nil
		},

		countUniqueViewersByCleanerIDFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 15, nil
		},

		countByCleanerIDSinceFn: func(
			context.Context,
			uint,
			time.Time,
		) (int, error) {
			return 6, nil
		},

		countByCleanerIDBetweenFn: func(
			context.Context,
			uint,
			time.Time,
			time.Time,
		) (int, error) {
			return 4, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		recentViewsAuthMiddleware(
			8,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me/analytics",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ProfileViewAnalytics_MethodNotAllowed(
	t *testing.T,
) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		recentViewsAuthMiddleware(
			8,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/me/analytics",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusMethodNotAllowed,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
