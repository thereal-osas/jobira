package jobpulse

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func withJobPulseURLParam(
	req *http.Request,
	key string,
	value string,
) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(
		key,
		value,
	)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	)

	return req.WithContext(ctx)
}

func TestHandler_GetJobPulse_Success(t *testing.T) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID:                  jobID,
				JobStatus:              "open",
				CreatedAt:              time.Now().UTC().Add(-12 * time.Hour),
				ApplicationCount:       5,
				RecentApplicationCount: 3,
				BookingCount:           0,
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/jobs/10",
		nil,
	)

	req = withJobPulseURLParam(
		req,
		"jobID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobPulse(
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

func TestHandler_GetJobPulse_InvalidJobID(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/jobs/nope",
		nil,
	)

	req = withJobPulseURLParam(
		req,
		"jobID",
		"nope",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobPulse(
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

func TestHandler_GetJobPulse_NotFound(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/jobs/999",
		nil,
	)

	req = withJobPulseURLParam(
		req,
		"jobID",
		"999",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobPulse(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetJobPulse_InternalServerError(t *testing.T) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-pulse/jobs/10",
		nil,
	)

	req = withJobPulseURLParam(
		req,
		"jobID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.GetJobPulse(
		recorder,
		req,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
