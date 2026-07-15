package model

import "context"

type ContextKey string

var (
	UserContextKey ContextKey = "user"
)

func GetCurrentUserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(UserContextKey).(User)
	return user, ok
}

func SetCurrentUserInContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}
