package availability

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CleanerAvailabilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	availability, err := h.service.Create(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidTimeRange) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityExists) ||
		errors.Is(err, ErrAvailabilityConflict) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, availability)
}

func (h *Handler) CreateBulk(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req BulkAvailabilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	records, err := h.service.CreateBulk(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidTimeRange) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityExists) ||
		errors.Is(err, ErrAvailabilityConflict) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, records)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	availabilityID, err := parseIDParam(r, "availabilityID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability id")
		return
	}

	availability, err := h.service.GetByID(
		r.Context(),
		availabilityID,
		currentUser.UserID,
		currentUser.Role,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, availability)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, err := h.service.ListMine(
		r.Context(),
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, records)
}

func (h *Handler) ListByCleaner(w http.ResponseWriter, r *http.Request) {
	cleanerID, err := parseIDParam(r, "cleanerID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cleaner id")
		return
	}

	records, err := h.service.ListByCleanerID(
		r.Context(),
		cleanerID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, records)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	availabilityID, err := parseIDParam(r, "availabilityID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability id")
		return
	}

	var req UpdateAvailabilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	availability, err := h.service.Update(
		r.Context(),
		availabilityID,
		currentUser.UserID,
		currentUser.Role,
		req,
	)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidTimeRange) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityConflict) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, availability)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	availabilityID, err := parseIDParam(r, "availabilityID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability id")
		return
	}

	err = h.service.Delete(
		r.Context(),
		availabilityID,
		currentUser.UserID,
		currentUser.Role,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "availability deleted",
	})
}

func (h *Handler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAvailabilityBlockRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	block, err := h.service.CreateBlock(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, block)
}

func (h *Handler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	blocks, err := h.service.ListBlocks(
		r.Context(),
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, blocks)
}

func (h *Handler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	blockID, err := parseIDParam(r, "blockID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability block id")
		return
	}

	err = h.service.DeleteBlock(
		r.Context(),
		blockID,
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrForbidden) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "availability block deleted",
	})
}

func (h *Handler) CreateRecurring(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRecurringAvailabilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	recurring, err := h.service.CreateRecurring(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidTimeRange) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, recurring)
}

func (h *Handler) CreateRecurringBulk(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req BulkRecurringAvailabilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	records, err := h.service.CreateRecurringBulk(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidTimeRange) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, records)
}

func (h *Handler) ListRecurring(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	records, err := h.service.ListRecurring(
		r.Context(),
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, records)
}

func (h *Handler) DeleteRecurring(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	recurringID, err := parseIDParam(r, "recurringID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid recurring availability id")
		return
	}

	err = h.service.DeleteRecurring(
		r.Context(),
		recurringID,
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrRecurringAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "recurring availability deleted",
	})
}

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	settings, err := h.service.GetSettings(
		r.Context(),
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateAvailabilitySettingsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	settings, err := h.service.UpdateSettings(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, settings)
}

func (h *Handler) CreateOverride(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAvailabilityOverrideRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	override, err := h.service.CreateOverride(
		r.Context(),
		currentUser.UserID,
		req,
	)

	if errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidTimeRange) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, override)
}

func (h *Handler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	overrideID, err := parseIDParam(r, "overrideID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid availability override id")
		return
	}

	err = h.service.DeleteOverride(
		r.Context(),
		overrideID,
		currentUser.UserID,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityOverrideNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "availability override deleted",
	})
}

func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	cleanerID, err := parseIDParam(r, "cleanerID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cleaner id")
		return
	}

	fromDate := strings.TrimSpace(
		r.URL.Query().Get("from"),
	)

	toDate := strings.TrimSpace(
		r.URL.Query().Get("to"),
	)

	if fromDate == "" || toDate == "" {
		response.Error(
			w,
			http.StatusBadRequest,
			"from and to dates are required",
		)
		return
	}

	slots, err := h.service.GetCalendar(
		r.Context(),
		cleanerID,
		fromDate,
		toDate,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, slots)
}

func (h *Handler) NextAvailable(w http.ResponseWriter, r *http.Request) {
	cleanerID, err := parseIDParam(r, "cleanerID")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cleaner id")
		return
	}

	fromDate := strings.TrimSpace(
		r.URL.Query().Get("from"),
	)

	if fromDate == "" {
		fromDate = time.Now().
			Format("2006-01-02")
	}

	days := 30

	if rawDays := strings.TrimSpace(
		r.URL.Query().Get("days"),
	); rawDays != "" {
		parsedDays, err := strconv.Atoi(rawDays)
		if err != nil || parsedDays <= 0 {
			response.Error(
				w,
				http.StatusBadRequest,
				"invalid days",
			)
			return
		}

		days = parsedDays
	}

	slot, err := h.service.NextAvailable(
		r.Context(),
		cleanerID,
		fromDate,
		days,
	)

	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrAvailabilityNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, slot)
}

func parseIDParam(r *http.Request, name string) (uint, error) {
	rawID := chi.URLParam(r, name)

	parsedID, err := strconv.ParseUint(
		rawID,
		10,
		64,
	)
	if err != nil || parsedID == 0 {
		return 0, ErrInvalidInput
	}

	return uint(parsedID), nil
}
