package health

import (
	"net/http"

	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

type Handler struct {}

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, HealthResponse{
		Status: "ok",
	})
}