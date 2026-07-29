package adminmoderation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router, 
	handler *Handler, 
	authMiddleware func(http.Handler) http.Handler, 
	adminOnlyMiddleware func(http.Handler) http.Handler, 
) {
	r.Route("/admin/moderation", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminOnlyMiddleware)

		r.Get("/reports", handler.ListReports)
		r.Get("/reports/open", handler.ListOpenReports)
		r.Patch("/reports/{reportID}", handler.UpdatedReportStatus)
		r.Get("/blocked-cleaners", handler.ListBlockCleaners)
	})
}