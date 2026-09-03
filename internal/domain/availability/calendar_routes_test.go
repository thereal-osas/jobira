package availability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func calendarRouteAuthMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: 8,
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func TestCalendarRoutes_NewEndpoints(t *testing.T) {
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

		createRecurringFn: func(
			context.Context,
			*RecurringAvailability,
		) error {
			return nil
		},

		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   1,
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},

		deleteRecurringFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return nil
		},

		upsertSettingsFn: func(
			context.Context,
			*AvailabilitySettings,
		) error {
			return nil
		},

		createOverrideFn: func(
			context.Context,
			*AvailabilityOverride,
		) error {
			return nil
		},

		deleteOverrideFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		calendarRouteAuthMiddleware,
	)

	tests := []struct {
		name           string
		method         string
		target         string
		body           string
		expectedStatus int
	}{
		{
			name:   "bulk availability",
			method: http.MethodPost,
			target: "/availability/bulk",
			body: `{
				"slots": [
					{
						"available_date": "2026-12-10",
						"start_time": "09:00",
						"end_time": "12:00"
					}
				]
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "create recurring",
			method: http.MethodPost,
			target: "/availability/recurring",
			body: `{
				"weekday": 1,
				"start_time": "09:00",
				"end_time": "17:00"
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "bulk recurring",
			method: http.MethodPost,
			target: "/availability/recurring/bulk",
			body: `{
				"slots": [
					{
						"weekday": 1,
						"start_time": "09:00",
						"end_time": "17:00"
					}
				]
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "list recurring",
			method:         http.MethodGet,
			target:         "/availability/recurring/me",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete recurring",
			method:         http.MethodDelete,
			target:         "/availability/recurring/4",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get settings",
			method:         http.MethodGet,
			target:         "/availability/settings/me",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "update settings",
			method: http.MethodPut,
			target: "/availability/settings/me",
			body: `{
				"min_notice_minutes": 120,
				"min_booking_minutes": 60,
				"max_booking_minutes": 480,
				"buffer_minutes": 30,
				"booking_horizon_days": 90,
				"timezone": "Europe/London"
			}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "create override",
			method: http.MethodPost,
			target: "/availability/overrides",
			body: `{
				"available_date": "2026-12-25",
				"start_time": "09:00",
				"end_time": "17:00",
				"status": "unavailable",
				"reason": "Christmas"
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "delete override",
			method:         http.MethodDelete,
			target:         "/availability/overrides/5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "calendar",
			method:         http.MethodGet,
			target:         "/availability/cleaners/8/calendar?from=2026-12-07&to=2026-12-07",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "next available",
			method:         http.MethodGet,
			target:         "/availability/cleaners/8/next-available?from=2026-12-07&days=7",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				var body *strings.Reader

				if tt.body == "" {
					body = strings.NewReader("")
				} else {
					body = strings.NewReader(tt.body)
				}

				req := httptest.NewRequest(
					tt.method,
					tt.target,
					body,
				)

				recorder := httptest.NewRecorder()

				router.ServeHTTP(
					recorder,
					req,
				)

				if recorder.Code != tt.expectedStatus {
					t.Fatalf(
						"expected %d, got %d: %s",
						tt.expectedStatus,
						recorder.Code,
						recorder.Body.String(),
					)
				}
			},
		)
	}
}
