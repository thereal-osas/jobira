package blockedcleaners

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/block-cleaners", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/{cleanerID}", handler.Block)
		r.Delete("/{cleanerID}", handler.Unblock)
		r.Get("/me", handler.ListMine)
	})
}