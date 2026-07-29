package billing

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateCheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invaid request body")
		return
	}

	session, err := h.service.CreateCheckoutSession(r.Context(), currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrPlanNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrStripePriceMissing) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrCheckoutSessionError) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, session)
}

func (h *Handler) CreateBillingPortalSession(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	portalSession, err := h.service.CreateBillingPortalSession(r.Context(), currentUser.UserID)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrCustomerNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrPortalSessionError) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, portalSession)
}

func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid webhook body")
		return
	}

	signatureHeader := r.Header.Get("Stripe-Signature")

	if err := h.service.HandleWebhook(r.Context(), payload, signatureHeader); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "webhook received",
	})
}


func (h *Handler) ValidatePromoCode(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ValidatePromoCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	promo, err := h.service.ValidatePromoCode(r.Context(), currentUser.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrPromoCodeNotFound) {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, ErrPromoCodeInvalid) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, promo)
}