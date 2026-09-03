package jobpulse

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(
	service *Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetJobPulse(
	w http.ResponseWriter,
	r *http.Request,
) {
	jobID, err := parseJobPulseID(
		r,
		"jobID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid job id",
		)
		return
	}

	result, err := h.service.GetJobPulse(
		r.Context(),
		jobID,
	)

	switch {
	case errors.Is(
		err,
		ErrInvalidInput,
	):
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return

	case errors.Is(
		err,
		ErrJobNotFound,
	):
		response.Error(
			w,
			http.StatusNotFound,
			err.Error(),
		)
		return

	case err != nil:
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
		result,
	)
}

func parseJobPulseID(
	r *http.Request,
	name string,
) (uint, error) {
	raw := chi.URLParam(
		r,
		name,
	)

	value, err := strconv.ParseUint(
		raw,
		10,
		64,
	)
	if err != nil ||
		value == 0 {
		return 0,
			ErrInvalidInput
	}

	return uint(value),
		nil
}
