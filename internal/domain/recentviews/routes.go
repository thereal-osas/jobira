package recentviews

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/recent-views", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/cleaners/{cleanerID}", handler.RecordView)
		r.Get("/me", handler.ListMine)
	})
}
