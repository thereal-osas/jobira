package workhistory

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
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(
		r.Context(),
	)
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	history, err := h.service.ListMine(
		r.Context(),
		currentUser.UserID,
		currentUser.Role,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(
			w,
			http.StatusForbidden,
			err.Error(),
		)
		return
	}

	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		history,
	)
}

func (h *Handler) GetBookingHistory(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(
		r.Context(),
	)
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	bookingID, err := parseBookingID(r)
	if err != nil || bookingID == 0 {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid booking id",
		)
		return
	}

	history, err := h.service.GetBookingHistory(
		r.Context(),
		bookingID,
		currentUser.UserID,
		currentUser.Role,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(
			w,
			http.StatusForbidden,
			err.Error(),
		)
		return
	}

	if errors.Is(err, ErrHistoryNotFound) {
		response.Error(
			w,
			http.StatusNotFound,
			err.Error(),
		)
		return
	}

	if err != nil {
		response.Error(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		history,
	)
}

func parseBookingID(r *http.Request) (uint, error) {
	rawID := chi.URLParam(
		r,
		"bookingID",
	)

	parsedID, err := strconv.ParseUint(
		rawID,
		10,
		64,
	)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
}
