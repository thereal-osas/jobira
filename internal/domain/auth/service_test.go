package auth

import (
	"context"
	"errors"
	"testing"

	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
	"github.com/rodrigueghenda/jobira/internal/security/password"
)

type MockRepository struct {
	CreateUserFunc     func(ctx context.Context, user *User) error
	GetUserByEmailFunc func(ctx context.Context, email string) (*User, error)
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}
	issuer := jwtsecurity.NewIssuer("secret")

	service := NewService(repo, issuer)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("repository not assigned")
	}

	if service.issuer != issuer {
		t.Fatal("issuer not assigned")
	}
}

func TestRegister_Success(t *testing.T) {
	repo := &MockRepository{}

	repo.CreateUserFunc = func(ctx context.Context, user *User) error {
		user.ID = 1
		return nil
	}

	service := NewService(
		repo,
		jwtsecurity.NewIssuer("secret"),
	)

	resp, err := service.Register(context.Background(), RegisterRequest{
		FullName: "John Smith",
		Email:    "JOHN@EMAIL.COM",
		Password: "password123",
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected response")
	}

	if resp.Token == "" {
		t.Fatal("expected token")
	}

	if resp.User.Email != "john@email.com" {
		t.Fatalf("expected normalized email got %s", resp.User.Email)
	}

	if resp.User.FullName != "John Smith" {
		t.Fatal("unexpected full name")
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
		jwtsecurity.NewIssuer("secret"),
	)

	tests := []RegisterRequest{
		{},
		{FullName: "", Email: "a@test.com", Password: "pass"},
		{FullName: "John", Email: "", Password: "pass"},
		{FullName: "John", Email: "a@test.com", Password: ""},
	}

	for _, tc := range tests {

		_, err := service.Register(context.Background(), tc)

		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput got %v", err)
		}
	}
}

func TestRegister_CreateUserError(t *testing.T) {

	expected := errors.New("database error")

	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			return expected
		},
	}

	service := NewService(
		repo,
		jwtsecurity.NewIssuer("secret"),
	)

	_, err := service.Register(context.Background(), RegisterRequest{
		FullName: "John",
		Email:    "john@test.com",
		Password: "password",
	})

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}
}

func TestLogin_Success(t *testing.T) {

	hash, _ := password.Hash("password123")

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:           1,
				FullName:     "John",
				Email:        "john@test.com",
				PasswordHash: hash,
				Role:         "user",
			}, nil
		},
	}

	service := NewService(
		repo,
		jwtsecurity.NewIssuer("secret"),
	)

	resp, err := service.Login(context.Background(), LoginRequest{
		Email:    " JOHN@TEST.COM ",
		Password: "password123",
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp.Token == "" {
		t.Fatal("expected token")
	}

	if resp.User.Email != "john@test.com" {
		t.Fatal("expected normalized email")
	}
}

func TestLogin_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
		jwtsecurity.NewIssuer("secret"),
	)

	tests := []LoginRequest{
		{},
		{Email: "", Password: "password"},
		{Email: "john@test.com", Password: ""},
		{Email: "   ", Password: "password"},
		{Email: "john@test.com", Password: "   "},
	}

	for _, tc := range tests {
		_, err := service.Login(context.Background(), tc)

		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	}
}

func TestLogin_GetUserByEmailError(t *testing.T) {
	expected := errors.New("database error")

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return nil, expected
		},
	}

	service := NewService(
		repo,
		jwtsecurity.NewIssuer("secret"),
	)

	_, err := service.Login(context.Background(), LoginRequest{
		Email:    "john@test.com",
		Password: "password123",
	})

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	hash, err := password.Hash("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &MockRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:           1,
				FullName:     "John",
				Email:        "john@test.com",
				PasswordHash: hash,
				Role:         "user",
			}, nil
		},
	}

	service := NewService(
		repo,
		jwtsecurity.NewIssuer("secret"),
	)

	_, err = service.Login(context.Background(), LoginRequest{
		Email:    "john@test.com",
		Password: "wrong-password",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRegister_NormalizesInput(t *testing.T) {
	repo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, user *User) error {
			user.ID = 7

			if user.FullName != "John Smith" {
				t.Fatalf("expected trimmed full name, got %q", user.FullName)
			}

			if user.Email != "john@test.com" {
				t.Fatalf("expected normalized email, got %q", user.Email)
			}

			if user.Role != "user" {
				t.Fatalf("expected user role, got %q", user.Role)
			}

			if user.PasswordHash == "" {
				t.Fatal("expected password hash")
			}

			if user.PasswordHash == "password123" {
				t.Fatal("expected password to be hashed")
			}

			return nil
		},
	}

	service := NewService(
		repo,
		jwtsecurity.NewIssuer("secret"),
	)

	resp, err := service.Register(context.Background(), RegisterRequest{
		FullName: "  John Smith  ",
		Email:    "  JOHN@TEST.COM  ",
		Password: "  password123  ",
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.User.ID != 7 {
		t.Fatalf("expected user ID 7, got %d", resp.User.ID)
	}

	if resp.User.Email != "john@test.com" {
		t.Fatalf("expected normalized response email, got %q", resp.User.Email)
	}
}
