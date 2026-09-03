package availablenow

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) SetMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized, "unauthorized",
		)
		return
	}

	if currentUser.Role != "cleaner" {
		response.Error(
			w,
			http.StatusForbidden, "cleaner access required",
		)
		return
	}

	var req SetAvailableNowRequest

	if err := json.NewDecoder(r.Body).Decode(
		&req,
	); err != nil {
		response.Error(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	availability, err :=
		h.service.SetAvailableNow(
			r.Context(),
			currentUser.UserID,
			req,
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
		availability,
	)
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

	if currentUser.Role != "cleaner" {
		response.Error(
			w,
			http.StatusForbidden,
			"cleaner access required",
		)
		return
	}

	status, err := h.service.GetStatus(r.Context(), currentUser.UserID)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(
			w,
			http.StatusBadRequest,
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
		status,
	)
}

func (h *Handler) DisableMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(
			w,
			http.StatusUnauthorized, "unauthorized",
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

	err =
		h.service.Disable(r.Context(), currentUser.UserID)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(
			w,
			http.StatusBadRequest,
			err.Error(),
		)

		return
	}

	if errors.Is(
		err,
		ErrAvailableNowNotFound,
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
		map[string]string{
			"message": "available now disabled",
		},
	)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	search :=
		AvailableNowSearchRequest{
			Location: r.URL.Query().Get(
				"location",
			),

			JobType: r.URL.Query().Get(
				"job_type",
			),

			VerifiedOnly: r.URL.Query().Get(
				"verified_only",
			) == "true",

			DBSRequired: r.URL.Query().Get(
				"dbs_required",
			) == "true",
		}

	if raw :=
		r.URL.Query().Get(
			"minimum_rating",
		); raw != "" {
		value, err :=
			strconv.ParseFloat(
				raw,
				64,
			)
		if err != nil {
			response.Error(
				w,
				http.StatusBadRequest,
				"invalid minimum rating",
			)
			return
		}

		search.MinimumRating = value
	}

	if raw :=
		r.URL.Query().Get(
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

		search.Limit = value
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

		search.Offset = value
	}

	cleaners, err := h.service.Search(r.Context(), search)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
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
		cleaners,
	)
}
