package messages

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/messages", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/jobs/{jobID}", handler.Send)
		r.Get("/jobs/{jobID}", handler.ListConversation)

	})
}