package jwt

import (
	"errors"
)

const (
	AccessTokenScope   = "access"
	RefreshTokenScope  = "refresh"
	refreshTokenCookie = "_rt"
	bearerHeaderPrefix = "Bearer "

	RoleAdmin = "iam.admin"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidTokenScope = errors.New("invalid token scope")
	ErrRoleHeaderMissing = errors.New("required role header missing")
	ErrTokenMissing      = errors.New("token missing")
	ErrAccessDenied      = errors.New("denied")
)
