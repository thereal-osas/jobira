package referrals

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/referrals", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/codes", handler.CreateCode)
		r.Get("/codes/me", handler.ListMyCodes)
		r.Post("/redeem", handler.Redeem)
		r.Get("/redemptions/me", handler.ListMyRedemptions)
	})
}
