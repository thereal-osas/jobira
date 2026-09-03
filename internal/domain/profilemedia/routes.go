package profilemedia

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/profile-media", func(r chi.Router) {

		// Public cleaner/profile media.
		r.Get("/users/{userID}", handler.ListByUser)

		// Public company media.
		r.Get("/companies/{companyID}", handler.ListByCompany)

		// Authenticated personal media management.
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Post("/me", handler.CreateMine)
			r.Get("/me", handler.ListMine)
			r.Put("/me/{mediaID}", handler.UpdateMine)
			r.Delete("/me/{mediaID}", handler.DeleteMine)

			r.Post(
				"/companies/{companyID}",
				handler.CreateForCompany,
			)

			r.Put(
				"/companies/{companyID}/{mediaID}",
				handler.UpdateForCompany,
			)

			r.Delete(
				"/companies/{companyID}/{mediaID}",
				handler.DeleteForCompany,
			)
		})
	})
}
