package availability

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newHandlerForTest(repo Repository) *Handler {
	return NewHandler(NewService(repo))
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
		t.Fatal("expected service to be assigned")
	}
}

func TestParseIDParam_Success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/12",
		nil,
	)

	req = requestWithURLParam(
		req,
		"availabilityID",
		"12",
	)

	id, err := parseIDParam(req, "availabilityID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 12 {
		t.Fatalf("expected ID 12, got %d", id)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/nope",
		nil,
	)

	req = requestWithURLParam(
		req,
		"availabilityID",
		"nope",
	)

	id, err := parseIDParam(req, "availabilityID")

	if err == nil {
		t.Fatal("expected error")
	}

	if id != 0 {
		t.Fatalf("expected ID 0, got %d", id)
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*CleanerAvailability,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	body := `{
		"available_date":"2026-08-10",
		"start_time":"09:00",
		"end_time":"17:00",
		"status":"available",
		"notes":"Available all day"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/",
		strings.NewReader(body),
	)
	req = requestWithUser(req, 8, "cleaner")

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
		"/availability/",
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
		"/availability/",
		strings.NewReader(`{"available_date":`),
	)
	req = requestWithUser(req, 8, "cleaner")

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
		"/availability/",
		strings.NewReader(`{
			"available_date":"",
			"start_time":"",
			"end_time":""
		}`),
	)
	req = requestWithUser(req, 8, "cleaner")

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

func TestHandler_Create_Conflict(t *testing.T) {
	repo := &mockRepository{
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/",
		strings.NewReader(`{
			"available_date":"2026-08-10",
			"start_time":"09:00",
			"end_time":"17:00"
		}`),
	)
	req = requestWithUser(req, 8, "cleaner")

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

func TestHandler_GetByID_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 99, "cleaner")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]CleanerAvailability, error) {
			return []CleanerAvailability{
				{ID: 1, CleanerID: 8},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/me",
		nil,
	)
	req = requestWithUser(req, 8, "cleaner")

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

func TestHandler_ListByCleaner_Success(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]CleanerAvailability, error) {
			return []CleanerAvailability{
				{ID: 1, CleanerID: 8},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/cleaners/8",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "8")

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			getCalls++

			return &CleanerAvailability{
				ID:            1,
				CleanerID:     8,
				AvailableDate: "2026-08-10",
				StartTime:     "09:00",
				EndTime:       "17:00",
				Status:        "available",
			}, nil
		},
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return false, nil
		},
		updateFn: func(
			context.Context,
			*CleanerAvailability,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/1",
		strings.NewReader(`{
			"available_date":"2026-08-11",
			"start_time":"10:00",
			"end_time":"18:00",
			"status":"busy",
			"notes":"Updated"
		}`),
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Conflict(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/1",
		strings.NewReader(`{
			"available_date":"2026-08-11",
			"start_time":"10:00",
			"end_time":"18:00",
			"status":"available"
		}`),
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		deleteFn: func(
			context.Context,
			uint,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
func TestHandler_CreateBlock_Success(t *testing.T) {
	repo := &mockRepository{
		createBlockFn: func(
			context.Context,
			*AvailabilityBlock,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	body := fmt.Sprintf(
		`{
			"start_at": %q,
			"end_at": %q,
			"reason": "personal appointment"
		}`,
		startAt.Format(time.RFC3339),
		endAt.Format(time.RFC3339),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/blocks",
		strings.NewReader(body),
	)

	req = req.WithContext(
		identity.WithUser(
			req.Context(),
			identity.UserIdentity{
				UserID: 8,
				Role:   "cleaner",
			},
		),
	)

	recorder := httptest.NewRecorder()

	handler.CreateBlock(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateBlock_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/blocks",
		strings.NewReader(`{"start_at":`),
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.CreateBlock(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_ListBlocks_Success(t *testing.T) {
	repo := &mockRepository{
		listBlocksByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]AvailabilityBlock, error) {
			return []AvailabilityBlock{
				{ID: 1, CleanerID: 8},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/blocks/me",
		nil,
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.ListBlocks(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_DeleteBlock_Success(t *testing.T) {
	repo := &mockRepository{
		deleteBlockFn: func(
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
		"/availability/blocks/1",
		nil,
	)
	req = requestWithURLParam(req, "blockID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.DeleteBlock(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_DeleteBlock_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteBlockFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrAvailabilityNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/blocks/1",
		nil,
	)
	req = requestWithURLParam(req, "blockID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.DeleteBlock(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
