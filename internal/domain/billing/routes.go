package billing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/billing", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/checkout", handler.CreateCheckoutSession)
		r.Post("/portal", handler.CreateBillingPortalSession)
		r.Post("/promo/validate", handler.ValidatePromoCode)
	})

	r.Post("/stripe/webhook", handler.StripeWebhook)
}
