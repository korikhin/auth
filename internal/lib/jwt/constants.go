package jwt

import (
	"errors"
)

const (
	AuthHeaderPrefix   = "Bearer"
	AccessTokenScope   = "a"
	RefreshTokenScope  = "r"
	refreshTokenCookie = "_studopolis.rt"
)

var (
	ErrTokenMissing      = errors.New("token is missing")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrTokenExpiredOnly  = errors.New("token is expired")
	ErrTokenInvalidScope = errors.New("token has invalid scope")
	// ErrRoleHeaderMissing = errors.New("required role header is missing")
	// ErrAccessDenied      = errors.New("denied")
)
