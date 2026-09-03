package messages

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_AllMessageRoutesRegistered(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
			nil,
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
			name:   "send",
			method: http.MethodPost,
			path:   "/messages/jobs/10",
		},
		{
			name:   "list conversation",
			method: http.MethodGet,
			path:   "/messages/jobs/10",
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

func TestRegisterRoutes_AllRoutesRequireAuthentication(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
			nil,
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
			"/messages/jobs/10",
		},
		{
			http.MethodGet,
			"/messages/jobs/10",
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
			nil,
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
			"/messages/jobs/10",
		},
		{
			http.MethodPut,
			"/messages/jobs/10",
		},
		{
			http.MethodPatch,
			"/messages/jobs/10",
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
