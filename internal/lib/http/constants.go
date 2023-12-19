package http

import (
	ctx "github.com/studopolis/auth-server/internal/lib/context"
)

const (
	AuthHeader         = "Authorization"
	RequestIDHeader    = "X-Request-ID"
	RequiredRoleHeader = "X-Required-Role"
)

var (
	StatusCtxKey  = &ctx.ContextKey{Name: "Status"}
	RequestCtxKey = &ctx.ContextKey{Name: "RequestID"}
	UserCtxKey    = &ctx.ContextKey{Name: "User"}

	// todo: add HTTP errors
)
