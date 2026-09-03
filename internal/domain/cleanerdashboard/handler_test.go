package cleanerdashboard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func cleanerDashboardRequestWithIdentity(
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

func TestHandler_Me_Success(t *testing.T) {
	repo := &mockRepository{
		countFavouritesFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 4, nil
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
				SubscriptionStatus: "active",
				Premium:            true,
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-dashboard/me",
		nil,
	)

	req = cleanerDashboardRequestWithIdentity(
		req,
		7,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_Me_Unauthorized(t *testing.T) {
	service := newDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
		&mockReputationRepository{},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-dashboard/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_Me_InvalidInput(t *testing.T) {
	service := newDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
		&mockReputationRepository{},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-dashboard/me",
		nil,
	)

	req = cleanerDashboardRequestWithIdentity(
		req,
		0,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_Me_InternalServerError(t *testing.T) {
	expectedErr := errors.New("dashboard failure")

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := newDashboardService(
		&mockRepository{},
		accessRepo,
		&mockReputationRepository{},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-dashboard/me",
		nil,
	)

	req = cleanerDashboardRequestWithIdentity(
		req,
		7,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
