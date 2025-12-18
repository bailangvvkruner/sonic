package util

import (
	"context"
)

type contextKey string

const (
	ClientIPKey  contextKey = "client_ip"
	UserAgentKey contextKey = "user_agent"
)

func GetClientIP(ctx context.Context) string {
	val := ctx.Value(ClientIPKey)
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetUserAgent(ctx context.Context) string {
	val := ctx.Value(UserAgentKey)
	if val == nil {
		return ""
	}
	return val.(string)
}

func SetClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPKey, ip)
}

func SetUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, UserAgentKey, ua)
}
