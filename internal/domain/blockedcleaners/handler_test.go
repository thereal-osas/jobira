package blockedcleaners

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
	service := NewService(repo)

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
	service := NewService(&mockRepository{})

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
		"/block-cleaners/8",
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
		http.MethodPost,
		"/block-cleaners/nope",
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
		http.MethodPost,
		"/block-cleaners/",
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

func TestHandler_Block_Success(t *testing.T) {
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
			block *BlockedCleaner,
		) error {
			block.ID = 12
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/block-cleaners/8",
		strings.NewReader(`{
			"reason":"Repeated no-shows"
		}`),
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

	handler.Block(recorder, req)

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
		`"cleaner_id":8`,
	) {
		t.Fatalf(
			"expected cleaner ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"reason":"Repeated no-shows"`,
	) {
		t.Fatalf(
			"expected reason in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Block_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/block-cleaners/8",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Block(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Block_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/block-cleaners/nope",
		strings.NewReader(`{}`),
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

	handler.Block(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Block_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/block-cleaners/8",
		strings.NewReader(`{"reason":`),
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

	handler.Block(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Block_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/block-cleaners/8",
		strings.NewReader(`{
			"reason":"Test"
		}`),
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

	handler.Block(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Block_SameClientAndCleaner(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/block-cleaners/5",
		strings.NewReader(`{
			"reason":"Test"
		}`),
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"5",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Block(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Block_AlreadyBlocked(t *testing.T) {
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
		"/block-cleaners/8",
		strings.NewReader(`{
			"reason":"Repeated no-shows"
		}`),
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

	handler.Block(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Block_InternalServerError(t *testing.T) {
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
		"/block-cleaners/8",
		strings.NewReader(`{
			"reason":"Repeated no-shows"
		}`),
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

	handler.Block(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Unblock_Success(t *testing.T) {
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
		"/block-cleaners/8",
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

	handler.Unblock(recorder, req)

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
		`"message":"cleaner unblocked"`,
	) {
		t.Fatalf(
			"expected success message: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Unblock_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/block-cleaners/8",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Unblock(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Unblock_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/block-cleaners/nope",
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

	handler.Unblock(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Unblock_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/block-cleaners/8",
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

	handler.Unblock(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Unblock_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrBlockedNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/block-cleaners/8",
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

	handler.Unblock(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Unblock_InternalServerError(t *testing.T) {
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
		"/block-cleaners/8",
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

	handler.Unblock(recorder, req)

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
		) ([]BlockedCleaner, error) {
			return []BlockedCleaner{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Reason:    "Repeated no-shows",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/block-cleaners/me",
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
			"expected blocked cleaner in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]BlockedCleaner, error) {
			return []BlockedCleaner{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/block-cleaners/me",
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
		"/block-cleaners/me",
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
		"/block-cleaners/me",
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
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]BlockedCleaner, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/block-cleaners/me",
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
