package chat

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	router.Route("/chat", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/conversations", handler.CreateConversation)
		r.Get("/conversations", handler.ListConversations)

		r.Get(
			"/conversations/{conversationID}",
			handler.GetConversation,
		)

		r.Post(
			"/conversations/{conversationID}/messages",
			handler.SendMessage,
		)

		r.Get(
			"/conversations/{conversationID}/messages",
			handler.ListMessages,
		)

		r.Post(
			"/conversations/{conversationID}/read",
			handler.MarkAsRead,
		)
	})
}
