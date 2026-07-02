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
