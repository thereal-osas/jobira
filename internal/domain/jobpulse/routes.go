package jobpulse

import "github.com/go-chi/chi/v5"

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
) {
	r.Route(
		"/job-pulse",
		func(r chi.Router) {
			r.Get(
				"/jobs/{jobID}",
				handler.GetJobPulse,
			)
		},
	)
}