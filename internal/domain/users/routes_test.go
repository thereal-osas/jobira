package users

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_UserMeRegistered(t *testing.T) {

	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/user/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal(
			"expected route to exist",
		)
	}
}

func TestRegisterRoutes_UserMeRequiresAuth(t *testing.T) {

	router := chi.NewRouter()

	handler := NewHandler(
		NewService(
			&mockRepository{},
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/user/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			rec.Code,
		)
	}
}
