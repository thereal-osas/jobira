package matchscore

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newMatchScoreHandlerForTest(
	repo Repository,
) *Handler {
	service := NewService(repo)

	return NewHandler(service)
}
func requestWithMatchScoreUser(
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

func requestWithMatchScoreJobID(
	method string,
	target string,
	jobID string,
) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		nil,
	)

	routeContext := chi.NewRouteContext()

	routeContext.URLParams.Add(
		"jobID",
		jobID,
	)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
	)

	return req.WithContext(ctx)
}

func TestHandler_GetJobMatches_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			_ context.Context,
			jobID uint,
			clientID uint,
		) (bool, error) {
			if jobID != 10 {
				t.Fatalf(
					"expected job ID 10, got %d",
					jobID,
				)
			}

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return true, nil
		},

		listCandidatesFn: func(
			_ context.Context,
			jobID uint,
			clientID uint,
		) ([]CandidateData, error) {
			return []CandidateData{
				{
					ApplicationID: 1,
					JobID:         jobID,
					ClientID:      clientID,
					CleanerID:     20,
					CleanerName:   "Maria Cleaner",

					JobType: "end_of_tenancy",

					JobLocation: "East London",

					CleanerLocation: "East London",

					ServicesOffered: "end of tenancy",

					AvailabilityStatus: "available",

					IsVerified: true,

					ReliabilityScore: 95,

					ResponseRate: 95,

					AverageResponseMinutes: 10,

					AverageRating: 4.9,
					TotalReviews:  25,

					Badge: "Elite Cleaner",
				},
			}, nil
		},
	}

	handler := newMatchScoreHandlerForTest(
		repo,
	)

	req := requestWithMatchScoreJobID(
		http.MethodGet,
		"/match-score/jobs/10",
		"10",
	)

	req = requestWithMatchScoreUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobMatches(
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

	var result JobMatches

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if result.JobID != 10 {
		t.Fatalf(
			"expected job ID 10, got %d",
			result.JobID,
		)
	}

	if result.TotalApplicants != 1 {
		t.Fatalf(
			"expected 1 applicant, got %d",
			result.TotalApplicants,
		)
	}

	if result.TopOffer == nil {
		t.Fatal(
			"expected Top Offer",
		)
	}

	if result.TopOffer.CleanerID != 20 {
		t.Fatalf(
			"expected cleaner 20, got %d",
			result.TopOffer.CleanerID,
		)
	}

	if !result.TopOffer.IsTopOffer {
		t.Fatal(
			"expected cleaner to be Top Offer",
		)
	}

	if len(result.TopOffer.WhyYoureSeeingThis) == 0 {
		t.Fatal(
			"expected why you're seeing this reasons",
		)
	}

	expectedReasonCodes := map[string]bool{
		"service_fit":  false,
		"availability": false,
		"location":     false,
		"reliability":  false,
		"verification": false,
		"response":     false,
	}

	for _, reason := range result.TopOffer.WhyYoureSeeingThis {
		if _, ok := expectedReasonCodes[reason.Code]; ok {
			expectedReasonCodes[reason.Code] = true
		}

		if reason.Points <= 0 {
			t.Fatalf(
				"expected positive points for reason %q, got %d",
				reason.Code,
				reason.Points,
			)
		}

		if reason.Label == "" {
			t.Fatalf(
				"expected label for reason %q",
				reason.Code,
			)
		}

		if reason.Message == "" {
			t.Fatalf(
				"expected message for reason %q",
				reason.Code,
			)
		}
	}

	for code, found := range expectedReasonCodes {
		if !found {
			t.Fatalf(
				"expected reason %q in response",
				code,
			)
		}
	}

}

func TestHandler_GetJobMatches_Unauthorized(
	t *testing.T,
) {
	handler := newMatchScoreHandlerForTest(
		&mockRepository{},
	)

	req := requestWithMatchScoreJobID(
		http.MethodGet,
		"/match-score/jobs/10",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobMatches(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetJobMatches_InvalidJobID(
	t *testing.T,
) {
	handler := newMatchScoreHandlerForTest(
		&mockRepository{},
	)

	req := requestWithMatchScoreJobID(
		http.MethodGet,
		"/match-score/jobs/nope",
		"nope",
	)

	req = requestWithMatchScoreUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobMatches(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_GetJobMatches_Forbidden(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	handler := newMatchScoreHandlerForTest(
		repo,
	)

	req := requestWithMatchScoreJobID(
		http.MethodGet,
		"/match-score/jobs/10",
		"10",
	)

	req = requestWithMatchScoreUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobMatches(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetJobMatches_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	handler := newMatchScoreHandlerForTest(
		repo,
	)

	req := requestWithMatchScoreJobID(
		http.MethodGet,
		"/match-score/jobs/10",
		"10",
	)

	req = requestWithMatchScoreUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobMatches(
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
