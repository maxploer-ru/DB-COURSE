package middleware

import (
	"context"
	"errors"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserContext struct {
	UserID int
	Role   string
}

func GetUserFromContext(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok || user == nil {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}
