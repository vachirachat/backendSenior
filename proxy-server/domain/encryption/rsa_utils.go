package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
)

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	if pub == nil {
		return nil, errors.New("encrypt error: public key is nil")
	}
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	if priv == nil {
		return nil, errors.New("decrypt error: private key is nil")
	}
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)

	if err != nil {
		return nil, err
	}
	return plaintext, err
}

// PrivateKeyToBytes private key to bytes
// it will begin with --- BEGIN RSA PRIVATE KEY ---
func PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return privBytes
}

// PublicKeyToBytes public key to bytes
// it will begin with --- BEGIN RSA PUBLIC KEY ---
func PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Println("public key to bytes error:", err)
		return nil
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

// BytesToPrivateKey bytes to private key
// keys need to begin with --- BEGIN RSA PRIVATE KEY ---
func BytesToPrivateKey(priv []byte) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode(priv)
	if block == nil {
		return nil, fmt.Errorf("bytes to private key: error decoding: %s | %s", block, rest)
	}

	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes

	var err error
	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, fmt.Errorf("x509/decrypt pem error: %v", err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("x509/prase pkcs1 privkey error: %v", err)
	}
	return key, nil
}

// BytesToPublicKey bytes to public key
func BytesToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, rest := pem.Decode(pub)
	if block == nil {
		return nil, fmt.Errorf("bytes to private key: error decoding: %s | %s", block, rest)
	}

	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, fmt.Errorf("x509/decrypt pem error: %v", err)
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, fmt.Errorf("x509/prase pkcs1 pubkey error: %v", err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("type assertion: key isn't RSA public key")
	}
	return key, nil
}
