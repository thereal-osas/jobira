package notifications

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestRegisterRoutes(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: 5,
					Email:  "test@example.com",
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	RegisterRoutes(router, handler, authMiddleware)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "list mine route",
			method: http.MethodGet,
			path:   "/notifications/",
		},
		{
			name:   "list unread route",
			method: http.MethodGet,
			path:   "/notifications/unread",
		},
		{
			name:   "count unread route",
			method: http.MethodGet,
			path:   "/notifications/count",
		},
		{
			name:   "mark all as read route",
			method: http.MethodPatch,
			path:   "/notifications/read-all",
		},
		{
			name:   "mark notification as read route",
			method: http.MethodPatch,
			path:   "/notifications/1/read",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(
				test.method,
				test.path,
				nil,
			)

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusNotFound {
				t.Fatalf(
					"expected route %s %s to be registered",
					test.method,
					test.path,
				)
			}

			if rr.Code == http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected method %s to be registered for %s",
					test.method,
					test.path,
				)
			}
		})
	}
}


