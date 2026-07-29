package bookingtimeline

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/booking-timeline", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/{bookingID}", handler.GetByBookingID)
	})
}