package availability

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, err:= identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CleanerAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	availability, err := h.service.Create(r.Context(), currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidStatus) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityExists) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	} 

	response.JSON(w, http.StatusCreated, availability)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	availabilityID, err := parseIDParam(r, "availabilityID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability id")
		return
	}

	availability, err := h.service.GetByID(r.Context(), availabilityID, currentUser.UserID, currentUser.Role)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusForbidden, err.Error())
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

	response.JSON(w, http.StatusOK, availability)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, err := h.service.ListMine(r.Context(), currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, records)
}

func (h *Handler) ListByCleaner(w http.ResponseWriter, r *http.Request) {
	cleanerID, err := parseIDParam(r, "cleanerID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cleaner_id")
		return
	}

	records, err := h.service.ListByCleanerID(r.Context(), cleanerID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, records)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	availabilityID, err := parseIDParam(r, "availabilityID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability id")
		return
	}

	var req UpdateAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	availability, err := h.service.Update(r.Context(), availabilityID, currentUser.UserID, currentUser.Role, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidStatus) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, availability)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	availabilityID, err := parseIDParam(r, "availabilityID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability id")
		return
	}

	err = h.service.Delete(r.Context(), availabilityID, currentUser.UserID, currentUser.Role)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
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

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "availability deleted",
	})
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
}

