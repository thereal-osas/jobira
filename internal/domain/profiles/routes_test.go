package profiles

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_AllProfileRoutesRegistered(
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
			name:   "search",
			method: http.MethodGet,
			path:   "/profiles/search",
		},
		{
			name:   "create",
			method: http.MethodPost,
			path:   "/profiles/",
		},
		{
			name:   "get mine",
			method: http.MethodGet,
			path:   "/profiles/me",
		},
		{
			name:   "update mine",
			method: http.MethodPut,
			path:   "/profiles/me",
		},
		{
			name:   "full history",
			method: http.MethodGet,
			path:   "/profiles/me/history",
		},
		{
			name:   "completed jobs",
			method: http.MethodGet,
			path:   "/profiles/me/completed-jobs",
		},
		{
			name:   "cancelled jobs",
			method: http.MethodGet,
			path:   "/profiles/me/cancelled-jobs",
		},
		{
			name:   "verification",
			method: http.MethodPatch,
			path:   "/profiles/10/verification",
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

func TestRegisterRoutes_AllRoutesUseAuthentication(
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
			http.MethodGet,
			"/profiles/search",
		},
		{
			http.MethodPost,
			"/profiles/",
		},
		{
			http.MethodGet,
			"/profiles/me",
		},
		{
			http.MethodPut,
			"/profiles/me",
		},
		{
			http.MethodGet,
			"/profiles/me/history",
		},
		{
			http.MethodGet,
			"/profiles/me/completed-jobs",
		},
		{
			http.MethodGet,
			"/profiles/me/cancelled-jobs",
		},
		{
			http.MethodPatch,
			"/profiles/10/verification",
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
				"expected 401 for %s %s, got %d",
				test.method,
				test.path,
				res.Code,
			)
		}
	}
}

func TestRegisterRoutes_VerificationRequiresAdmin(
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

	req := httptest.NewRequest(
		http.MethodPatch,
		"/profiles/10/verification",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if !adminCalled {
		t.Fatal(
			"expected admin middleware to run",
		)
	}

	if res.Code != http.StatusForbidden {
		t.Fatalf(
			"expected 403 got %d",
			res.Code,
		)
	}
}

func TestRegisterRoutes_NormalRoutesDoNotUseAdmin(
	t *testing.T,
) {
	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&MockRepository{
				profile: &CleanerProfile{
					UserID: 1,
				},
			},
		),
	)

	adminCalled := false

	RegisterRoutes(
		router,
		handler,
		profilesTestAuth(
			1,
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
		"/profiles/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if adminCalled {
		t.Fatal(
			"admin middleware should not run for /profiles/me",
		)
	}

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}
