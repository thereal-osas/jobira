package availability

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/availability", func(r chi.Router) {
		// PUBLIC ROUTE USED BY CLIENTS VIEWING CLEANER PROFILES.
		r.Get("/cleaners/{cleanerID}", handler.ListByCleaner)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// CLEANER AVAILABILITY CRUD
			r.Post("/", handler.Create)
			r.Get("/me", handler.ListMine)
			r.Get("/{availabilityID}", handler.GetByID)
			r.Put("/{availabilityID}", handler.Update)
			r.Delete("/{availabilityID}", handler.Delete)

			//AVAILABILITY BLOCK MANAGEMENT.
			r.Post("/blocks", handler.CreateBlock)
			r.Get("/blocks/me", handler.ListBlocks)
			r.Delete("/blocks/{blockID}", handler.DeleteBlock)
		})
	})
}
