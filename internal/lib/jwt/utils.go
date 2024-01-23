package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	httplib "github.com/studopolis/auth-server/internal/lib/http"

	"github.com/golang-jwt/jwt/v5"
)

func GetAccessToken(r *http.Request) (string, error) {
	const op = "jwt.GetAccessToken"

	h := r.Header.Get(httplib.HeaderAuth)
	b, a, found := strings.Cut(h, fmt.Sprintf("%s ", headerAuthPrefix))

	if !found || b != "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenInvalid)
	}

	return a, nil
}

func SetAccessToken(w http.ResponseWriter, token string) {
	// const op = "jwt.SetAccessToken"

	h := fmt.Sprintf("%s %s", headerAuthPrefix, token)
	w.Header().Set(httplib.HeaderAuth, h)
}

func GetRefreshToken(r *http.Request) (string, error) {
	const op = "jwt.GetRefreshToken"

	c, err := r.Cookie(refreshTokenCookie)
	if err != nil || strings.TrimSpace(c.Value) == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	return c.Value, nil
}

func SetRefreshToken(w http.ResponseWriter, token string, exp time.Time) {
	// const op = "jwt.SetRefreshToken"

	c := &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  exp,
	}

	http.SetCookie(w, c)
}

func expiredOnly(err error) bool {
	if errors.Is(err, jwt.ErrTokenExpired) {
		if errs, ok := err.(interface{ Unwrap() []error }); ok {
			claimErrs := errs.Unwrap()[1]
			if errs, ok = claimErrs.(interface{ Unwrap() []error }); ok {
				if len(errs.Unwrap()) == 1 {
					return true
				}
			}
		}
	}
	return false
}
