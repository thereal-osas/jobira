package clientdashboard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func clientDashboardRequestWithIdentity(
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
				SubscriptionStatus: "active",
				Premium:            true,
				CanPostJob:         true,
			}, nil
		},
	}

	service := newClientDashboardService(
		repo,
		accessRepo,
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/client-dashboard/me",
		nil,
	)

	req = clientDashboardRequestWithIdentity(
		req,
		7,
		"client",
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
	service := newClientDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/client-dashboard/me",
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
	service := newClientDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/client-dashboard/me",
		nil,
	)

	req = clientDashboardRequestWithIdentity(
		req,
		0,
		"client",
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
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.ClientAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := newClientDashboardService(
		&mockRepository{},
		accessRepo,
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/client-dashboard/me",
		nil,
	)

	req = clientDashboardRequestWithIdentity(
		req,
		7,
		"client",
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
