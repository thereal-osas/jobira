package applicationtimeline

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

func (h *Handler) GetTimeline(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	applicationID, err := parseIDParam(
		r,
		"applicationID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid application id",
		)
		return
	}

	timeline, err := h.service.GetTimeline(
		r.Context(),
		applicationID,
		currentUser.UserID,
		currentUser.Role,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			response.Error(
				w,
				http.StatusBadRequest,
				err.Error(),
			)

		case errors.Is(err, ErrApplicationNotFound):
			response.Error(
				w,
				http.StatusNotFound,
				err.Error(),
			)

		case errors.Is(err, ErrTimelineNotFound):
			response.Error(
				w,
				http.StatusNotFound,
				err.Error(),
			)

		case errors.Is(err, ErrForbidden):
			response.Error(
				w,
				http.StatusForbidden,
				err.Error(),
			)

		default:
			response.Error(
				w,
				http.StatusInternalServerError,
				err.Error(),
			)
		}

		return
	}

	response.JSON(
		w,
		http.StatusOK,
		timeline,
	)
}

func parseIDParam(
	r *http.Request,
	name string,
) (uint, error) {
	rawID := chi.URLParam(r, name)

	parseID, err := strconv.ParseUint(
		rawID,
		10,
		64,
	)
	if err != nil || parseID == 0 {
		return 0, ErrInvalidInput
	}

	return uint(parseID), nil
}
