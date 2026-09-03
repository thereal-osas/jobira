package applicationtimeline

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
		"/application-timeline",
		func(r chi.Router) {
			r.Use(authMiddleware)

			r.Get(
				"/applications/{applicationID}",
				handler.GetTimeline,
			)
		},
	)
}
