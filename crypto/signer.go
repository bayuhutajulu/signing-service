package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

type RSASigner struct {
	privateKey *rsa.PrivateKey
}

func NewRSASigner(privateKey *rsa.PrivateKey) *RSASigner {
	return &RSASigner{
		privateKey: privateKey,
	}
}

func (s *RSASigner) Sign(dataTobeSigned []byte) ([]byte, error) {
	hash := sha256.Sum256(dataTobeSigned)
	return rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
}

type ECDSASigner struct {
	privateKey *ecdsa.PrivateKey
}

func NewECDSASigner(privateKey *ecdsa.PrivateKey) *ECDSASigner {
	return &ECDSASigner{
		privateKey: privateKey,
	}
}

func (s *ECDSASigner) Sign(dataTobeSigned []byte) ([]byte, error) {
	hash := sha256.Sum256(dataTobeSigned)
	return ecdsa.SignASN1(rand.Reader, s.privateKey, hash[:])
}
