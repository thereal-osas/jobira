package jobinvitations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/job-invitations", func (r chi.Router)  {
		r.Use(authMiddleware)

		r.Post("/jobs/{jobID}", handler.Create)
		r.Get("/sent", handler.ListSent)
		r.Get("/me", handler.ListReceived)
	})
}
