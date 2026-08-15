package model

import "context"

type ContextKey string

var (
	UserContextKey    ContextKey = "user"
	SessionContextKey ContextKey = "session"
)

func GetCurrentUserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(UserContextKey).(User)
	return user, ok
}

func SetUserInContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetSessionFromContext(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(SessionContextKey).(Session)
	return session, ok
}

func SetSessionInContext(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}
