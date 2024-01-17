package secrets

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/studopolis/auth-server/internal/config"

	"github.com/ilyakaznacheev/cleanenv"
)

type Keys struct {
	Private string `yaml:"jwt.keys.private"`
	Public  string `yaml:"jwt.keys.public"`
}

const (
	keyTypePrivate = "EC PRIVATE KEY"
	keyTypePublic  = "PUBLIC KEY"
)

func MustLoadKeys() *Keys {
	path := config.FetchConfigPath()
	if path == "" {
		panic("config is not set")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file does not exist: %s", path))
	}

	var keys Keys
	if err := cleanenv.ReadConfig(path, &keys); err != nil {
		panic(fmt.Sprintf("failed to read config: %v", err))
	}

	return &keys
}

func GetPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read private key file")
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != keyTypePrivate {
		return nil, fmt.Errorf("cannot decode PEM block containing private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key")
	}

	return key, nil
}

func GetPublicKey(path string) (*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key file")
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != keyTypePublic {
		return nil, fmt.Errorf("cannot decode PEM block containing public key")
	}

	untyped, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key")
	}

	key, ok := untyped.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is of the wrong type")
	}

	return key, nil
}
