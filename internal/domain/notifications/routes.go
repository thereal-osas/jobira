package notifications

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/notifications", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", handler.ListMine)
		r.Get("/unread", handler.ListUnreadMine)
		r.Get("/count", handler.CountUnreadMine)
		r.Patch("/read-all", handler.MarkAllAsRead)
		r.Patch("/{id}/read", handler.MarkAsRead)
	})
}
