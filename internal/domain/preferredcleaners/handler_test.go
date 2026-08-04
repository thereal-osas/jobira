package preferredcleaners

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
	service := NewService(repo, nil)
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
		http.MethodDelete,
		"/preferred-cleaners/8",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "8")

	id, err := parseIDParam(req, "cleanerID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 8 {
		t.Fatalf("expected ID 8, got %d", id)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/nope",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "nope")

	id, err := parseIDParam(req, "cleanerID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf("expected ID 0, got %d", id)
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*PreferredCleaners,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/preferred-cleaners",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
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
		"/preferred-cleaners",
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

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/preferred-cleaners",
		strings.NewReader(`{"cleaner_id":`),
	)
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
		"/preferred-cleaners",
		strings.NewReader(`{
			"cleaner_id":0
		}`),
	)
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

func TestHandler_Create_AlreadyPreferred(t *testing.T) {
	repo := &mockRepository{
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
		"/preferred-cleaners",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
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

func TestHandler_Create_CleanerBlocked(t *testing.T) {
	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(
		&mockRepository{},
		blockChecker,
	)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/preferred-cleaners",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
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
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*PreferredCleaners,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/preferred-cleaners",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
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

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]PreferredCleaners, error) {
			return []PreferredCleaners{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/preferred-cleaners/me",
		nil,
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/preferred-cleaners/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMine_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/preferred-cleaners/me",
		nil,
	)
	req = requestWithUser(req, 0, "client")

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]PreferredCleaners, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/preferred-cleaners/me",
		nil,
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_Success(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/8",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "8")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/8",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Remove_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/nope",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "nope")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Remove_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/8",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "8")
	req = requestWithUser(req, 0, "client")

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_InternalServerError(t *testing.T) {
	expectedErr := errors.New("delete failed")

	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/preferred-cleaners/8",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "8")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
