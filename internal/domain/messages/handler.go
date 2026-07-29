package messages

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
	
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	currentUser,  err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID, err := parseJobID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	var req SendMessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	message, err := h.service.Send(r.Context(),jobID,currentUser.UserID, req)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return	
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, message)

}


func (h *Handler) ListConversation(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID, err := parseJobID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}
	messages, err := h.service.ListConversation(r.Context(), jobID, currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, messages)
}

func parseJobID(r *http.Request) (uint, error) {
	rawID := chi.URLParam(r, "jobID")

	parsedID, err := strconv.ParseUint(rawID,  10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
}