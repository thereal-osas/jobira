package favorites

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/favorites", func (r chi.Router) { 
		r.Use(authMiddleware)

		r.Post("/", handler.Save)
		r.Get("/", handler.ListMine)
		r.Delete("/{cleanerID}", handler.Remove)
	})
}