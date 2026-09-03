package bookings

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newHandlerForTest(repo Repository) *Handler {
	service := NewService(
		repo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	return NewHandler(service)
}

func requestWithBookingID(
	method string,
	target string,
	rawID string,
) *http.Request {
	req := httptest.NewRequest(method, target, nil)

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("bookingID", rawID)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
	)

	return req.WithContext(ctx)
}

func TestNewHandler(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service to be assigned")
	}
}

func TestParseIDParam_Success(t *testing.T) {
	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)

	id, err := parseIDParam(req, "bookingID")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if id != 12 {
		t.Fatalf("expected ID 12, got %d", id)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/nope",
		"nope",
	)

	id, err := parseIDParam(req, "bookingID")

	if err == nil {
		t.Fatal("expected error")
	}

	if id != 0 {
		t.Fatalf("expected ID 0, got %d", id)
	}
}

func TestParseIDParam_MissingID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings",
		nil,
	)

	id, err := parseIDParam(req, "bookingID")

	if err == nil {
		t.Fatal("expected error")
	}

	if id != 0 {
		t.Fatalf("expected ID 0, got %d", id)
	}
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

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			booking.ID = 12
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	startAt := time.Now().Add(48 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	body := fmt.Sprintf(
		`{
		"job_id": 10,
		"cleaner_id": 30,
		"scheduled_at": %q,
		"scheduled_end_at": %q
	}`,
		startAt.Format(time.RFC3339),
		endAt.Format(time.RFC3339),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(body),
	)
	req = requestWithUser(req, 5, "client")

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
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(`{"job_id":3,"cleaner_id":8}`),
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

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(`{"job_id":`),
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(`{"job_id":0,"cleaner_id":0}`),
	)
	req = requestWithUser(req, 5, "client")

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

func TestHandler_Create_BookingExists(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			return ErrBookingExists
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(`{"job_id":3,"cleaner_id":8}`),
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	expectedErr := errors.New("database failed")

	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(`{"job_id":3,"cleaner_id":8}`),
	)
	req = requestWithUser(req, 5, "client")

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

func TestHandler_GetByID_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        12,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/nope",
		"nope",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_GetByID_InvalidInput(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, ErrInvalidInput
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, ErrBookingNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        12,
				ClientID:  5,
				CleanerID: 8,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)
	req = requestWithUser(req, 99, "client")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_InternalServerError(t *testing.T) {
	expectedErr := errors.New("database failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodGet,
		"/bookings/12",
		"12",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

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
		listByUserIDFn: func(ctx context.Context, userID uint) ([]Booking, error) {
			if userID != 5 {
				t.Fatalf("expected user ID 5, got %d", userID)
			}

			return []Booking{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				},
				{
					ID:        2,
					ClientID:  9,
					CleanerID: 5,
					Status:    "pending",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings/me",
		nil,
	)
	req = requestWithUser(req, 5, "client")

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
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings/me",
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
	repo := &mockRepository{
		listByUserIDFn: func(ctx context.Context, userID uint) ([]Booking, error) {
			return nil, ErrInvalidInput
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings/me",
		nil,
	)
	req = requestWithUser(req, 5, "client")

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
	expectedErr := errors.New("database failed")

	repo := &mockRepository{
		listByUserIDFn: func(ctx context.Context, userID uint) ([]Booking, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings/me",
		nil,
	)
	req = requestWithUser(req, 5, "client")

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

func TestHandler_Confirm_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "pending",
				}, nil
			}

			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/confirm",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Confirm_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/confirm",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Confirm_InvalidID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/nope/confirm",
		"nope",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Confirm_InvalidInput(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/confirm",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Confirm_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, ErrBookingNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/confirm",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Confirm_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/confirm",
		"1",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Confirm_InternalServerError(t *testing.T) {
	expectedErr := errors.New("update failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/confirm",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Confirm(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Start_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/start",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Start_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/start",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Start_InvalidID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/nope/start",
		"nope",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Start_InvalidInput(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/start",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Start_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, ErrBookingNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/start",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Start_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/start",
		"1",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Start_InternalServerError(t *testing.T) {
	expectedErr := errors.New("update failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/start",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Start(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Complete_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "in_progress",
				}, nil
			}

			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
		markJobCompletedFn: func(ctx context.Context, jobID uint) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/complete",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Complete_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/complete",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Complete_InvalidID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/nope/complete",
		"nope",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Complete_InvalidInput(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/complete",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Complete_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, ErrBookingNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/complete",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Complete_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/complete",
		"1",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Complete_InternalServerError(t *testing.T) {
	expectedErr := errors.New("complete failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/complete",
		"1",
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Complete(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Cancel_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:                 1,
				ClientID:           5,
				CleanerID:          8,
				Status:             "cancelled",
				CancellationReason: "Plans changed",
			}, nil
		},
		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			if bookingID != 1 {
				t.Fatalf("expected booking id 1, got %d", bookingID)
			}

			if cancelledBy != 5 {
				t.Fatalf("expected cancelled by 5, got %d", cancelledBy)
			}

			if reason != "Plans changed" {
				t.Fatalf("unexpected reason %q", reason)
			}
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)
	req.Body = io.NopCloser(
		strings.NewReader(`{"reason":"Plans changed"}`),
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Cancel_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Cancel_InvalidID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/nope/cancel",
		"nope",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Cancel_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)
	req.Body = io.NopCloser(
		strings.NewReader(`{"reason":`),
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Cancel_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)
	req.Body = io.NopCloser(
		strings.NewReader(`{"reason":"   "}`),
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Cancel_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, ErrBookingNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)
	req.Body = io.NopCloser(
		strings.NewReader(`{"reason":"Plans changed"}`),
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Cancel_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)
	req.Body = io.NopCloser(
		strings.NewReader(`{"reason":"Plans changed"}`),
	)
	req = requestWithUser(req, 99, "client")

	recorder := httptest.NewRecorder()

	handler.Cancel(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Cancel_InternalServerError(t *testing.T) {
	expectedErr := errors.New("cancel failed")

	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			id uint,
		) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},

		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			if bookingID != 1 {
				t.Fatalf(
					"expected booking id 1, got %d",
					bookingID,
				)
			}

			if cancelledBy != 5 {
				t.Fatalf(
					"expected cancelled by 5, got %d",
					cancelledBy,
				)
			}

			if reason != "Plans changed" {
				t.Fatalf(
					"unexpected reason %q",
					reason,
				)
			}

			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/cancel",
		"1",
	)

	req.Body = io.NopCloser(
		strings.NewReader(
			`{"reason":"Plans changed"}`,
		),
	)

	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Cancel(
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

func TestHandler_Close_Success(t *testing.T) {
	getCalls := 0
	rating := 5
	wouldHireAgain := true

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				}, nil
			}

			return &Booking{
				ID:                        1,
				JobID:                     7,
				ClientID:                  5,
				CleanerID:                 8,
				Status:                    "closed",
				ClosureStatus:             "client_confirmed",
				ClosureComment:            "Excellent cleaner",
				ClientConfirmedCompletion: true,
				ClientWouldHireAgain:      &wouldHireAgain,
				ClientRating:              &rating,
			}, nil
		},
		closeWithTransactionFn: func(
			ctx context.Context,
			transaction CloseBookingTransaction,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/close",
		"1",
	)
	req.Body = io.NopCloser(strings.NewReader(`{
		"client_rating": 5,
		"client_would_hire_again": true,
		"closure_comment": "Excellent cleaner",
		"add_to_favourites": true,
		"set_as_preffered": true
	}`))
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Close(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Close_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/close",
		"1",
	)

	recorder := httptest.NewRecorder()

	handler.Close(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Close_InvalidID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/nope/close",
		"nope",
	)
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Close(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Close_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/close",
		"1",
	)
	req.Body = io.NopCloser(strings.NewReader(`{"client_rating":`))
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Close(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Close_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/close",
		"1",
	)
	req.Body = io.NopCloser(strings.NewReader(`{
		"client_rating": 0,
		"closure_comment": ""
	}`))
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Close(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Close_InternalServerError(t *testing.T) {
	expectedErr := errors.New("close transaction failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
		closeWithTransactionFn: func(
			ctx context.Context,
			transaction CloseBookingTransaction,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := requestWithBookingID(
		http.MethodPatch,
		"/bookings/1/close",
		"1",
	)
	req.Body = io.NopCloser(strings.NewReader(`{
		"client_rating": 5,
		"client_would_hire_again": true,
		"closure_comment": "Excellent cleaner"
	}`))
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.Close(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
