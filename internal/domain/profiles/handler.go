package profiles

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	profiles, err := h.service.Create(r.Context(), currentUser.UserID, req)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidAvailabilityStatus) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrProfileAlreadyExists) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, profiles)
}

func (h *Handler) GetMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profiles, err := h.service.GetMine(r.Context(), currentUser.UserID)

	if errors.Is(err, ErrProfileNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, profiles)
}

func (h *Handler) GetMyProfileStrength(w http.ResponseWriter, r *http.Request) {
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

	result, err :=
		h.service.GetMyProfileStrength(
			r.Context(),
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
		ErrProfileNotFound,
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
func (h *Handler) GetNewOnJobiraStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	_, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	userID, err := parseUserIDParam(r)
	if err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)
		return
	}

	result, err := h.service.GetNewOnJobiraStatus(
		r.Context(),
		userID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			response.Error(
				w,
				http.StatusBadRequest,
				err.Error(),
			)

		case errors.Is(err, ErrProfileNotFound):
			response.Error(
				w,
				http.StatusNotFound,
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
		result,
	)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	profiles, err := h.service.Update(r.Context(), currentUser.UserID, req)

	if errors.Is(err, ErrInvalidAvailabilityStatus) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrProfileNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, profiles)
}

func (h *Handler) UpdateVerificationStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserIDParam(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req UpdateVerificationStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	profiles, err := h.service.UpdateVerificationStatus(
		r.Context(),
		userID,
		req.VerificationStatus,
	)

	if errors.Is(err, ErrInvalidVerificationState) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrProfileNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, profiles)
}

func parseUserIDParam(r *http.Request) (uint, error) {
	rawID := chi.URLParam(r, "userID")

	parsedID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || parsedID == 0 {
		return 0, ErrInvalidInput
	}

	return uint(parsedID), nil

}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {

	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	minExperience := 0
	maxHourlyRate := 0
	MaxTravelRadius := 0

	if raw := r.URL.Query().Get("min_experience"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			minExperience = parsed
		}
	}

	if raw := r.URL.Query().Get("max_hourly_rate"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			maxHourlyRate = parsed
		}
	}

	if raw := r.URL.Query().Get("max_travel_radius"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			MaxTravelRadius = parsed
		}
	}

	var IsVerified *bool

	if raw := r.URL.Query().Get("is_verified"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err == nil {
			IsVerified = &parsed
		}
	}

	req := SearchProfilesRequest{
		ClientID:           currentUser.UserID,
		Country:            r.URL.Query().Get("country"),
		City:               r.URL.Query().Get("city"),
		Region:             r.URL.Query().Get("region"),
		PostcodeArea:       r.URL.Query().Get("postcode_area"),
		AvailabilityStatus: r.URL.Query().Get("availability_status"),
		ServicesOffered:    r.URL.Query().Get("services_offered"),
		MinExperience:      minExperience,
		MaxHourlyRate:      maxHourlyRate,
		MaxTravelRadius:    MaxTravelRadius,
		IsVerified:         IsVerified,
	}

	profiles, err := h.service.Search(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, profiles)
}

func (h *Handler) GetCompletedJobs(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	history, err := h.service.GetCompletedJobs(r.Context(), currentUser.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, history)
}

func (h *Handler) GetCancelledJobs(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	history, err := h.service.GetCancelledJobs(r.Context(), currentUser.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, history)
}

func (h *Handler) GetFullHistory(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	history, err := h.service.GetFullHistory(r.Context(), currentUser.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, history)
}
