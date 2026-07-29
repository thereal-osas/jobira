package availability

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/availability", func(r chi.Router) {
		r.Get("/cleaners/{cleanerID}", handler.ListByCleaner)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Post("/", handler.Create)
			r.Get("/me", handler.ListMine)
			r.Get("/{availabilityID}", handler.GetByID)
			r.Put("/{availabilityID}", handler.Update)
			r.Delete("/{availabilityID}", handler.Delete)
		})
	})
}