package secrets

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	keyTypePublic  string = "PUBLIC KEY"
	keyTypePrivate string = "EC PRIVATE KEY"
)

func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	path := os.Getenv("PRIVATE_KEY_PATH")
	if path == "" {
		return nil, fmt.Errorf("PRIVATE_KEY_PATH is not set")
	}

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

func GetPublicKey() (*ecdsa.PublicKey, error) {
	path := os.Getenv("PUBLIC_KEY_PATH")
	if path == "" {
		return nil, fmt.Errorf("PUBLIC_KEY_PATH is not set")
	}

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
