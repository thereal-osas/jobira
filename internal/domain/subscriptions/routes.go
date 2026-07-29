package subscriptions

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/plans", handler.ListPlans)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Get("/me", handler.GetMine)
			r.Post("/me", handler.CreateOrUpdateMine)
			r.Patch("/me/status", handler.UpdateMineStatus)
		})
	})
}