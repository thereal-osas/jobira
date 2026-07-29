package blockedcleaners

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
	return &Handler{service: service}
}

func (h *Handler) Block(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cleanerID, err := parseIDParam(r, "cleanerID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cleaner_id")
		return
	}

	var req BlockedCleanerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	block, err := h.service.Block(r.Context(), currentUser.UserID, cleanerID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAlreadyBlocked) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, block)
}

func (h *Handler) Unblock(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cleanerID, err := parseIDParam(r, "cleanerID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cleaner id")
		return
	}

	err = h.service.Unblock(r.Context(), currentUser.UserID, cleanerID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrBlockedNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "cleaner unblocked",
	})
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	blocks, err := h.service.ListMine(r.Context(), currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, blocks)
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
}

