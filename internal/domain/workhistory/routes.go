package workhistory

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, 
	handler *Handler, 
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route(
		"/work-history",
		func(r chi.Router) {
			r.Use(authMiddleware)

			r.Get(
				"/me",
				handler.GetMine,
			)

			r.Get(
				"/bookings/{bookingID}",
				handler.GetBookingHistory,
			)
		},
	)
}