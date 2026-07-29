package profiles

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
	r.Route("/profiles", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/search", handler.Search)

		r.Post("/", handler.Create)
		r.Get("/me", handler.GetMine)
		r.Put("/me", handler.Update)
		
		r.Get("/me/history", handler.GetFullHistory)
		r.Get("/me/completed-jobs", handler.GetCompletedJobs)
		r.Get("/me/cancelled-jobs", handler.GetCancelledJobs)

		r.Group(func(r chi.Router) {
			r.Use(adminOnlyMiddleware)

			r.Patch("/{userID}/verification", handler.UpdateVerificationStatus)
		})
	})
}