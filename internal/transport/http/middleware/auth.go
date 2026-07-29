package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

func Auth(jwtIssuer *jwtsecurity.Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			claims, err := jwtIssuer.Parse(parts[1])
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			userID, err := strconv.ParseUint(claims.Subject, 10, 64)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid or expire token")
				return
			}

			ctx := identity.WithUser(r.Context(), identity.UserIdentity{
				UserID: uint(userID),
				Email:  claims.Email,
				Role:   claims.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
