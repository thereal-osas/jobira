package subscriptions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_AllSubscriptionRoutesRegistered(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
	)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "list plans",
			method: http.MethodGet,
			path:   "/subscriptions/plans",
		},
		{
			name:   "get mine",
			method: http.MethodGet,
			path:   "/subscriptions/me",
		},
		{
			name:   "create or update mine",
			method: http.MethodPost,
			path:   "/subscriptions/me",
		},
		{
			name:   "update mine status",
			method: http.MethodPatch,
			path:   "/subscriptions/me/status",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(
				test.method,
				test.path,
				nil,
			)

			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code == http.StatusNotFound {
				t.Fatalf(
					"route not registered: %s %s",
					test.method,
					test.path,
				)
			}

			if res.Code == http.StatusMethodNotAllowed {
				t.Fatalf(
					"method not registered: %s %s",
					test.method,
					test.path,
				)
			}
		})
	}
}

func TestRegisterRoutes_PlansIsPublic(t *testing.T) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{
				plans: []SubscriptionPlan{},
			},
		),
	)

	authCalled := false

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					authCalled = true
					next.ServeHTTP(w, r)
				},
			)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/plans",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if authCalled {
		t.Fatal(
			"auth middleware should not run for /subscriptions/plans",
		)
	}

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}

func TestRegisterRoutes_ProtectedRoutesRequireAuthentication(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					http.Error(
						w,
						"unauthorized",
						http.StatusUnauthorized,
					)
				},
			)
		},
	)

	tests := []struct {
		method string
		path   string
	}{
		{
			http.MethodGet,
			"/subscriptions/me",
		},
		{
			http.MethodPost,
			"/subscriptions/me",
		},
		{
			http.MethodPatch,
			"/subscriptions/me/status",
		},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			test.method,
			test.path,
			nil,
		)

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected 401 for %s %s got %d",
				test.method,
				test.path,
				res.Code,
			)
		}
	}
}

func TestRegisterRoutes_UnsupportedMethodsNotRegistered(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
	)

	tests := []struct {
		method string
		path   string
	}{
		{
			http.MethodDelete,
			"/subscriptions/me",
		},
		{
			http.MethodPut,
			"/subscriptions/me",
		},
		{
			http.MethodPost,
			"/subscriptions/me/status",
		},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			test.method,
			test.path,
			nil,
		)

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		if res.Code != http.StatusMethodNotAllowed {
			t.Fatalf(
				"expected 405 for %s %s got %d",
				test.method,
				test.path,
				res.Code,
			)
		}
	}
}
