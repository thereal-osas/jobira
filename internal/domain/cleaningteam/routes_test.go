package cleaningteam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func cleaningTeamAuthMiddlewareForTest(
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

func TestCleaningTeamRoutes_GetMine(
	t *testing.T,
) {
	repo := &mockRepository{
		listTeamMembersFn: func(
			context.Context,
			uint,
		) ([]TeamMemberData, error) {
			return []TeamMemberData{
				{
					CleanerID:   20,
					CleanerName: "Maria",

					ServicesOffered: "housekeeping",

					LastJobType: "housekeeping",

					IsPreferred: true,
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
		cleaningTeamAuthMiddlewareForTest,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaning-team/me",
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

func TestCleaningTeamRoutes_NotFound(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	handler := NewHandler(
		service,
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleaningTeamAuthMiddlewareForTest,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaning-team/not-a-route",
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
