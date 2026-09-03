package companyaccounts

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

func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	company, err := h.service.CreateCompany(r.Context(), currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrMemberExists) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, company)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	companyID, err := parseIDParam(r, "companyID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid company id")
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.AddMember(r.Context(), companyID, currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrCompanyNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if errors.Is(err, ErrSeatLimitReached) {
		response.Error(w, http.StatusPaymentRequired, err.Error())
		return
	}

	if errors.Is(err, ErrMemberExists) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{
		"message": "company member added",
	})
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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

	companyID, err := parseIDParam(r, "companyID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid company id")
		return
	}

	userID, err := parseIDParam(r, "userID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.service.RemoveMember(r.Context(), companyID, currentUser.UserID, userID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrCompanyNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "company member removed",
	})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	currentUser, err :=
		identity.FromContext(
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

	companyID, err := parseIDParam(r, "companyID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid company id")
		return
	}

	members, err := h.service.ListMembers(r.Context(), companyID, currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, members)
}

func (h *Handler) UpdateMemberStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentUser, err :=
		identity.FromContext(
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

	companyID, err :=
		parseIDParam(
			r,
			"companyID",
		)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid company id",
		)
		return
	}

	userID, err :=
		parseIDParam(
			r,
			"userID",
		)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)
		return
	}

	var req UpdateMemberStatusRequest

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

	err =
		h.service.UpdateMemberStatus(
			r.Context(),
			companyID,
			currentUser.UserID,
			userID,
			req,
		)

	writeCompanyMutationError(
		w,
		err,
	)
	if err != nil {
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "company member status updated",
		},
	)
}

func (h *Handler) UpdateMemberRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentUser, err :=
		identity.FromContext(
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

	companyID, err :=
		parseIDParam(
			r,
			"companyID",
		)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid company id",
		)
		return
	}

	userID, err :=
		parseIDParam(
			r,
			"userID",
		)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)
		return
	}

	var req UpdateMemberRolesRequest

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

	err =
		h.service.UpdateMemberRole(
			r.Context(),
			companyID,
			currentUser.UserID,
			userID,
			req,
		)

	writeCompanyMutationError(
		w,
		err,
	)
	if err != nil {
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]string{
			"message": "company member role updated",
		},
	)
}

func (h *Handler) GetDashboard(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentUser, err :=
		identity.FromContext(
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

	companyID, err :=
		parseIDParam(
			r,
			"companyID",
		)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid company id",
		)
		return
	}

	result, err :=
		h.service.GetDashboard(
			r.Context(),
			companyID,
			currentUser.UserID,
		)

	if errors.Is(
		err,
		ErrInvalidInput,
	) {
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	if errors.Is(
		err,
		ErrForbidden,
	) {
		response.Error(
			w,
			http.StatusForbidden,
			err.Error(),
		)
		return
	}

	if errors.Is(
		err,
		ErrCompanyNotFound,
	) {
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
		result,
	)
}

func writeCompanyMutationError(
	w http.ResponseWriter,
	err error,
) {
	if err == nil {
		return
	}

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

	case errors.Is(
		err,
		ErrCompanyNotFound,
	):
		response.Error(
			w,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(
		err,
		ErrMemberNotFound,
	):
		response.Error(
			w,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(
		err,
		ErrForbidden,
	):
		response.Error(
			w,
			http.StatusForbidden,
			err.Error(),
		)

	case errors.Is(
		err,
		ErrSeatLimitReached,
	):
		response.Error(
			w,
			http.StatusPaymentRequired,
			err.Error(),
		)

	case errors.Is(
		err,
		ErrSeatPlanNotFound,
	):
		response.Error(
			w,
			http.StatusPaymentRequired,
			err.Error(),
		)

	case errors.Is(
		err,
		ErrMemberExists,
	):
		response.Error(
			w,
			http.StatusConflict,
			err.Error(),
		)

	default:
		response.Error(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
	}
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedID), nil
}
