package clientnotes

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
		http.MethodGet,
		"/cleaners/8/notes",
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
		http.MethodGet,
		"/cleaners/nope/notes",
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
		http.MethodGet,
		"/cleaners/notes",
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

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			note *ClientCleanerNote,
		) error {
			if note.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					note.ClientID,
				)
			}

			if note.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					note.CleanerID,
				)
			}

			if note.Note != "Reliable cleaner" {
				t.Fatalf(
					"expected note Reliable cleaner, got %q",
					note.Note,
				)
			}

			note.ID = 12

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/8/notes",
		strings.NewReader(`{
			"note":"Reliable cleaner"
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

	handler.Create(recorder, req)

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

	if !strings.Contains(
		recorder.Body.String(),
		`"note":"Reliable cleaner"`,
	) {
		t.Fatalf(
			"expected note in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/8/notes",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/nope/notes",
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

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/8/notes",
		strings.NewReader(`{"note":`),
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

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/8/notes",
		strings.NewReader(`{
			"note":"   "
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

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_ZeroClientID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/8/notes",
		strings.NewReader(`{
			"note":"Reliable cleaner"
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

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	expectedErr := errors.New("create note failed")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*ClientCleanerNote,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaners/8/notes",
		strings.NewReader(`{
			"note":"Reliable cleaner"
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

	handler.Create(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleaner_Success(t *testing.T) {
	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
			uint,
		) ([]ClientCleanerNote, error) {
			return []ClientCleanerNote{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Note:      "Reliable",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaners/8/notes",
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

	handler.GetByCleaner(recorder, req)

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
		`"note":"Reliable"`,
	) {
		t.Fatalf(
			"expected note in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleaner_Empty(t *testing.T) {
	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
			uint,
		) ([]ClientCleanerNote, error) {
			return []ClientCleanerNote{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaners/8/notes",
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

	handler.GetByCleaner(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleaner_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaners/8/notes",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetByCleaner(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetByCleaner_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaners/nope/notes",
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

	handler.GetByCleaner(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleaner_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaners/8/notes",
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

	handler.GetByCleaner(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByCleaner_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list notes failed")

	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
			uint,
		) ([]ClientCleanerNote, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaners/8/notes",
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

	handler.GetByCleaner(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	repo := &mockRepository{
		updateFn: func(
			ctx context.Context,
			noteID uint,
			clientID uint,
			note string,
		) error {
			if noteID != 12 {
				t.Fatalf(
					"expected note ID 12, got %d",
					noteID,
				)
			}

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if note != "Updated note" {
				t.Fatalf(
					"expected Updated note, got %q",
					note,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/12",
		strings.NewReader(`{
			"note":"Updated note"
		}`),
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

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
		`"message":"note updated"`,
	) {
		t.Fatalf(
			"expected success message: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/12",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Update_InvalidNoteID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/nope",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/12",
		strings.NewReader(`{"note":`),
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/12",
		strings.NewReader(`{
			"note":"   "
		}`),
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	repo := &mockRepository{
		updateFn: func(
			context.Context,
			uint,
			uint,
			string,
		) error {
			return ErrNoteNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/12",
		strings.NewReader(`{
			"note":"Updated note"
		}`),
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InternalServerError(t *testing.T) {
	expectedErr := errors.New("update note failed")

	repo := &mockRepository{
		updateFn: func(
			context.Context,
			uint,
			uint,
			string,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/notes/12",
		strings.NewReader(`{
			"note":"Updated note"
		}`),
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			ctx context.Context,
			noteID uint,
			clientID uint,
		) error {
			if noteID != 12 {
				t.Fatalf(
					"expected note ID 12, got %d",
					noteID,
				)
			}

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/notes/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

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
		`"message":"note deleted"`,
	) {
		t.Fatalf(
			"expected success message: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/notes/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Delete_InvalidNoteID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/notes/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/notes/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrNoteNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/notes/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_InternalServerError(t *testing.T) {
	expectedErr := errors.New("delete note failed")

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
		"/notes/12",
		nil,
	)
	req = requestWithURLParam(
		req,
		"noteID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
