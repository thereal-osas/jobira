package verifications

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
	r.Route("/verifications", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", handler.Create)
		r.Get("/me", handler.ListMine)
		r.Get("/{requestID}", handler.GetByID)
	})

	r.Route("/admin/verifications", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminOnlyMiddleware)

		r.Get("/", handler.ListAll)
		r.Get("/pending", handler.ListPending)
		r.Patch("/{requestID}", handler.Review)
	})
}
