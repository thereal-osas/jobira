package favorites

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
		http.MethodDelete,
		"/favorites/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)

	id, err := parseIDParam(req, "cleanerID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 8 {
		t.Fatalf(
			"expected ID 8, got %d",
			id,
		)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"nope",
	)

	id, err := parseIDParam(req, "cleanerID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf(
			"expected ID 0, got %d",
			id,
		)
	}
}

func TestParseIDParam_MissingID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/",
		nil,
	)

	id, err := parseIDParam(req, "cleanerID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf(
			"expected ID 0, got %d",
			id,
		)
	}
}

func TestHandler_Save_Success(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			ctx context.Context,
			favorite *FavoriteCleaner,
		) error {
			if favorite.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					favorite.ClientID,
				)
			}

			if favorite.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					favorite.CleanerID,
				)
			}

			favorite.ID = 12

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"client_id":5`,
	) {
		t.Fatalf(
			"expected client ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"cleaner_id":8`,
	) {
		t.Fatalf(
			"expected cleaner ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Save_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{"cleaner_id":`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":0
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_ZeroClientID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_SameClientAndCleaner(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":5
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_CleanerBlocked(t *testing.T) {
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
		nil,
		blockChecker,
	)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_AlreadySaved(t *testing.T) {
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
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Save_InternalServerError(t *testing.T) {
	expectedErr := errors.New("exists check failed")

	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/favorites/",
		strings.NewReader(`{
			"cleaner_id":8
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Save(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
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
		) ([]FavoriteCleaner, error) {
			return []FavoriteCleaner{
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
		"/favorites/",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"cleaner_id":8`,
	) {
		t.Fatalf(
			"expected favorite cleaner in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]FavoriteCleaner, error) {
			return []FavoriteCleaner{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/favorites/",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/favorites/",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMine_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/favorites/",
		nil,
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list favorites failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]FavoriteCleaner, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/favorites/",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_Success(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) error {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"message":"favorite removed"`,
	) {
		t.Fatalf(
			"expected success message: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/8",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Remove_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrFavoriteNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/favorites/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Remove_InternalServerError(t *testing.T) {
	expectedErr := errors.New("remove favorite failed")

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
		"/favorites/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Remove(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
