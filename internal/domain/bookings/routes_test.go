package bookings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestRegisterRoutes(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			ctx context.Context,
			userID uint,
		) ([]Booking, error) {
			return []Booking{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
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

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
	)

	recorder := httptest.NewRecorder()

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings/me",
		nil,
	)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
