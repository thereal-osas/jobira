package jobs

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	subscriptionaccess "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdJob, err := h.service.Create(r.Context(), currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, subscriptionaccess.ErrDailyJobPostLimit) {
		response.Error(
			w,
			http.StatusForbidden, "You have reached today's free job post limit, Upgrade or pay per post to continue posting",
		)
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, createdJob)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	foundJob, err := h.service.GetByID(r.Context(), id)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, foundJob)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.service.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, jobs)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobs, err := h.service.ListByClientID(r.Context(), currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, jobs)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid")
		return
	}

	var req UpdateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedJob, err := h.service.Update(r.Context(), id, currentUser.UserID, currentUser.Role, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrJobNotFound) {
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

	response.JSON(w, http.StatusOK, updatedJob)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseIDParam(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid job id")
		return
	}

	err = h.service.Delete(r.Context(), id, currentUser.UserID, currentUser.Role)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrJobNotFound) {
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
		"message": "job deleted",
	})
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	ParsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(ParsedID), nil
}
