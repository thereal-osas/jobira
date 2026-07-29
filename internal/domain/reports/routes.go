package reports

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) {

	r.Route("/reports", func(r chi.Router) {

		r.Use(authMiddleware)

		r.Post("/", handler.Create)
		r.Get("/me", handler.ListMine)
		r.Get("/{reportID}", handler.GetByID)
	})

	r.Route("/admin/reports", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminMiddleware)

		r.Get("/", handler.ListAll)
		r.Get("/open", handler.ListOpen)
		r.Patch("/{reportID}", handler.Review)
	})
}
