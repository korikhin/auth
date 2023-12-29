package http

import (
	ctx "github.com/studopolis/auth-server/internal/lib/context"
)

const (
	ContentTypeHeader = "Content-Type"
	AuthHeader        = "Authorization"
	RequestIDHeader   = "X-Request-ID"
	// RequiredRoleHeader = "X-Required-Role"

	ContentTypeJSON = "application/json"
)

var (
	StatusCtxKey  = &ctx.ContextKey{Name: "Status"}
	RequestCtxKey = &ctx.ContextKey{Name: "RequestID"}
	UserCtxKey    = &ctx.ContextKey{Name: "User"}

	// todo: add HTTP errors
)
