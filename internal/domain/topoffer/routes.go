package topoffer

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/top-offers", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/jobs/{jobID}",
			handler.ListJobOffers,
		)

		r.Get("/applications/{applicationID}/mine",
			handler.GetMyApplicationRanking)
	})
}
