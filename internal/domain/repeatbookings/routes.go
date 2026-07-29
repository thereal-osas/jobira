package repeatbookings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/repeat-bookings", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/cleaners/{cleanerID}", handler.Create)
		r.Get("/me", handler.ListMine)
		r.Post("/bookings/{bookingID}/book-again", handler.BookAgain)
	})
}
