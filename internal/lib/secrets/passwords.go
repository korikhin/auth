package secrets

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	cost int = 7
)

func GenerateFromPassword(p string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(p), cost)
}

func CompareHashAndPassword(h []byte, p string) error {
	return bcrypt.CompareHashAndPassword(h, []byte(p))
}
