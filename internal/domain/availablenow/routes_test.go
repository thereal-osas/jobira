package availablenow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func availableNowTestAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: 10,
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func TestRegisterRoutes_SetMine(t *testing.T) {
	repo := &mockRepository{
		upsertFn: func(
			context.Context,
			*CleanerAvailableNow,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		availableNowTestAuth,
	)

	body := `{
		"duration_minutes": 120,
		"location": "Stratford",
		"travel_radius_miles": 8,
		"job_types": ["domestic"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(body),
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

func TestRegisterRoutes_GetMine(t *testing.T) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailableNow, error) {
			return &CleanerAvailableNow{
				CleanerID:      10,
				IsAvailable:    true,
				AvailableFrom:  now.Add(-time.Hour),
				AvailableUntil: now.Add(time.Hour),
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		availableNowTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/me",
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

func TestRegisterRoutes_DisableMine(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		availableNowTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/available-now/me",
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

func TestRegisterRoutes_SearchPublic(t *testing.T) {
	repo := &mockRepository{
		listAvailableCleanersFn: func(
			context.Context,
			AvailableNowSearchRequest,
			time.Time,
		) ([]AvailableCleaner, error) {
			return []AvailableCleaner{
				{
					CleanerID: 10,
					FullName:  "Sarah Cleaner",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		availableNowTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/cleaners",
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

func TestRegisterRoutes_MeRequiresIdentity(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
