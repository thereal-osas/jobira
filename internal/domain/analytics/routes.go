package analytics

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
	adminOnlyMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/admin/analytics", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminOnlyMiddleware)

		r.Get("/dashboard", handler.Dashboard)
	})
}
