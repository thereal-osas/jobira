package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
)

func TestRegisterRoutes(t *testing.T) {
	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			user.ID = 1
			return nil
		},
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return nil, ErrInvalidCredentials
		},
	}

	service := NewService(repo, jwtsecurity.NewIssuer("secret"))
	handler := NewHandler(service)

	router := chi.NewRouter()
	RegisterRoutes(router, handler)

	t.Run("POST /auth/register", func(t *testing.T) {
		body := `{
			"full_name":"John Smith",
			"email":"john@test.com",
			"password":"password123"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/auth/register",
			strings.NewReader(body),
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf(
				"expected %d got %d (%s)",
				http.StatusCreated,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("POST /auth/login", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			strings.NewReader(`{"email":"john@test.com","password":"wrong"}`),
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected %d got %d (%s)",
				http.StatusUnauthorized,
				rec.Code,
				rec.Body.String(),
			)
		}
	})
}

