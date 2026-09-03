package repeatbookings

import (
	"context"
	"errors"
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
		nil,
	)

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
		"/repeat-bookings/cleaners/8",
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
		http.MethodGet,
		"/repeat-bookings/cleaners/nope",
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
		createJobFn: func(
			context.Context,
			uint,
			CreateRepeatBookingRequest,
		) (uint, error) {
			return 12, nil
		},
		createInvitationFn: func(
			context.Context,
			uint,
			uint,
			uint,
			string,
		) error {
			return nil
		},
		createRepeatBookingFn: func(
			ctx context.Context,
			booking *RepeatBooking,
		) error {
			booking.ID = 20
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/cleaners/8",
		strings.NewReader(`{
			"title":"Weekly domestic clean",
			"description":"Clean a two-bedroom flat",
			"location":"East London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":70,
			"message":"Would you clean for me again?"
		}`),
	)
	req = requestWithURLParam(req, "cleanerID", "8")
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
		"/repeat-bookings/cleaners/8",
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
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/cleaners/nope",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(req, "cleanerID", "nope")
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

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/cleaners/8",
		strings.NewReader(`{"title":`),
	)
	req = requestWithURLParam(req, "cleanerID", "8")
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
		"/repeat-bookings/cleaners/8",
		strings.NewReader(`{
			"title":"",
			"description":"",
			"location":"",
			"job_type":""
		}`),
	)
	req = requestWithURLParam(req, "cleanerID", "8")
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

func TestHandler_Create_InternalServerError(t *testing.T) {
	expectedErr := errors.New("create job failed")

	repo := &mockRepository{
		createJobFn: func(
			context.Context,
			uint,
			CreateRepeatBookingRequest,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/cleaners/8",
		strings.NewReader(`{
			"title":"Weekly clean",
			"description":"Clean a flat",
			"location":"London",
			"job_type":"domestic"
		}`),
	)
	req = requestWithURLParam(req, "cleanerID", "8")
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

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RepeatBooking, error) {
			return []RepeatBooking{
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
		"/repeat-bookings/me",
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
		"/repeat-bookings/me",
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

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RepeatBooking, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/repeat-bookings/me",
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

func TestHandler_BookAgain_Success(t *testing.T) {
	newBookingID := uint(40)

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
		createBookAgainTransactionFn: func(
			ctx context.Context,
			input BookAgainTransaction,
		) (*RepeatBookingRequest, error) {
			return &RepeatBookingRequest{
				ID:                50,
				OriginalBookingID: input.OriginalBookingID,
				NewBookingID:      &newBookingID,
				ClientID:          input.ClientID,
				CleanerID:         input.CleanerID,
				JobID:             input.JobID,
				ScheduledAt:       input.ScheduledAt,
				Status:            input.Status,
				Message:           input.Message,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{
	"scheduled_at":"2026-09-10T10:00:00Z",
	"scheduled_end_at":"2026-09-10T12:00:00Z",
	"message":"Please bring the same equipment."
}`),
	)
	req = requestWithURLParam(req, "bookingID", "10")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_BookAgain_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_BookAgain_InvalidBookingID(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/nope/book-again",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(req, "bookingID", "nope")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_BookAgain_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{"scheduled_at":`),
	)
	req = requestWithURLParam(req, "bookingID", "10")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_BookAgain_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{
			"scheduled_at":"",
			"message":""
		}`),
	)
	req = requestWithURLParam(req, "bookingID", "10")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_BookAgain_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{
	"scheduled_at":"2026-09-10T10:00:00Z",
	"scheduled_end_at":"2026-09-10T12:00:00Z"
}`),
	)
	req = requestWithURLParam(req, "bookingID", "10")
	req = requestWithUser(req, 99, "client")

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_BookAgain_InternalServerError(t *testing.T) {
	expectedErr := errors.New("database failed")

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},

		createRepeatBookingsFn: func(
			context.Context,
			*BookingSnapshot,
			time.Time,
			time.Time,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{
	"scheduled_at":"2026-09-10T10:00:00Z",
	"scheduled_end_at":"2026-09-10T12:00:00Z"
}`),
	)
	req = requestWithURLParam(req, "bookingID", "10")
	req = requestWithUser(req, 5, "client")

	recorder := httptest.NewRecorder()

	handler.BookAgain(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
