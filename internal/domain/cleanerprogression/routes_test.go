package cleanerprogression

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func cleanerProgressionTestAuth(
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
					UserID: 10,
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

func TestRegisterRoutes_GetMine(t *testing.T) {
	repo := &mockRepository{
		getCleanerReputationFn: func(
			context.Context,
			uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID:                10,
				AverageRating:            4.9,
				TotalReviews:             15,
				CompletedJobs:            35,
				RepeatClients:            5,
				RecommendationPercentage: 94,
				ReliabilityScore:         92,
				TotalBookings:            40,
				CleanerCancellations:     2,
				EligibleResponseMessages: 30,
				RespondedMessages:        27,
				Badge:                    "Top Rated",
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
		cleanerProgressionTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-progression/me",
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
		"/cleaner-progression/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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
