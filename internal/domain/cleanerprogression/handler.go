package cleanerprogression

import (
	"errors"
	"net/http"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
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

func (h *Handler) GetMine(w http.ResponseWriter, r *http.Request) {
	currentUser, err := identity.FromContext(
		r.Context(),
	)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	progression, err := h.service.GetMyProgression(
		r.Context(), currentUser.UserID,
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
		reputationdomain.ErrReputationNotFound,
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
			w, http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		progression,
	)
}
