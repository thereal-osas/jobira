package chat

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, role, err := getAuthUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateConversationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	conversation, err := h.service.CreateConversation(r.Context(), userID, role, req)
	if err != nil {
		h.handlerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, ConversationResponse{
		Conversation: *conversation,
	},
	)
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, role, err := getAuthUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unautorized")
		return
	}

	conversationID, err := parseUintParam(
		chi.URLParam(r, "conversationID"),
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	conversation, err := h.service.GetConversation(
		r.Context(),
		conversationID,
		userID,
		role,
	)
	if err != nil {
		log.Printf("GetMessages error: %+v\n", err)
		h.handlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ConversationResponse{
		Conversation: *conversation,
	},
	)
}

func (h *Handler) ListConversations(
	w http.ResponseWriter, r *http.Request,
) {
	userID, _, err := getAuthUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversations, err := h.service.ListConversations(
		r.Context(), userID,
	)
	if err != nil {
		h.handlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ConversationListResponse{
		Conversations: conversations,
	},
	)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, role, err := getAuthUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := parseUintParam(
		chi.URLParam(r, "conversationID"),
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	var req SendMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	message, err := h.service.SendMessage(r.Context(), conversationID, userID, role, req)
	if err != nil {
		log.Printf("SendMessage error: %+v\n", err)
		h.handlerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, MessageResponse{
		Message: *message,
	},
	)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, role, err := getAuthUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := parseUintParam(
		chi.URLParam(r, "conversationID"),
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	limit := parseIntQuery(r.URL.Query().Get("limit"), 50)
	offset := parseIntQuery(r.URL.Query().Get("offset"), 0)

	messages, err := h.service.ListMessages(
		r.Context(),
		conversationID,
		userID,
		role,
		limit,
		offset,
	)
	if err != nil {
		h.handlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, MessageListResponse{
		Messages: messages,
	},
	)
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, role, err := getAuthUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := parseUintParam(
		chi.URLParam(r, "conversationID"),
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	err = h.service.MarkAsRead(
		r.Context(),
		conversationID,
		userID,
		role,
	)
	if err != nil {
		h.handlerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "conversation marked as read",
	},
	)
}

func (h *Handler) handlerError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrConversationNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrMessageNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	if errors.Is(err, ErrConversationExists) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidMessage) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrMessageTooLong) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeError(w, http.StatusInternalServerError, "internal server error")
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, ErrInvalidInput
	}

	return uint(parsed), nil
}

func parseIntQuery(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getAuthUser(r *http.Request) (uint, string, error) {
	user, err := identity.FromContext(r.Context())
	if err != nil {
		return 0, "", ErrForbidden
	}

	if user.UserID == 0 || user.Role == "" {
		return 0, "", ErrForbidden
	}

	return user.UserID, user.Role, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	},
	)
}
