package jobinvitations

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newHandlerForTest(repo Repository) *Handler {
	service := NewService(
		repo,
		nil,
		nil,
	)

	return NewHandler(service)
}

func requestWithUser(
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

func requestWithURLParam(
	req *http.Request,
	name string,
	value string,
) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(name, value)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
	)

	return req.WithContext(ctx)
}

func TestNewHandler(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestParseIDParam_Success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		nil,
	)
	req = requestWithURLParam(req, "jobID", "7")

	id, err := parseIDParam(req, "jobID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 7 {
		t.Fatalf("expected ID 7, got %d", id)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/nope",
		nil,
	)
	req = requestWithURLParam(req, "jobID", "nope")

	id, err := parseIDParam(req, "jobID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf("expected ID 0, got %d", id)
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*JobInvitation,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{
			"cleaner_id":8,
			"message":"Please apply"
		}`),
	)
	req = requestWithURLParam(req, "jobID", "7")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidJobID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/nope",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(req, "jobID", "nope")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{"cleaner_id":`),
	)
	req = requestWithURLParam(req, "jobID", "7")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{
			"cleaner_id":0
		}`),
	)
	req = requestWithURLParam(req, "jobID", "7")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvitationExists(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithURLParam(req, "jobID", "7")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 99, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithURLParam(req, "jobID", "7")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	expectedErr := errors.New("job lookup failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/job-invitations/jobs/7",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithURLParam(req, "jobID", "7")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListSent_Success(t *testing.T) {
	repo := &mockRepository{
		listSentByClientIDFn: func(
			context.Context,
			uint,
		) ([]JobInvitation, error) {
			return []JobInvitation{
				{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/sent",
		nil,
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.ListSent(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListSent_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/sent",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListSent(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListSent_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/sent",
		nil,
	)
	req = requestWithUser(req, 0, "client")

	recorder := httptest.NewRecorder()

	handler.ListSent(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListSent_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list sent failed")

	repo := &mockRepository{
		listSentByClientIDFn: func(
			context.Context,
			uint,
		) ([]JobInvitation, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/sent",
		nil,
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.ListSent(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListReceived_Success(t *testing.T) {
	repo := &mockRepository{
		listReceivedByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]JobInvitation, error) {
			return []JobInvitation{
				{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/me",
		nil,
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.ListReceived(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListReceived_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListReceived(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListReceived_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/me",
		nil,
	)
	req = requestWithUser(req, 0, "cleaner")

	recorder := httptest.NewRecorder()

	handler.ListReceived(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListReceived_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list received failed")

	repo := &mockRepository{
		listReceivedByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]JobInvitation, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/job-invitations/me",
		nil,
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.ListReceived(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
