package bookings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/bookings", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", handler.Create)
		r.Get("/me", handler.ListMine)
		r.Get("/{bookingID}", handler.GetByID)
		r.Patch("/{bookingID}/confirm", handler.Confirm)
		r.Patch("/{bookingID}/start", handler.Start)
		r.Patch("/{bookingID}/complete", handler.Complete)
		r.Patch("/{bookingID}/cancel", handler.Cancel)
		r.Patch("/{bookingID}/close", handler.Close)
	})
}
