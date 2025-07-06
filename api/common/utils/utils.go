package utils

import (
	"context"
)

type AuthContext struct {
	UserID       int64
	StytchUserID string
	SessionToken string
}

type ctxKey string

const authCtxKey ctxKey = "auth"

func WithAuthContext(ctx context.Context, auth AuthContext) context.Context {
	return context.WithValue(ctx, authCtxKey, auth)
}

func GetAuthContext(ctx context.Context) (AuthContext, bool) {
	val, ok := ctx.Value(authCtxKey).(AuthContext)
	return val, ok
}

func GetSessionToken(ctx context.Context) string {
	if auth, ok := ctx.Value(authCtxKey).(AuthContext); ok {
		return auth.SessionToken
	}
	return ""
}

func GetUserID(ctx context.Context) int64 {
	if auth, ok := ctx.Value(authCtxKey).(AuthContext); ok {
		return auth.UserID
	}
	return 0
}

func GetStytchUserID(ctx context.Context) string {
	if auth, ok := ctx.Value(authCtxKey).(AuthContext); ok {
		return auth.StytchUserID
	}
	return ""
}
