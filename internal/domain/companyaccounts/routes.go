package companyaccounts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/companies", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", handler.CreateCompany)
		r.Post("/{companyID}/members", handler.AddMember)
		r.Get("/{companyID}/members", handler.ListMembers)
		r.Delete("/{companyID}/members/{userID}", handler.RemoveMember)
	})
}
