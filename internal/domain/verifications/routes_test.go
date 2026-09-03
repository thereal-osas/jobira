package verifications

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_AllVerificationRoutesRegistered(
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
			name:   "create verification",
			method: http.MethodPost,
			path:   "/verifications/",
		},
		{
			name:   "list mine",
			method: http.MethodGet,
			path:   "/verifications/me",
		},
		{
			name:   "get by id",
			method: http.MethodGet,
			path:   "/verifications/10",
		},
		{
			name:   "admin list all",
			method: http.MethodGet,
			path:   "/admin/verifications/",
		},
		{
			name:   "admin pending",
			method: http.MethodGet,
			path:   "/admin/verifications/pending",
		},
		{
			name:   "admin review",
			method: http.MethodPatch,
			path:   "/admin/verifications/10",
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
		func(next http.Handler) http.Handler {
			return next
		},
	)

	tests := []struct {
		method string
		path   string
	}{
		{
			http.MethodPost,
			"/verifications/",
		},
		{
			http.MethodGet,
			"/verifications/me",
		},
		{
			http.MethodGet,
			"/verifications/10",
		},
		{
			http.MethodGet,
			"/admin/verifications/",
		},
		{
			http.MethodGet,
			"/admin/verifications/pending",
		},
		{
			http.MethodPatch,
			"/admin/verifications/10",
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

func TestRegisterRoutes_AdminRoutesRequireAdminMiddleware(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	adminCalled := false

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					adminCalled = true

					http.Error(
						w,
						"forbidden",
						http.StatusForbidden,
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
			"/admin/verifications/",
		},
		{
			http.MethodGet,
			"/admin/verifications/pending",
		},
		{
			http.MethodPatch,
			"/admin/verifications/10",
		},
	}

	for _, test := range tests {
		adminCalled = false

		req := httptest.NewRequest(
			test.method,
			test.path,
			nil,
		)

		res := httptest.NewRecorder()

		router.ServeHTTP(res, req)

		if !adminCalled {
			t.Fatalf(
				"expected admin middleware for %s %s",
				test.method,
				test.path,
			)
		}

		if res.Code != http.StatusForbidden {
			t.Fatalf(
				"expected 403 got %d",
				res.Code,
			)
		}
	}
}

func TestRegisterRoutes_UserRoutesDoNotUseAdminMiddleware(
	t *testing.T,
) {
	router := chi.NewRouter()

	repo := &MockRepository{
		requests: []VerificationRequest{},
	}

	handler := NewHandler(
		NewService(repo),
	)

	adminCalled := false

	RegisterRoutes(
		router,
		handler,
		verificationsTestAuth(
			5,
			"cleaner",
		),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					adminCalled = true

					next.ServeHTTP(w, r)
				},
			)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if adminCalled {
		t.Fatal(
			"admin middleware should not run for user verification routes",
		)
	}

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
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
		),
	)

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
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
			"/verifications/10",
		},
		{
			http.MethodPut,
			"/verifications/me",
		},
		{
			http.MethodPost,
			"/admin/verifications/10",
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
