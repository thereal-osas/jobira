package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RequestID := uuid.NewString()
		ctx := context.WithValue(r.Context(), RequestIDKey, RequestID)

		w.Header().Set("X-Request-ID", RequestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
