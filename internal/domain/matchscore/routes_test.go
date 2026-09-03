package matchscore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func matchScoreAuthMiddlewareForTest(
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

func TestMatchScoreRoutes_GetJobMatches(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		listCandidatesFn: func(
			context.Context,
			uint,
			uint,
		) ([]CandidateData, error) {
			return []CandidateData{
				{
					ApplicationID: 1,
					JobID:         10,
					ClientID:      5,
					CleanerID:     20,

					JobType: "domestic",

					JobLocation: "London",

					CleanerLocation: "London",

					ServicesOffered: "domestic",

					AvailabilityStatus: "available",

					IsVerified: true,

					ReliabilityScore: 90,

					ResponseRate: 95,

					AverageResponseMinutes: 10,
				},
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		matchScoreAuthMiddlewareForTest,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/match-score/jobs/10",
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

func TestMatchScoreRoutes_NotFound(
	t *testing.T,
) {
	repo := &mockRepository{}

	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		matchScoreAuthMiddlewareForTest,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/match-score/not-a-route",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}
