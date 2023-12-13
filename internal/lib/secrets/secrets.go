package secrets

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	keyPath := os.Getenv("PRIVATE_KEY_PATH")
	if keyPath == "" {
		return nil, fmt.Errorf("PRIVATE_KEY_PATH is not set")
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read private key file")
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil || keyBlock.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("cannot decode PEM block containing private key")
	}

	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse EC private key")
	}

	return key, nil
}

func GetPublicKey() (*ecdsa.PublicKey, error) {
	keyPath := os.Getenv("PUBLIC_KEY_PATH")
	if keyPath == "" {
		return nil, fmt.Errorf("PUBLIC_KEY_PATH is not set")
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key file")
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil || keyBlock.Type != "EC PUBLIC KEY" {
		return nil, fmt.Errorf("cannot decode PEM block containing public key")
	}

	keyUntyped, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse EC public key")
	}

	key, ok := keyUntyped.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is of the wrong type")
	}

	return key, nil
}
