package subscriptionaccess

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func requestWithIdentity(
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

func TestHandler_CleanerMe_Success(t *testing.T) {
	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return &CleanerAccessStatus{
				UserID:                userID,
				SubscriptionStatus:    "active",
				Premium:               true,
				CanViewJobs:           true,
				CanApply:              true,
				CanViewClientDetails:  true,
				HasInstantAlerts:      true,
				HasPriorityPlacement:  true,
				HasPremiumProfile:     true,
				HasEarlyJobAccess:     true,
				ApplicationCount:      2,
				FreeApplicationLimit:  5,
				ApplicationsToday:     1,
				DailyApplicationLimit: 5,
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/me",
		nil,
	)

	req = requestWithIdentity(
		req,
		7,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.CleanerMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_CleanerMe_Unauthorized(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.CleanerMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_CleanerMe_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/me",
		nil,
	)

	req = requestWithIdentity(
		req,
		0,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.CleanerMe(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_CleanerMe_InternalServerError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/me",
		nil,
	)

	req = requestWithIdentity(
		req,
		7,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.CleanerMe(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_ClientMe_Success(t *testing.T) {
	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return &ClientAccessStatus{
				UserID:              userID,
				SubscriptionStatus:  "active",
				Premium:             true,
				CanPostJob:          true,
				CanViewApplicants:   true,
				CanContactCleaners:  true,
				CanUseRepeatBooking: true,
				JobsPostedToday:     1,
				DailyJobPostLimit:   5,
				FreeJobPostLimit:    5,
				JobPostCount:        2,
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/client/me",
		nil,
	)

	req = requestWithIdentity(
		req,
		11,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ClientMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ClientMe_Unauthorized(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/client/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ClientMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_ClientMe_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/client/me",
		nil,
	)

	req = requestWithIdentity(
		req,
		0,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ClientMe(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_ClientMe_InternalServerError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscription-access/client/me",
		nil,
	)

	req = requestWithIdentity(
		req,
		11,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.ClientMe(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
