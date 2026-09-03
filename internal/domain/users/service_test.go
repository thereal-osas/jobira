package users

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	user *User
	err  error
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*User, error) {
	return m.user, m.err
}

func TestService_GetCurrentUser_Success(t *testing.T) {
	expected := &User{
		ID:       1,
		FullName: "John Cleaner",
		Email:    "john@test.com",
		Role:     "cleaner",
	}

	service := NewService(
		&mockRepository{
			user: expected,
		},
	)

	result, err := service.GetCurrentUser(
		context.Background(),
		1,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result.ID != expected.ID {
		t.Fatalf(
			"expected user id %d got %d",
			expected.ID,
			result.ID,
		)
	}
}

func TestService_GetCurrentUser_Error(t *testing.T) {
	service := NewService(
		&mockRepository{
			err: errors.New("database error"),
		},
	)

	_, err := service.GetCurrentUser(
		context.Background(),
		1,
	)

	if err == nil {
		t.Fatal(
			"expected error got nil",
		)
	}
}
