package adminmoderation

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

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.ListReports(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, reports)
}

func (h *Handler) ListOpenReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.ListOpenReports(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, reports)
}



func (h *Handler) ListBlockCleaners(w http.ResponseWriter, r *http.Request) {
	blocks, err := h.service.ListBlockCleaners(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, blocks)
}

func (h *Handler) UpdatedReportStatus(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reportID, err := parseIDParam(r, "reportID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid report id")
		return
	}

	var req UpdateReportStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.UpdateReportStatus(
		r.Context(), 
		reportID, 
		currentUser.UserID, 
		req,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidStatus) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrReportNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "report status updated", 
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