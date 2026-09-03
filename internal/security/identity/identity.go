package identity

import (
	"context"
	"errors"
)

type contextKey string

const UserKey contextKey = "authenticated_user"

type UserIdentity struct {
	UserID uint
	Email  string
	Role   string
}

func WithUser(ctx context.Context, user UserIdentity) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

func FromContext(ctx context.Context) (UserIdentity, error) {
	user, ok := ctx.Value(UserKey).(UserIdentity)
	if !ok {
		return UserIdentity{}, errors.New("authenticated user not found in context")
	}

	return user, nil
}
