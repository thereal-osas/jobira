package reviews

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/reviews", func(r chi.Router) {
		r.Get("/cleaners/{cleanerID}", handler.ListByCleaner)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Post("/jobs/{jobID}", handler.Create)
			r.Get("/me", handler.ListMine)
		})
	})
}
