package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
	"github.com/rodrigueghenda/jobira/internal/security/password"
)

func newTestAuthHandler(repo Repository) *Handler {
	issuer := jwtsecurity.NewIssuer("test-secret")
	service := NewService(repo, issuer)

	return NewHandler(service)
}

func TestNewHandler(t *testing.T) {
	service := NewService(
		&MockRepository{},
		jwtsecurity.NewIssuer("test-secret"),
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service to be assigned")
	}
}

func TestHandler_Register_Success(t *testing.T) {
	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			user.ID = 1
			return nil
		},
	}

	handler := newTestAuthHandler(repo)

	body := `{
		"full_name": "John Smith",
		"email": "john@test.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("expected token in response: %s", rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"john@test.com"`) {
		t.Fatalf("expected email in response: %s", rec.Body.String())
	}
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	handler := newTestAuthHandler(&MockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"full_name":`),
	)

	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if !strings.Contains(rec.Body.String(), "invalid request body") {
		t.Fatalf(
			"expected invalid request body error, got %s",
			rec.Body.String(),
		)
	}
}

func TestHandler_Register_InvalidInput(t *testing.T) {
	handler := newTestAuthHandler(&MockRepository{})

	body := `{
		"full_name": "",
		"email": "john@test.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), ErrInvalidInput.Error()) {
		t.Fatalf(
			"expected %q, got %s",
			ErrInvalidInput.Error(),
			rec.Body.String(),
		)
	}
}

func TestHandler_Register_EmailAlreadyInUse(t *testing.T) {
	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			return ErrEmailAlreadyInUse
		},
	}

	handler := newTestAuthHandler(repo)

	body := `{
		"full_name": "John Smith",
		"email": "john@test.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), ErrEmailAlreadyInUse.Error()) {
		t.Fatalf(
			"expected %q, got %s",
			ErrEmailAlreadyInUse.Error(),
			rec.Body.String(),
		)
	}
}

func TestHandler_Register_InternalServerError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			return expected
		},
	}

	handler := newTestAuthHandler(repo)

	body := `{
		"full_name": "John Smith",
		"email": "john@test.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), expected.Error()) {
		t.Fatalf(
			"expected %q, got %s",
			expected.Error(),
			rec.Body.String(),
		)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	hash, err := password.Hash("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:           1,
				FullName:     "John Smith",
				Email:        "john@test.com",
				PasswordHash: hash,
				Role:         "user",
			}, nil
		},
	}

	handler := newTestAuthHandler(repo)

	body := `{
		"email": "john@test.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("expected token in response: %s", rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"john@test.com"`) {
		t.Fatalf("expected email in response: %s", rec.Body.String())
	}
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
	handler := newTestAuthHandler(&MockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"email":`),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if !strings.Contains(rec.Body.String(), "invalid requestbody") {
		t.Fatalf(
			"expected invalid requestbody error, got %s",
			rec.Body.String(),
		)
	}
}

func TestHandler_Login_UnknownField(t *testing.T) {
	handler := newTestAuthHandler(&MockRepository{})

	body := `{
		"email": "john@test.com",
		"password": "password123",
		"unexpected": "field"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_Login_InvalidInput(t *testing.T) {
	handler := newTestAuthHandler(&MockRepository{})

	body := `{
		"email": "",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), ErrInvalidInput.Error()) {
		t.Fatalf(
			"expected %q, got %s",
			ErrInvalidInput.Error(),
			rec.Body.String(),
		)
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	hash, err := password.Hash("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:           1,
				FullName:     "John Smith",
				Email:        email,
				PasswordHash: hash,
				Role:         "user",
			}, nil
		},
	}

	handler := newTestAuthHandler(repo)

	body := `{
		"email": "john@test.com",
		"password": "wrong-password"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), ErrInvalidCredentials.Error()) {
		t.Fatalf(
			"expected %q, got %s",
			ErrInvalidCredentials.Error(),
			rec.Body.String(),
		)
	}
}

func TestHandler_Login_InternalServerError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return nil, expected
		},
	}

	handler := newTestAuthHandler(repo)

	body := `{
		"email": "john@test.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}

	if !strings.Contains(rec.Body.String(), expected.Error()) {
		t.Fatalf(
			"expected %q, got %s",
			expected.Error(),
			rec.Body.String(),
		)
	}
}

func TestDecodeJSON_Success(t *testing.T) {
	body := []byte(`{
		"email": "john@test.com",
		"password": "password123"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewReader(body),
	)

	var result LoginRequest

	err := decodeJSON(req, &result)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Email != "john@test.com" {
		t.Fatalf("expected email john@test.com, got %q", result.Email)
	}

	if result.Password != "password123" {
		t.Fatalf("expected password password123, got %q", result.Password)
	}
}

func TestDecodeJSON_UnknownField(t *testing.T) {
	body := []byte(`{
		"email": "john@test.com",
		"password": "password123",
		"unknown": true
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewReader(body),
	)

	var result LoginRequest

	err := decodeJSON(req, &result)
	if err == nil {
		t.Fatal("expected unknown-field error")
	}
}
