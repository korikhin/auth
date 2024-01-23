package secrets

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Expected key paths
const (
	privateKeyPath = "secrets/.PRIVATE.pem"
	publicKeyPath  = "secrets/.PUBLIC.pem"
)

// PEM block types
const (
	keyTypePrivate = "EC PRIVATE KEY"
	keyTypePublic  = "PUBLIC KEY"
)

func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(privateKeyPath)
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
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key file")
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != keyTypePublic {
		return nil, fmt.Errorf("cannot decode PEM block containing public key")
	}

	ukey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key")
	}

	key, ok := ukey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is of the wrong type")
	}

	return key, nil
}
