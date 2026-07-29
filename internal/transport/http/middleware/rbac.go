package middleware

import (
	"net/http"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

func RequireRole( role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentUser, err := identity.FromContext(r.Context())
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized")
			}

			if currentUser.Role != role {
				response.Error(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}