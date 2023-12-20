package jwt

import (
	"errors"
)

const (
	AuthHeaderPrefix   = "Bearer"
	AccessTokenScope   = "access"
	RefreshTokenScope  = "refresh"
	refreshTokenCookie = "_rt"

	RoleAdmin = "iam.admin"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidTokenScope = errors.New("invalid token scope")
	ErrRoleHeaderMissing = errors.New("required role header missing")
	ErrTokenMissing      = errors.New("token missing")
	ErrAccessDenied      = errors.New("denied")
)
