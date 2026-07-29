package bookingtimeline

import (
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

func (h *Handler) GetByBookingID(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookingID, err := parseBookingID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	timeline, err := h.service.GetTimeline(
		r.Context(),
		bookingID,
		currentUser.UserID,
		currentUser.Role,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

		if errors.Is(err, ErrBookingNotFound) {
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

	response.JSON(w, http.StatusOK, timeline)
}

func parseBookingID(r *http.Request) (uint, error) {
	rawID := chi.URLParam(r, "bookingID")

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
}