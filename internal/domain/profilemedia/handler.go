package profilemedia

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
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	media, err := h.service.CreateForUser(
		r.Context(),
		currentUser.UserID,
		req,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusCreated, media)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	media, err := h.service.ListForUser(
		r.Context(),
		currentUser.UserID,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, media)
}

func (h *Handler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseIDParam(r, "userID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	media, err := h.service.ListForUser(
		r.Context(),
		userID,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, media)
}

func (h *Handler) UpdateMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseIDParam(r, "mediaID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid media id")
		return
	}

	var req UpdateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	media, err := h.service.UpdateForUser(
		r.Context(),
		currentUser.UserID,
		mediaID,
		req,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, media)
}

func (h *Handler) DeleteMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseIDParam(r, "mediaID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid media id")
		return
	}

	err = h.service.DeleteForUser(
		r.Context(),
		currentUser.UserID,
		mediaID,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "profile media deleted",
	})
}

func (h *Handler) CreateForCompany(w http.ResponseWriter, r *http.Request) {
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

	var req CreateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	media, err := h.service.CreateForCompany(
		r.Context(),
		currentUser.UserID,
		companyID,
		req,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusCreated, media)
}

func (h *Handler) ListByCompany(w http.ResponseWriter, r *http.Request) {
	companyID, err := parseIDParam(r, "companyID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid company id")
		return
	}

	media, err := h.service.ListForCompany(
		r.Context(),
		companyID,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, media)
}

func (h *Handler) UpdateForCompany(w http.ResponseWriter, r *http.Request) {
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

	mediaID, err := parseIDParam(r, "mediaID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid media id")
		return
	}

	var req UpdateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	media, err := h.service.UpdateForCompany(
		r.Context(),
		currentUser.UserID,
		companyID,
		mediaID,
		req,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, media)
}

func (h *Handler) DeleteForCompany(w http.ResponseWriter, r *http.Request) {
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

	mediaID, err := parseIDParam(r, "mediaID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid media id")
		return
	}

	err = h.service.DeleteForCompany(
		r.Context(),
		currentUser.UserID,
		companyID,
		mediaID,
	)
	if handleServiceError(w, err) {
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "company profile media deleted",
	})
}

func handleServiceError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrInvalidMediaType):
		response.Error(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, Errforbidden):
		response.Error(w, http.StatusForbidden, err.Error())

	case errors.Is(err, ErrMediaNotFound),
		errors.Is(err, ErrCompanyNotFound):
		response.Error(w, http.StatusNotFound, err.Error())

	case errors.Is(err, ErrProfilePhotoLimitReached),
		errors.Is(err, ErrPortfolioLimitReached),
		errors.Is(err, ErrPricingLimitReached),
		errors.Is(err, ErrCompanyLogoLimitReached):
		response.Error(w, http.StatusConflict, err.Error())

	default:
		response.Error(w, http.StatusInternalServerError, err.Error())
	}

	return true
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || parsedID == 0 {
		return 0, ErrInvalidInput
	}

	return uint(parsedID), nil
}
