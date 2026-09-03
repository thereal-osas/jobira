package cleaningteam

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newCleaningTeamHandlerForTest(
	repo Repository,
) *Handler {
	service := NewService(repo)

	return NewHandler(service)
}

func requestWithCleaningTeamUser(
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

func TestHandler_GetMyCleaningTeam_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		listTeamMembersFn: func(
			_ context.Context,
			clientID uint,
		) ([]TeamMemberData, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []TeamMemberData{
				{
					CleanerID:   20,
					CleanerName: "Maria",

					ServicesOffered: "housekeeping, laundry",

					LastJobType: "housekeeping",

					IsPreferred: true,

					CompletedJobsTogether: 4,

					ReliabilityScore: 92,
				},
			}, nil
		},
	}

	handler := newCleaningTeamHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaning-team/me",
		nil,
	)

	req = requestWithCleaningTeamUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyCleaningTeam(
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

	var result CleaningTeam

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if result.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			result.ClientID,
		)
	}

	if result.TotalMembers != 1 {
		t.Fatalf(
			"expected one member, got %d",
			result.TotalMembers,
		)
	}

	if result.Members[0].ServiceRole !=
		"Housekeeper" {
		t.Fatalf(
			"expected Housekeeper role, got %q",
			result.Members[0].ServiceRole,
		)
	}
}

func TestHandler_GetMyCleaningTeam_Unauthorized(
	t *testing.T,
) {
	handler := newCleaningTeamHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaning-team/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetMyCleaningTeam(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetMyCleaningTeam_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"repository failed",
	)

	repo := &mockRepository{
		listTeamMembersFn: func(
			context.Context,
			uint,
		) ([]TeamMemberData, error) {
			return nil, expectedErr
		},
	}

	handler := newCleaningTeamHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaning-team/me",
		nil,
	)

	req = requestWithCleaningTeamUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyCleaningTeam(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
