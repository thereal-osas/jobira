package applications

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
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

func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID, err := parseIDParam(r, "jobID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	var req CreateApplicationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdApplication, err := h.service.Apply(r.Context(), jobID, currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, subscriptionaccess.ErrDailyApplicationLimit) {
		response.Error(
			w,
			http.StatusForbidden,
			"You have reached today's free application limit. Upgrade to Premium for unlimited applications.",
		)
		return
	}

	if errors.Is(err, ErrAlreadyApplied) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if errors.Is(err, ErrCannotApplyToOwnJob) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, createdApplication)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	applicationID, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid application id")
		return
	}

	foundApplication, err := h.service.GetByID(r.Context(), applicationID, currentUser.UserID, currentUser.Role)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrApplicationNotFound) {
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

	response.JSON(w, http.StatusOK, foundApplication)
}

 func (h *Handler) ListForJob(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID, err := parseIDParam(r, "jobID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	applications, err := h.service.ListByJobID(r.Context(), jobID, currentUser.UserID, currentUser.Role)
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

	response.JSON(w, http.StatusOK, applications)
 }

 func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	applications, err := h.service.ListMine(r.Context(), currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, applications)
 }

 func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	applicationID, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid application id ")
		return
	}

	var req UpdateApplicationStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updateApplication, err := h.service.UpdateStatus(r.Context(), applicationID, currentUser.UserID, currentUser.Role, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidApplicationStatus) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrApplicationNotFound) {
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

	response.JSON(w, http.StatusOK, updateApplication)
 }

 func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
 }