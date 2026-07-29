package applications

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/applications", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/jobs/{jobID}", handler.Apply)
		r.Get("/mine", handler.ListMine)
		r.Get("/jobs/{jobID}", handler.ListForJob)
		r.Get("/{id}", handler.GetByID)
		r.Patch("/{id}/status", handler.UpdateStatus)
		
	})
}