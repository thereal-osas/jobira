package wokrproof

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route(
		"/work-proof",
		func(r chi.Router) {
			// Public cleaner verified-work summary.
			r.Get(
				"/cleaners/{cleanerID}",
				handler.GetVerifiedWork,
			)

			// Private booking proof routes.
			r.Group(
				func(r chi.Router) {
					r.Use(
						authMiddleware,
					)

					r.Post(
						"/bookings/{bookingID}",
						handler.Create,
					)

					r.Get(
						"/bookings/{bookingID}",
						handler.GetBookingProof,
					)

					r.Delete(
						"/{proofID}",
						handler.Delete,
					)
				},
			)
		},
	)
}
