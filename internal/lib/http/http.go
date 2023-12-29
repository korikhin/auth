package http

import (
	ctx "github.com/studopolis/auth-server/internal/lib/context"
)

// Headers
const (
	HeaderAuth            = "Authorization"
	HeaderCacheControl    = "Cache-Control"
	HeaderContentType     = "Content-Type"
	HeaderDNT             = "DNT"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderRange           = "Range"
	HeaderUserAgent       = "User-Agent"

	// Custom HTTP header for request ID
	HeaderRequestID = "X-Request-ID"
	// RequiredRoleHeader = "X-Required-Role"
)

// Content types
const (
	ContentTypeJSON = "application/json"
)

// Context keys
var (
	StatusCtxKey  = &ctx.ContextKey{Name: "Status"}
	RequestCtxKey = &ctx.ContextKey{Name: "RequestID"}
	UserCtxKey    = &ctx.ContextKey{Name: "User"}
)
