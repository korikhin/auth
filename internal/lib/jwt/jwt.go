package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/domain/models"
	httplib "github.com/studopolis/auth-server/internal/lib/http"
	"github.com/studopolis/auth-server/internal/lib/secrets"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserRole   string `json:"uro"`
	TokenScope string `json:"scp"`
	jwt.RegisteredClaims
}

var (
	publicKey  interface{}
	privateKey interface{}
)

func init() {
	var err error
	publicKey, err = secrets.GetPublicKey()
	if err != nil {
		panic(fmt.Sprintf("Error loading public key: %v", err))
	}

	privateKey, err = secrets.GetPrivateKey()
	if err != nil {
		panic(fmt.Sprintf("Error loading private key: %v", err))
	}
}

func GetAccessToken(r *http.Request) (string, error) {
	const op = "jwt.AccessToken"

	h := strings.TrimSpace(r.Header.Get(httplib.AuthHeader))
	if h == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	t := strings.TrimSpace(strings.TrimPrefix(h, bearerHeaderPrefix))
	if t == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	return t, nil
}

func SetAccessToken(w http.ResponseWriter, token string) {
	// const op = "jwt.SetAccessToken"

	h := fmt.Sprintf("%s%s", bearerHeaderPrefix, token)
	w.Header().Set(httplib.AuthHeader, h)
}

func GetRefreshToken(r *http.Request) (string, error) {
	const op = "jwt.RefreshToken"

	c, err := r.Cookie(refreshTokenCookie)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	t := strings.TrimSpace(c.Value)
	if t == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	return t, nil
}

func SetRefreshToken(w http.ResponseWriter, token string) error {
	const op = "jwt.SetRefreshToken"

	exp, err := tokenExpiration(token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c := http.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  exp,
	}

	http.SetCookie(w, &c)
	return nil
}

func Validate(token string) (*Claims, error) {
	const op = "jwt.Validate"

	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	return c, nil
}

func Issue(user *models.User, scope string, config *config.JWT) (string, error) {
	const op = "jwt.Issue"
	var ttl time.Duration

	switch scope {
	case AccessTokenScope:
		ttl = config.AccessTTL
	case RefreshTokenScope:
		ttl = config.RefreshTTL
	default:
		return "", fmt.Errorf("%s: %w", op, ErrInvalidTokenScope)
	}

	c := &Claims{
		UserRole:   user.Role,
		TokenScope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprint(user.ID),
			Issuer:    config.Issuer,
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256, c)
	signed, err := t.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return signed, nil
}

func tokenExpiration(token string) (time.Time, error) {
	const op = "jwt.tokenExpiration"

	t, _, err := jwt.NewParser().ParseUnverified(token, &Claims{})
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	if c, ok := t.Claims.(*Claims); ok {
		return c.ExpiresAt.Time, nil
	}

	return time.Time{}, fmt.Errorf("%s: %w", op, errors.New("cannot extract expiration time"))
}
