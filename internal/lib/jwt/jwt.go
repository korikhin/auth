package jwt

import (
	"fmt"
	"time"

	"github.com/studopolis/auth-server/internal/config"
	"github.com/studopolis/auth-server/internal/domain/models"
	"github.com/studopolis/auth-server/internal/lib/secrets"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserRole   string `json:"uro"`
	TokenScope string `json:"scp"`
	jwt.RegisteredClaims
}

const (
	accessTokenScope  string = "access"
	refreshTokenScope string = "refresh"

	RoleAdmin string = "iam.admin"
)

var publicKey interface{}
var privateKey interface{}

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

func Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func Issue(user *models.User, scope string, config *config.JWT) (string, error) {
	var ttl time.Duration

	switch scope {
	case accessTokenScope:
		ttl = config.AccessTTL
	case refreshTokenScope:
		ttl = config.RefreshTTL
	}

	claims := &Claims{
		UserRole:   user.Role,
		TokenScope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprint(user.ID),
			Issuer:    config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(privateKey)
}
