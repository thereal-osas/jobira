package topoffer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func topOfferRequestWithUser(
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

func TestHandler_ListJobOffers_Success(t *testing.T) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				Budget:   100,
				Status:   "open",
			}, nil
		},

		listOfferCandidatesFn: func(
			context.Context,
			uint,
			int,
			int,
		) ([]OfferCandidate, error) {
			return []OfferCandidate{
				{
					ApplicationID:            1,
					JobID:                    10,
					CleanerID:                20,
					CleanerName:              "Sarah",
					ProposedRate:             95,
					ApplicationStatus:        "pending",
					AverageRating:            4.9,
					CompletedJobs:            50,
					ReliabilityScore:         95,
					RecommendationPercentage: 96,
					AverageResponseMinutes:   10,
					IsVerified:               true,
					AppliedAt:                now,
				},
			}, nil
		},

		countOffersFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 1, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/top-offers/jobs/10",
		nil,
	)

	req = topOfferRequestWithUser(
		req,
		5,
		"client",
	)

	req = withTopOfferURLParam(
		req,
		"jobID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.ListJobOffers(
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

func TestHandler_ListJobOffers_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/top-offers/jobs/10",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListJobOffers(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListJobOffers_InvalidJobID(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/top-offers/jobs/nope",
		nil,
	)

	req = topOfferRequestWithUser(
		req,
		5,
		"client",
	)

	req = withTopOfferURLParam(
		req,
		"jobID",
		"nope",
	)

	recorder := httptest.NewRecorder()

	handler.ListJobOffers(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_ListJobOffers_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				Status:   "open",
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := topOfferRequestWithUser(
		httptest.NewRequest(
			http.MethodGet,
			"/top-offers/jobs/10",
			nil,
		),
		99,
		"client",
	)

	req = withTopOfferURLParam(
		req,
		"jobID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.ListJobOffers(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListJobOffers_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := topOfferRequestWithUser(
		httptest.NewRequest(
			http.MethodGet,
			"/top-offers/jobs/10",
			nil,
		),
		5,
		"client",
	)

	req = withTopOfferURLParam(
		req,
		"jobID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.ListJobOffers(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetMyApplicationRanking_Success(
	t *testing.T,
) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getApplicationCandidateFn: func(
			context.Context,
			uint,
		) (*OfferCandidate, error) {
			return &OfferCandidate{
				ApplicationID: 1,
				JobID:         10,
				CleanerID:     20,
			}, nil
		},

		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				Budget:   100,
				Status:   "open",
			}, nil
		},

		listOfferCandidatesFn: func(
			context.Context,
			uint,
			int,
			int,
		) ([]OfferCandidate, error) {
			return []OfferCandidate{
				{
					ApplicationID:            1,
					JobID:                    10,
					CleanerID:                20,
					ProposedRate:             95,
					AverageRating:            4.9,
					CompletedJobs:            50,
					ReliabilityScore:         95,
					RecommendationPercentage: 96,
					AverageResponseMinutes:   10,
					IsVerified:               true,
					AppliedAt:                now,
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := topOfferRequestWithUser(
		httptest.NewRequest(
			http.MethodGet,
			"/top-offers/applications/1/mine",
			nil,
		),
		20,
		"cleaner",
	)

	req = withTopOfferURLParam(
		req,
		"applicationID",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyApplicationRanking(
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

func TestHandler_GetMyApplicationRanking_ForbiddenRole(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := topOfferRequestWithUser(
		httptest.NewRequest(
			http.MethodGet,
			"/top-offers/applications/1/mine",
			nil,
		),
		5,
		"client",
	)

	req = withTopOfferURLParam(
		req,
		"applicationID",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyApplicationRanking(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			recorder.Code,
		)
	}
}

func withTopOfferURLParam(
	req *http.Request,
	key string,
	value string,
) *http.Request {
	ctx := chi.NewRouteContext()

	ctx.URLParams.Add(
		key,
		value,
	)

	return req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			ctx,
		),
	)
}
