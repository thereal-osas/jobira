package jobs

import (
	"net/http"
	"strconv"

	"github.com/rodrigueghenda/jobira/internal/transport/http/response"
)

type SearchJobRequest struct {
	Location    string
	JobType     string
	ListingType string
	MinBudget   int
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	minBudget := 0

	if rawBudget := r.URL.Query().Get("min_budget"); rawBudget != "" {
		parseBudget, err := strconv.Atoi(rawBudget)
		if err == nil {
			minBudget = parseBudget
		}
	}

	req := SearchJobRequest{
		Location:    r.URL.Query().Get("location"),
		JobType:     r.URL.Query().Get("job_type"),
		ListingType: r.URL.Query().Get("listing_type"),
		MinBudget:   minBudget,
	}

	jobs, err := h.service.Search(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, jobs)
}
