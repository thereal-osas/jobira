package preferredcleaners

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler, 
) {
	r.Route("/preferred-cleaners", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", handler.Create)
		r.Get("/me", handler.ListMine)
		r.Delete("/{cleanerID}", handler.Remove)
	})
}

