package identity

import (
	"context"
	"testing"
)

func TestWithUserAndFromContext(t *testing.T) {
	expected := UserIdentity{
		UserID: 42,
		Email:  "cleaner@example.com",
		Role:   "cleaner",
	}

	ctx := WithUser(
		context.Background(),
		expected,
	)

	got, err := FromContext(ctx)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if got.UserID != expected.UserID {
		t.Fatalf(
			"expected user id %d, got %d",
			expected.UserID,
			got.UserID,
		)
	}

	if got.Email != expected.Email {
		t.Fatalf(
			"expected email %q, got %q",
			expected.Email,
			got.Email,
		)
	}

	if got.Role != expected.Role {
		t.Fatalf(
			"expected role %q, got %q",
			expected.Role,
			got.Role,
		)
	}
}

func TestFromContext_MissingUser(t *testing.T) {
	ctx := context.Background()

	user, err := FromContext(ctx)

	if err == nil {
		t.Fatal(
			"expected missing identity error",
		)
	}

	if user != (UserIdentity{}) {
		t.Fatalf(
			"expected zero identity, got %+v",
			user,
		)
	}

	expected :=
		"authenticated user not found in context"

	if err.Error() != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			err.Error(),
		)
	}
}

func TestFromContext_WrongValueType(
	t *testing.T,
) {
	ctx := context.WithValue(
		context.Background(),
		UserKey,
		"not-a-user",
	)

	user, err := FromContext(ctx)

	if err == nil {
		t.Fatal(
			"expected invalid context value error",
		)
	}

	if user != (UserIdentity{}) {
		t.Fatalf(
			"expected zero identity, got %+v",
			user,
		)
	}
}

func TestWithUser_PreservesExistingContext(
	t *testing.T,
) {
	type testKey string

	const key testKey = "existing"

	base := context.WithValue(
		context.Background(),
		key,
		"preserved",
	)

	ctx := WithUser(
		base,
		UserIdentity{
			UserID: 1,
			Role:   "client",
		},
	)

	if got := ctx.Value(key); got != "preserved" {
		t.Fatalf(
			"expected preserved value, got %v",
			got,
		)
	}
}

func TestWithUser_ReplacesExistingIdentity(
	t *testing.T,
) {
	ctx := WithUser(
		context.Background(),
		UserIdentity{
			UserID: 1,
			Email:  "first@example.com",
			Role:   "client",
		},
	)

	ctx = WithUser(
		ctx,
		UserIdentity{
			UserID: 2,
			Email:  "second@example.com",
			Role:   "cleaner",
		},
	)

	user, err := FromContext(ctx)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if user.UserID != 2 {
		t.Fatalf(
			"expected user id 2, got %d",
			user.UserID,
		)
	}

	if user.Email != "second@example.com" {
		t.Fatalf(
			"unexpected email %q",
			user.Email,
		)
	}

	if user.Role != "cleaner" {
		t.Fatalf(
			"unexpected role %q",
			user.Role,
		)
	}
}
