package wokrproof

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

func NewHandler(
	service *Service,
) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	bookingID, err := parseWorkProofUintParam(
		r,
		"bookingID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid booking id",
		)
		return
	}

	var req CreateWorkProofRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(
		&req,
	); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	proof, err :=
		h.service.Create(
			r.Context(),
			bookingID,
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
		ErrInvalidProofType,
	):
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return

	case errors.Is(
		err,
		ErrBookingNotFound,
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
		ErrBookingNotEligible,
	):
		response.Error(
			w,
			http.StatusConflict,
			err.Error(),
		)
		return

	case errors.Is(
		err,
		ErrProofLimitReached,
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
		http.StatusCreated,
		proof,
	)
}

func (h *Handler) GetBookingProof(w http.ResponseWriter, r *http.Request) {
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

	bookingID, err := parseWorkProofUintParam(
		r,
		"bookingID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid booking id",
		)
		return
	}

	result, err :=
		h.service.GetBookingProof(
			r.Context(),
			bookingID,
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
		ErrBookingNotFound,
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

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	proofID, err := parseWorkProofUintParam(
		r,
		"proofID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid proof id",
		)
		return
	}

	err = h.service.Delete(
		r.Context(),
		proofID,
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
		ErrProofNotFound,
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
		map[string]string{
			"message": "work proof deleted",
		},
	)
}

func (h *Handler) GetVerifiedWork(w http.ResponseWriter, r *http.Request) {
	cleanerID, err := parseWorkProofUintParam(
		r,
		"cleanerID",
	)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid cleaner id",
		)
		return
	}

	result, err :=
		h.service.GetVerifiedWork(
			r.Context(),
			cleanerID,
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

func parseWorkProofUintParam(r *http.Request, name string) (uint, error) {
	raw := chi.URLParam(
		r,
		name,
	)

	value, err :=
		strconv.ParseUint(
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
