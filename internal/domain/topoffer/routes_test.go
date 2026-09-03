package topoffer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func topOfferClientAuth(
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
					UserID: 5,
					Role:   "client",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func topOfferCleanerAuth(
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
					UserID: 20,
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

func TestRegisterRoutes_ListJobOffers(t *testing.T) {
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

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		topOfferClientAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/top-offers/jobs/10",
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

func TestRegisterRoutes_GetMyApplicationRanking(
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

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		topOfferCleanerAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/top-offers/applications/1/mine",
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

func TestRegisterRoutes_RequiresAuth(t *testing.T) {
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
		"/top-offers/jobs/10",
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
