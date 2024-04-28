package jwt

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/korikhin/auth/internal/config"
	"github.com/korikhin/auth/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

const (
	headerAuthPrefix   = "Bearer"
	refreshTokenCookie = "_korikhin.rt"
	scopeAccess        = "a"
	scopeRefresh       = "r"
)

var (
	ErrTokenMissing      = errors.New("token is missing")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrTokenExpiredOnly  = errors.New("token is expired")
	ErrTokenInvalidScope = errors.New("token has invalid scope")
	// ErrRoleHeaderMissing = errors.New("required role header is missing")
	// ErrAccessDenied      = errors.New("denied")
)

type Claims struct {
	jwt.RegisteredClaims

	// UserID     uint64 `json:"uid"`
	// UserRole   string `json:"uro"`
	TokenScope string `json:"scp"`
}

// Check required claims
func (c Claims) Validate() error {
	if c.TokenScope != scopeAccess && c.TokenScope != scopeRefresh {
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

	// NOTE: The type of Subject may change to string
	if _, err := strconv.ParseUint(c.Subject, 10, 64); err != nil {
		return jwt.ErrTokenInvalidSubject
	}

	return nil
}

type ValidationOptions struct {
	Audience string
	Issuer   string
	Leeway   time.Duration
	Subject  string
}

func (opts *ValidationOptions) WithOptions() []jwt.ParserOption {
	var p = make([]jwt.ParserOption, 0, 5)

	// Required
	p = append(p, jwt.WithIssuedAt())

	if opts.Audience != "" {
		p = append(p, jwt.WithAudience(opts.Audience))
	}

	if opts.Issuer != "" {
		p = append(p, jwt.WithIssuer(opts.Issuer))
	}

	if opts.Leeway > 0 {
		p = append(p, jwt.WithLeeway(opts.Leeway))
	}

	if opts.Subject != "" {
		p = append(p, jwt.WithSubject(opts.Subject))
	}

	return p
}

type JWTService struct {
	pk      interface{}
	pubk    interface{}
	once    sync.Once
	Options config.JWT
}

func NewService(c config.JWT) *JWTService {
	return &JWTService{Options: c}
}

// Lazy load
func (a *JWTService) loadKeys() {
	const fatalMsg = "failed to initialize key management: please check system configuration"

	a.once.Do(func() {
		pk, err := getPrivateKey()
		if err != nil {
			log.Fatal(fatalMsg)
		}
		a.pk = pk

		pubk, err := getPublicKey()
		if err != nil {
			log.Fatal(fatalMsg)
		}
		a.pubk = pubk
	})
}

func (a *JWTService) validate(token, scope string, opts ValidationOptions) (*Claims, error) {
	const op = "jwt.Validate"

	a.loadKeys()

	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.pubk, nil
	}, opts.WithOptions()...)

	var isExpiredOnly bool
	if err != nil {
		isExpiredOnly = ExpiredOnly(err)
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

func (a *JWTService) ValidateAccess(token string, opts ValidationOptions) (*Claims, error) {
	return a.validate(token, scopeAccess, opts)
}

func (a *JWTService) ValidateRefresh(token string, opts ValidationOptions) (*Claims, error) {
	return a.validate(token, scopeRefresh, opts)
}

func (a *JWTService) issue(user models.User, scope string) (string, time.Time, error) {
	const op = "jwt.Issue"

	a.loadKeys()

	var ttl time.Duration
	switch scope {
	case scopeAccess:
		ttl = a.Options.AccessTTL
	case scopeRefresh:
		ttl = a.Options.RefreshTTL
	default:
		return "", time.Time{}, fmt.Errorf("%s: %w", op, ErrTokenInvalidScope)
	}

	exp := time.Now().Add(ttl)
	c := &Claims{
		// UserID:     user.ID,
		// UserRole:   user.Role,
		TokenScope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprint(user.ID),
			Issuer:    a.Options.Issuer,
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256, c)
	s, err := t.SignedString(a.pk)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return s, exp, nil
}

func (a *JWTService) IssueAccess(user models.User) (string, time.Time, error) {
	return a.issue(user, scopeAccess)
}

func (a *JWTService) IssueRefresh(user models.User) (string, time.Time, error) {
	return a.issue(user, scopeRefresh)
}
