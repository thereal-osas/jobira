package preferredcleaners

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func preferredCleanersAuthMiddleware(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: userID,
					Role:   role,
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func TestRegisterRoutes_Create(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*PreferredCleaners,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		preferredCleanersAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/preferred-cleaners/",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMine(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]PreferredCleaners, error) {
			return []PreferredCleaners{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		preferredCleanersAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/preferred-cleaners/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_Remove(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		preferredCleanersAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/8",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
