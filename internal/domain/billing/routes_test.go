package billing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_AllBillingRoutesRegistered(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
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
			name:   "checkout",
			method: http.MethodPost,
			path:   "/billing/checkout",
		},
		{
			name:   "portal",
			method: http.MethodPost,
			path:   "/billing/portal",
		},
		{
			name:   "validate promo",
			method: http.MethodPost,
			path:   "/billing/promo/validate",
		},
		{
			name:   "stripe webhook",
			method: http.MethodPost,
			path:   "/stripe/webhook",
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

			router.ServeHTTP(
				res,
				req,
			)

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

func TestRegisterRoutes_BillingRoutesRequireAuthentication(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
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
			http.MethodPost,
			"/billing/checkout",
		},
		{
			http.MethodPost,
			"/billing/portal",
		},
		{
			http.MethodPost,
			"/billing/promo/validate",
		},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			test.method,
			test.path,
			nil,
		)

		res := httptest.NewRecorder()

		router.ServeHTTP(
			res,
			req,
		)

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

func TestRegisterRoutes_WebhookIsPublic(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
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

					next.ServeHTTP(
						w,
						r,
					)
				},
			)
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/stripe/webhook",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if authCalled {
		t.Fatal(
			"auth middleware should not run for Stripe webhook",
		)
	}

	if res.Code == http.StatusNotFound {
		t.Fatal(
			"expected webhook route registered",
		)
	}
}

func TestRegisterRoutes_UnsupportedMethodsNotRegistered(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
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
			http.MethodGet,
			"/billing/checkout",
		},
		{
			http.MethodDelete,
			"/billing/portal",
		},
		{
			http.MethodPatch,
			"/billing/promo/validate",
		},
		{
			http.MethodGet,
			"/stripe/webhook",
		},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			test.method,
			test.path,
			nil,
		)

		res := httptest.NewRecorder()

		router.ServeHTTP(
			res,
			req,
		)

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
