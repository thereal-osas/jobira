package topoffer

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

func (h *Handler) ListJobOffers(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(
		r.Context(),
	)

	if err != nil {
		response.Error(w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	jobID, err := parseUintParam(
		r, "jobID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid job id",
		)
		return
	}

	req := ListTopOffersRequest{}

	if raw := r.URL.Query().Get(
		"limit",
	); raw != "" {
		value, err :=
			strconv.Atoi(
				raw,
			)

		if err != nil {
			response.Error(
				w,
				http.StatusBadRequest,
				"invalid limit",
			)
			return
		}

		req.Limit =
			value
	}

	if raw := r.URL.Query().Get(
		"offset",
	); raw != "" {
		value, err :=
			strconv.Atoi(
				raw,
			)

		if err != nil {
			response.Error(
				w,
				http.StatusBadRequest,
				"invalid offset",
			)
			return
		}

		req.Offset =
			value
	}

	result, err := h.service.RankJobOffers(
		r.Context(),
		jobID,
		currentUser.UserID,
		req,
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

	case errors.Is(
		err,
		ErrForbidden,
	):

		response.Error(
			w,
			http.StatusForbidden,
			err.Error(),
		)
		return

	case errors.Is(
		err,
		ErrJobClosed,
	):

		response.Error(
			w,
			http.StatusConflict,
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

func (h *Handler) GetMyApplicationRanking(w http.ResponseWriter, r *http.Request) {
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

	if currentUser.Role != "cleaner" {
		response.Error(
			w,
			http.StatusForbidden,
			"cleaner access required",
		)
		return
	}

	applicationID, err := parseUintParam(
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

	result, err := h.service.GetApplicationRanking(
		r.Context(),
		applicationID,
		currentUser.UserID,
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
		ErrApplicationNotFound,
	):

		response.Error(
			w,
			http.StatusNotFound,
			err.Error(),
		)
		return

	case errors.Is(
		err,
		ErrForbidden,
	):

		response.Error(
			w,
			http.StatusForbidden,
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

func parseUintParam(r *http.Request, name string) (uint, error) {
	raw := chi.URLParam(
		r,
		name,
	)

	value, err := strconv.ParseUint(
		raw,
		10,
		64,
	)
	if err != nil || value == 0 {
		return 0,
			ErrInvalidInput
	}

	return uint(value),
		nil
}
