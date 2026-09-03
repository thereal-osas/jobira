package availability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func calendarHandlerRequest(
	method string,
	target string,
	body string,
	userID uint,
	role string,
) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		strings.NewReader(body),
	)

	if userID != 0 {
		req = req.WithContext(
			identity.WithUser(
				req.Context(),
				identity.UserIdentity{
					UserID: userID,
					Role:   role,
				},
			),
		)
	}

	return req
}

func calendarHandlerParam(
	req *http.Request,
	name string,
	value string,
) *http.Request {
	routeContext := chi.NewRouteContext()

	routeContext.URLParams.Add(
		name,
		value,
	)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
	)

	return req.WithContext(ctx)
}

func TestCalendarHandler_CreateBulk_Success(t *testing.T) {
	repo := &calendarMockRepository{
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

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodPost,
		"/availability/bulk",
		`{
			"slots": [
				{
					"available_date": "2026-12-10",
					"start_time": "09:00",
					"end_time": "12:00",
					"status": "available"
				},
				{
					"available_date": "2026-12-11",
					"start_time": "13:00",
					"end_time": "17:00",
					"status": "available"
				}
			]
		}`,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateBulk(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_CreateRecurring_Success(t *testing.T) {
	repo := &calendarMockRepository{
		createRecurringFn: func(
			_ context.Context,
			recurring *RecurringAvailability,
		) error {
			recurring.ID = 12
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodPost,
		"/availability/recurring",
		`{
			"weekday": 1,
			"start_time": "09:00",
			"end_time": "17:00",
			"status": "available"
		}`,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateRecurring(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_CreateRecurringBulk_Success(t *testing.T) {
	repo := &calendarMockRepository{
		createRecurringFn: func(
			context.Context,
			*RecurringAvailability,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodPost,
		"/availability/recurring/bulk",
		`{
			"slots": [
				{
					"weekday": 1,
					"start_time": "09:00",
					"end_time": "17:00"
				},
				{
					"weekday": 3,
					"start_time": "10:00",
					"end_time": "16:00"
				}
			]
		}`,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateRecurringBulk(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_ListRecurring_Success(t *testing.T) {
	repo := &calendarMockRepository{
		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					ID:        1,
					CleanerID: 8,
					Weekday:   1,
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodGet,
		"/availability/recurring/me",
		"",
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.ListRecurring(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_DeleteRecurring_Success(t *testing.T) {
	repo := &calendarMockRepository{
		deleteRecurringFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodDelete,
		"/availability/recurring/4",
		"",
		8,
		"cleaner",
	)

	req = calendarHandlerParam(
		req,
		"recurringID",
		"4",
	)

	recorder := httptest.NewRecorder()

	handler.DeleteRecurring(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_GetSettings_Success(t *testing.T) {
	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   120,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BufferMinutes:      30,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodGet,
		"/availability/settings/me",
		"",
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.GetSettings(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_UpdateSettings_Success(t *testing.T) {
	repo := &calendarMockRepository{
		upsertSettingsFn: func(
			context.Context,
			*AvailabilitySettings,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodPut,
		"/availability/settings/me",
		`{
			"min_notice_minutes": 120,
			"min_booking_minutes": 60,
			"max_booking_minutes": 480,
			"buffer_minutes": 30,
			"booking_horizon_days": 90,
			"timezone": "Europe/London"
		}`,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateSettings(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_CreateOverride_Success(t *testing.T) {
	repo := &calendarMockRepository{
		createOverrideFn: func(
			_ context.Context,
			override *AvailabilityOverride,
		) error {
			override.ID = 20
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodPost,
		"/availability/overrides",
		`{
			"available_date": "2026-12-25",
			"start_time": "09:00",
			"end_time": "17:00",
			"status": "unavailable",
			"reason": "Christmas"
		}`,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateOverride(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_DeleteOverride_Success(t *testing.T) {
	repo := &calendarMockRepository{
		deleteOverrideFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := calendarHandlerRequest(
		http.MethodDelete,
		"/availability/overrides/5",
		"",
		8,
		"cleaner",
	)

	req = calendarHandlerParam(
		req,
		"overrideID",
		"5",
	)

	recorder := httptest.NewRecorder()

	handler.DeleteOverride(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_GetCalendar_Success(t *testing.T) {
	date := time.Date(
		2026,
		time.December,
		7,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	repo := &calendarMockRepository{
		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   int(date.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/cleaners/8/calendar?from=2026-12-07&to=2026-12-07",
		nil,
	)

	req = calendarHandlerParam(
		req,
		"cleanerID",
		"8",
	)

	recorder := httptest.NewRecorder()

	handler.GetCalendar(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestCalendarHandler_NextAvailable_Success(t *testing.T) {
	date := time.Date(
		2026,
		time.December,
		7,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	repo := &calendarMockRepository{
		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   int(date.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/cleaners/8/next-available?from=2026-12-07&days=7",
		nil,
	)

	req = calendarHandlerParam(
		req,
		"cleanerID",
		"8",
	)

	recorder := httptest.NewRecorder()

	handler.NextAvailable(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
