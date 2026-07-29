package jobalerts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/job-alerts", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", handler.Create)
		r.Get("/me", handler.ListMine)
		r.Get("/{alertID}", handler.GetByID)
		r.Put("/{alertID}", handler.Update)
		r.Delete("/{alertID}", handler.Delete)
	})
}
