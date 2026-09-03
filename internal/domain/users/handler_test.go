package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

type mockService struct {
	user *User
	err  error
}

func (m *mockService) GetCurrentUser(ctx context.Context, userID uint) (*User, error) {
	return m.user, m.err
}

func TestHandler_Me_Success(t *testing.T) {

	service := NewService(
		&mockRepository{
			user: &User{
				ID:       1,
				FullName: "John Cleaner",
				Email:    "john@test.com",
				Role:     "cleaner",
			},
		},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/user/me",
		nil,
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 1,
			Role:   "cleaner",
		},
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			rec.Code,
		)
	}

	var response User

	err := json.NewDecoder(rec.Body).Decode(&response)

	if err != nil {
		t.Fatal(err)
	}

	if response.ID != 1 {
		t.Fatalf(
			"expected user id 1 got %d",
			response.ID,
		)
	}
}

func TestHandler_Me_Unauthorized(t *testing.T) {

	service := NewService(
		&mockRepository{},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/user/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			rec.Code,
		)
	}
}

func TestHandler_Me_UserNotFound(t *testing.T) {

	service := NewService(
		&mockRepository{
			err: ErrUserNotFound,
		},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/user/me",
		nil,
	)

	req = req.WithContext(
		identity.WithUser(
			req.Context(),
			identity.UserIdentity{
				UserID: 1,
				Role:   "cleaner",
			},
		),
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			rec.Code,
		)
	}
}

func TestHandler_Me_ServiceError(t *testing.T) {

	service := NewService(
		&mockRepository{
			err: errors.New("database failed"),
		},
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/user/me",
		nil,
	)

	req = req.WithContext(
		identity.WithUser(
			req.Context(),
			identity.UserIdentity{
				UserID: 1,
				Role:   "cleaner",
			},
		),
	)

	rec := httptest.NewRecorder()

	handler.Me(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			rec.Code,
		)
	}
}
