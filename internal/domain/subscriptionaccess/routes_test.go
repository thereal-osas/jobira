package subscriptionaccess

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func subscriptionAccessTestAuth(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
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
			},
		)
	}
}

func TestRegisterRoutes_CleanerMe(t *testing.T) {
	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return &CleanerAccessStatus{
				UserID:             userID,
				SubscriptionStatus: "none",
				CanViewJobs:        true,
				CanApply:           true,
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		subscriptionAccessTestAuth(
			7,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_ClientMe(t *testing.T) {
	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return &ClientAccessStatus{
				UserID:             userID,
				SubscriptionStatus: "none",
				CanPostJob:         true,
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		subscriptionAccessTestAuth(
			11,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/client/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_CleanerMe_MethodNotAllowed(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		subscriptionAccessTestAuth(
			7,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscription-access/me",
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

func TestRegisterRoutes_ClientMe_MethodNotAllowed(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		subscriptionAccessTestAuth(
			11,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscription-access/client/me",
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

	RegisterRoutes(
		router,
		handler,
		subscriptionAccessTestAuth(
			7,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access",
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
