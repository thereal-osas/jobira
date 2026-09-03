package cleanerprogression

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func requestWithCleaner(
	req *http.Request,
	cleanerID uint,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: cleanerID,
			Role:   "cleaner",
		},
	)

	return req.WithContext(ctx)
}

func TestHandler_GetMine_Success(t *testing.T) {
	repo := &mockRepository{
		getCleanerReputationFn: func(
			context.Context,
			uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID:                10,
				AverageRating:            4.9,
				TotalReviews:             15,
				CompletedJobs:            35,
				RepeatClients:            5,
				RecommendationPercentage: 94,
				ReliabilityScore:         92,
				TotalBookings:            40,
				CleanerCancellations:     2,
				EligibleResponseMessages: 30,
				RespondedMessages:        27,
				Badge:                    "Top Rated",
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-progression/me",
		nil,
	)

	req = requestWithCleaner(
		req,
		10,
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

	var result ProgressSnapshot

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&result); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if result.CleanerID != 10 {
		t.Fatalf(
			"expected cleaner ID 10, got %d",
			result.CleanerID,
		)
	}

	if result.CurrentBadge != "Top Rated" {
		t.Fatalf(
			"expected Top Rated, got %q",
			result.CurrentBadge,
		)
	}
}

func TestHandler_GetMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-progression/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetMine(
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

func TestHandler_GetMine_NotFound(t *testing.T) {
	repo := &mockRepository{
		getCleanerReputationFn: func(
			context.Context,
			uint,
		) (*reputationdomain.CleanerReputation, error) {
			return nil,
				reputationdomain.ErrReputationNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := requestWithCleaner(
		httptest.NewRequest(
			http.MethodGet,
			"/cleaner-progression/me",
			nil,
		),
		10,
	)

	recorder := httptest.NewRecorder()

	handler.GetMine(
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

func TestHandler_GetMine_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"progression failed",
	)

	repo := &mockRepository{
		getCleanerReputationFn: func(
			context.Context,
			uint,
		) (*reputationdomain.CleanerReputation, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := requestWithCleaner(
		httptest.NewRequest(
			http.MethodGet,
			"/cleaner-progression/me",
			nil,
		),
		10,
	)

	recorder := httptest.NewRecorder()

	handler.GetMine(
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
