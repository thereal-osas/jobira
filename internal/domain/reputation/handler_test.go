package reputation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func newReputationTestRouter(repository Repository) http.Handler {
	service := NewService(repository)
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

	return router
}

func TestHandler_GetByCleanerID_Success(t *testing.T) {
	now := time.Now().UTC()

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			if cleanerID != 10 {
				t.Fatalf("expected cleaner ID 10, got %d", cleanerID)
			}

			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return &CleanerReputation{
				CleanerID:                cleanerID,
				AverageRating:            5.0,
				TotalReviews:             25,
				CompletedJobs:            60,
				RepeatClients:            12,
				WouldHireAgainCount:      23,
				RecommendationPercentage: 100,

				TotalBookings:        60,
				CleanerCancellations: 0,

				EligibleResponseMessages: 30,
				RespondedMessages:        30,
				AverageResponseMinutes:   8,

				Badge:     "Old Badge",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/10",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var result CleanerReputation

	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.CleanerID != 10 {
		t.Fatalf("expected cleaner ID 10, got %d", result.CleanerID)
	}

	if result.AverageRating != 5 {
		t.Fatalf(
			"expected average rating 4.9, got %.2f",
			result.AverageRating,
		)
	}

	if result.TotalReviews != 25 {
		t.Fatalf(
			"expected 25 reviews, got %d",
			result.TotalReviews,
		)
	}

	if result.Badge != "Elite Cleaner" {
		t.Fatalf(
			"expected Elite Cleaner badge, got %q",
			result.Badge,
		)
	}
}

func TestHandler_GetByCleanerID_InvalidTextID(t *testing.T) {
	refreshCalled := false

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			refreshCalled = true
			return nil
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/not-a-number",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if refreshCalled {
		t.Fatal("expected repository Refresh not to be called")
	}

	if !strings.Contains(
		recorder.Body.String(),
		"invalid cleaner id",
	) {
		t.Fatalf(
			"expected invalid cleaner id response, got %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleanerID_ZeroID(t *testing.T) {
	refreshCalled := false

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			refreshCalled = true
			return nil
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/0",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if refreshCalled {
		t.Fatal("expected repository Refresh not to be called")
	}

	if !strings.Contains(
		recorder.Body.String(),
		"invalid cleaner id",
	) {
		t.Fatalf(
			"expected invalid cleaner id response, got %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleanerID_NegativeID(t *testing.T) {
	repository := &mockRepository{}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/-4",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"invalid cleaner id",
	) {
		t.Fatalf(
			"expected invalid cleaner id response, got %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleanerID_ReputationNotFound(t *testing.T) {
	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return nil, ErrReputationNotFound
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/99",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		ErrReputationNotFound.Error(),
	) {
		t.Fatalf(
			"expected response to contain %q, got %s",
			ErrReputationNotFound.Error(),
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleanerID_RefreshError(t *testing.T) {
	expectedErr := errors.New("failed to refresh cleaner reputation")

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return expectedErr
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/10",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		expectedErr.Error(),
	) {
		t.Fatalf(
			"expected response to contain %q, got %s",
			expectedErr.Error(),
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleanerID_RepositoryGetError(t *testing.T) {
	expectedErr := errors.New("failed to retrieve cleaner reputation")

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return nil, expectedErr
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/10",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		expectedErr.Error(),
	) {
		t.Fatalf(
			"expected response to contain %q, got %s",
			expectedErr.Error(),
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetReputationRouteExists(t *testing.T) {
	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return &CleanerReputation{
				CleanerID:     cleanerID,
				CompletedJobs: 1,
			}, nil
		},
	}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/7",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code == http.StatusNotFound {
		t.Fatalf(
			"expected reputation route to exist, got status %d",
			recorder.Code,
		)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_UnsupportedMethod(t *testing.T) {
	repository := &mockRepository{}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodPost,
		"/reputation/10",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusMethodNotAllowed,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_MissingCleanerID(t *testing.T) {
	repository := &mockRepository{}

	router := newReputationTestRouter(repository)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
