package clientnotes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler, 
) {
	r.Route("/cleaners", func (r chi.Router)  {
		r.Use(authMiddleware)

		r.Post("/{cleanerID}/notes", handler.Create)
		r.Get("/{cleanerID}/notes", handler.GetByCleaner)
	})

	r.Route("/notes", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Put("/{noteID}", handler.Update)
		r.Delete("/{noteID}", handler.Delete)

	})
}

