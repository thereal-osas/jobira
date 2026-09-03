package availablenow

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func availableNowRequestWithUser(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
			Role:   role,
		},
	)

	return req.WithContext(ctx)
}

func TestHandler_SetMine_Success(t *testing.T) {
	repo := &mockRepository{
		upsertFn: func(
			ctx context.Context,
			availability *CleanerAvailableNow,
		) error {
			availability.ID = 20
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := `{
		"duration_minutes": 180,
		"location": "Stratford",
		"travel_radius_miles": 8,
		"job_types": ["domestic", "airbnb"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(body),
	)

	req = availableNowRequestWithUser(
		req,
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.SetMine(
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

func TestHandler_SetMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.SetMine(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_SetMine_ForbiddenForClient(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(`{}`),
	)

	req = availableNowRequestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.SetMine(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			recorder.Code,
		)
	}
}

func TestHandler_SetMine_InvalidBody(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(`{invalid`),
	)

	req = availableNowRequestWithUser(
		req,
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.SetMine(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_SetMine_InvalidInput(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	body := `{
		"duration_minutes": 10,
		"location": "Stratford",
		"travel_radius_miles": 8,
		"job_types": ["domestic"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(body),
	)

	req = availableNowRequestWithUser(
		req,
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.SetMine(
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

func TestHandler_SetMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New(
		"upsert failed",
	)

	repo := &mockRepository{
		upsertFn: func(
			context.Context,
			*CleanerAvailableNow,
		) error {
			return expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := `{
		"duration_minutes": 120,
		"location": "Stratford",
		"travel_radius_miles": 8,
		"job_types": ["domestic"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/available-now/me",
		strings.NewReader(body),
	)

	req = availableNowRequestWithUser(
		req,
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.SetMine(
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

func TestHandler_GetMine_Success(t *testing.T) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailableNow, error) {
			return &CleanerAvailableNow{
				CleanerID:      10,
				IsAvailable:    true,
				AvailableFrom:  now.Add(-time.Hour),
				AvailableUntil: now.Add(time.Hour),
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/me",
		nil,
	)

	req = availableNowRequestWithUser(
		req,
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.GetMine(
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

func TestHandler_GetMine_ForbiddenForClient(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := availableNowRequestWithUser(
		httptest.NewRequest(
			http.MethodGet,
			"/available-now/me",
			nil,
		),
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetMine(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			recorder.Code,
		)
	}
}

func TestHandler_DisableMine_Success(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := availableNowRequestWithUser(
		httptest.NewRequest(
			http.MethodDelete,
			"/available-now/me",
			nil,
		),
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.DisableMine(
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

func TestHandler_DisableMine_NotFound(t *testing.T) {
	repo := &mockRepository{
		disableFn: func(
			context.Context,
			uint,
		) error {
			return ErrAvailableNowNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := availableNowRequestWithUser(
		httptest.NewRequest(
			http.MethodDelete,
			"/available-now/me",
			nil,
		),
		10,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.DisableMine(
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

func TestHandler_Search_Success(t *testing.T) {
	repo := &mockRepository{
		listAvailableCleanersFn: func(
			context.Context,
			AvailableNowSearchRequest,
			time.Time,
		) ([]AvailableCleaner, error) {
			return []AvailableCleaner{
				{
					CleanerID: 10,
					FullName:  "Sarah Cleaner",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/cleaners?location=Stratford&job_type=domestic&minimum_rating=4.5&limit=10",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Search(
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

func TestHandler_Search_InvalidRating(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/cleaners?minimum_rating=nope",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Search(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Search_InternalServerError(t *testing.T) {
	expectedErr := errors.New(
		"search failed",
	)

	repo := &mockRepository{
		listAvailableCleanersFn: func(
			context.Context,
			AvailableNowSearchRequest,
			time.Time,
		) ([]AvailableCleaner, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/available-now/cleaners",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Search(
		recorder,
		req,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}
