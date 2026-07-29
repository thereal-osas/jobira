package reputation

import "github.com/go-chi/chi/v5"

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
) {
	r.Route("/reputation", func(r chi.Router) {
		r.Get("/{cleanerID}", handler.GetByCleanerID)
	})
}
