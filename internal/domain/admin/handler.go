package admin

import (
	"net/http"

	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

type Handler struct {
	service *Service
}

type DashboardMessage struct {
	Message string `json:"message"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, DashboardMessage{
		Message: h.service.DashboardMessage(),
	})
}