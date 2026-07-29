package subscriptionaccess

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/subscription-access", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/me", handler.CleanerMe)
		r.Get("/client/me", handler.ClientMe )
	})
}