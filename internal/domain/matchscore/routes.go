package matchscore

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
		"/match-score",
		func(r chi.Router) {
			r.Use(authMiddleware)

			r.Get(
				"/jobs/{jobID}",
				handler.GetJobMatches,
			)
		},
	)
}
