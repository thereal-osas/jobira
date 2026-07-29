package auth

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.handleRegisterError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid requestbody")
		return
	}

	result, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.handleLoginError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func (h *Handler) handleRegisterError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrEmailAlreadyInUse) {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	response.Error(w, http.StatusInternalServerError, err.Error())
}

func (h *Handler) handleLoginError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidInput) {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, ErrInvalidCredentials) {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.Error(w, http.StatusInternalServerError, err.Error())
} 