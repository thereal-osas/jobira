package adminsubscriptions

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
	r.Route("/admin/subscriptions", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminOnlyMiddleware)

		r.Get("/plans", handler.ListPlans)
		r.Post("/plans", handler.CreatePlan)
		r.Put("/plans/{planID}", handler.UpdatePlan)
		r.Delete("/plans/{planID}", handler.DisablePlan)

		r.Get("/users", handler.ListUserSubscriptions)
		r.Get("/users/{userID}", handler.GetUserSubscriptions)
		r.Patch("/users/{userID}", handler.UpdateUserSubscription)
	})
}