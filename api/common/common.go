package common

import (
	"context"
)

type contextKey string

const (
	SessionTokenKey contextKey = "sessionToken"
)

func GetSessionToken(ctx context.Context) string {
	if token, ok := ctx.Value(SessionTokenKey).(string); ok {
		return token
	}
	return ""
}
