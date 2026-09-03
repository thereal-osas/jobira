package availablenow

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/available-now", func(r chi.Router) {
		//Public/client-facing search.
		r.Get("/cleaners", handler.Search)

		//Cleaner controls
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Post("/me", handler.SetMine)
			r.Get("/me", handler.GetMine)
			r.Delete("/me", handler.DisableMine)

		})
	})
}
