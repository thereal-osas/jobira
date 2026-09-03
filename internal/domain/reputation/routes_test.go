package reputation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_GetByCleanerID(t *testing.T) {
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

	service := NewService(repository)
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/12",
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
}

func TestRegisterRoutes_InvalidCleanerID(t *testing.T) {
	repository := &mockRepository{}

	service := NewService(repository)
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/invalid",
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
}

func TestRegisterRoutes_ZeroCleanerID(t *testing.T) {
	repository := &mockRepository{}

	service := NewService(repository)
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

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
}

func TestRegisterRoutes_MethodNotAllowed(t *testing.T) {
	repository := &mockRepository{}

	service := NewService(repository)
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/reputation/12",
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

func TestRegisterRoutes_UnknownReputationRoute(t *testing.T) {
	repository := &mockRepository{}

	service := NewService(repository)
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/reputation/12/history",
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
