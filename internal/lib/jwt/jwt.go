package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/domain/models"
	httplib "github.com/studopolis/auth-server/internal/lib/http"
	"github.com/studopolis/auth-server/internal/lib/secrets"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	// UserID     uint64 `json:"uid"`
	UserRole   string `json:"uro"`
	TokenScope string `json:"scp"`
	jwt.RegisteredClaims
}

type ValidationMask struct {
	Audience string
	IssuedAt bool
	Issuer   string
	Leeway   time.Duration
	Subject  string
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
	const op = "jwt.GetAccessToken"

	h := strings.TrimSpace(r.Header.Get(httplib.AuthHeader))
	if h == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	p := strings.SplitN(h, " ", 2)
	if len(p) != 2 || p[0] != AuthHeaderPrefix {
		return "", fmt.Errorf("%s: %w", op, ErrTokenInvalid)
	}

	t := p[1]
	if t == "" {
		return "", fmt.Errorf("%s: %w", op, ErrTokenMissing)
	}

	return t, nil
}

func SetAccessToken(w http.ResponseWriter, token string) {
	// const op = "jwt.SetAccessToken"

	h := fmt.Sprintf("%s %s", AuthHeaderPrefix, token)
	w.Header().Set(httplib.AuthHeader, h)
}

func GetRefreshToken(r *http.Request) (string, error) {
	const op = "jwt.GetRefreshToken"

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

	exp, err := getExpirationTime(token)
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

func Validate(token, scope string, opts *ValidationMask) (*Claims, error) {
	const op = "jwt.Validate"

	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	}, opts.WithOptions()...)

	var isExpiredOnly bool
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			if errs, ok := err.(interface{ Unwrap() []error }); ok {
				claimErrs := errs.Unwrap()[1]
				if errs, ok = claimErrs.(interface{ Unwrap() []error }); ok {
					if len(errs.Unwrap()) == 1 {
						isExpiredOnly = true
					}
				}
			}
		}

		if !isExpiredOnly {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid && !isExpiredOnly {
		return nil, fmt.Errorf("%s: %w", op, ErrTokenInvalid)
	}

	if c.TokenScope != scope {
		return nil, fmt.Errorf("%s: %w", op, ErrTokenInvalidScope)
	}

	if isExpiredOnly {
		return c, ErrTokenExpiredOnly
	}

	return c, nil
}

func Issue(user *models.User, scope string, config config.JWT) (string, error) {
	const op = "jwt.Issue"
	var ttl time.Duration

	switch scope {
	case AccessTokenScope:
		ttl = config.AccessTTL
	case RefreshTokenScope:
		ttl = config.RefreshTTL
	default:
		return "", fmt.Errorf("%s: %w", op, ErrTokenInvalidScope)
	}

	c := &Claims{
		// UserID:     user.ID,
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
	s, err := t.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return s, nil
}

// Check required claims
func (c Claims) Validate() error {
	if c.TokenScope == "" {
		return ErrTokenInvalidScope
	}

	// if c.UserRole == "" {
	// 	return jwt.ErrTokenRequiredClaimMissing
	// }

	if c.Issuer == "" {
		return jwt.ErrTokenRequiredClaimMissing
	}

	if c.IssuedAt.IsZero() {
		return jwt.ErrTokenRequiredClaimMissing
	}

	// if type of Subject changes in the future (e.g. UUID)
	// just use: if c.Subject == "" { return jwt.ErrTokenRequiredClaimMissing }
	// or other checks if needed
	if _, err := strconv.ParseUint(c.Subject, 10, 64); err != nil {
		return jwt.ErrTokenInvalidSubject
	}

	return nil
}

func (m *ValidationMask) WithOptions() []jwt.ParserOption {
	var opts = make([]jwt.ParserOption, 0, 5)

	if m.Audience != "" {
		opts = append(opts, jwt.WithAudience(m.Audience))
	}

	if m.IssuedAt {
		opts = append(opts, jwt.WithIssuedAt())
	}

	if m.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(m.Issuer))
	}

	if m.Leeway > 0 {
		opts = append(opts, jwt.WithLeeway(m.Leeway))
	}

	if m.Subject != "" {
		opts = append(opts, jwt.WithSubject(m.Subject))
	}

	return opts
}

// WARNING: Use for getting expiration time only. Don't validate tokens with this function
func getExpirationTime(token string) (time.Time, error) {
	const op = "jwt.tokenExpiration"

	t, _, err := jwt.NewParser().ParseUnverified(token, &Claims{})
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", op, ErrTokenInvalid)
	}

	if c, ok := t.Claims.(*Claims); ok {
		return c.ExpiresAt.Time, nil
	}

	return time.Time{}, fmt.Errorf("%s: %w", op, errors.New("cannot get expiration time"))
}
