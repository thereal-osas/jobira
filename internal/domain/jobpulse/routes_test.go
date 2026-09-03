package jobpulse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_GetJobPulse(t *testing.T) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID:                  jobID,
				JobStatus:              "open",
				CreatedAt:              time.Now().UTC().Add(-8 * time.Hour),
				ApplicationCount:       4,
				RecentApplicationCount: 2,
				BookingCount:           0,
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/jobs/10",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_InvalidJobID(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/jobs/nope",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_NotFound(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/not-a-route",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}
