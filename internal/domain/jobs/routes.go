package jobs

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/jobs", func(r chi.Router) {
		r.Get("/search", handler.Search)
		r.Get("/", handler.List)
		r.Get("/{id}", handler.GetByID)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Post("/", handler.Create)
			r.Get("/mine", handler.ListMine)
			r.Put("/{id}", handler.Update)
			r.Delete("/{id}", handler.Delete)
		})
	})
}
