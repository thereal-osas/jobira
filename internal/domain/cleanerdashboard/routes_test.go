package cleanerdashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func cleanerDashboardTestAuth(
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
		countFavouritesFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 3, nil
		},
		countPreferredClientsFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 2, nil
		},
		listUpcomingBookingsFn: func(
			ctx context.Context,
			cleanerID uint,
			limit int,
		) ([]UpcomingBooking, error) {
			return []UpcomingBooking{}, nil
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return &subscriptionaccessdomain.CleanerAccessStatus{
				UserID:             userID,
				SubscriptionStatus: "none",
				CanApply:           true,
			}, nil
		},
	}

	reputationRepo := &mockReputationRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID: cleanerID,
			}, nil
		},
	}

	service := newDashboardService(
		repo,
		accessRepo,
		reputationRepo,
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleanerDashboardTestAuth(
			7,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-dashboard/me",
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
	service := newDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
		&mockReputationRepository{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleanerDashboardTestAuth(
			7,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-dashboard/me",
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
	service := newDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
		&mockReputationRepository{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleanerDashboardTestAuth(
			7,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-dashboard",
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
