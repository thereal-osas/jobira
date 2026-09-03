package clientdashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func clientDashboardTestAuth(
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

func TestRegisterRoutes_Me(t *testing.T) {
	repo := &mockRepository{
		listActiveJobsFn: func(
			ctx context.Context,
			clientID uint,
			limit int,
		) ([]ActiveJob, error) {
			return []ActiveJob{}, nil
		},
		listUpcomingBookingsFn: func(
			ctx context.Context,
			clientID uint,
			limit int,
		) ([]ClientBooking, error) {
			return []ClientBooking{}, nil
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.ClientAccessStatus, error) {
			return &subscriptionaccessdomain.ClientAccessStatus{
				UserID:             userID,
				SubscriptionStatus: "none",
				CanPostJob:         true,
			}, nil
		},
	}

	service := newClientDashboardService(
		repo,
		accessRepo,
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		clientDashboardTestAuth(
			7,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/client-dashboard/me",
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

func TestRegisterRoutes_MethodNotAllowed(t *testing.T) {
	service := newClientDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		clientDashboardTestAuth(
			7,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/client-dashboard/me",
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
	service := newClientDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		clientDashboardTestAuth(
			7,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/client-dashboard",
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
