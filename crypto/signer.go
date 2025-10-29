package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// Signer defines a contract for cryptographic signing operations.
// New algorithms can be added by implementing this interface.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

// RSASigner implements signing using RSA with PKCS#1 v1.5 and SHA-256.
type RSASigner struct {
	privateKey *rsa.PrivateKey
}

// NewRSASigner creates an RSA signer with the provided private key.
func NewRSASigner(privateKey *rsa.PrivateKey) *RSASigner {
	return &RSASigner{
		privateKey: privateKey,
	}
}

// Sign generates an RSA signature by hashing data with SHA-256 then signing with PKCS#1v15.
// Returns raw signature bytes. The hash[:] slice conversion is required by the signing API.
func (s *RSASigner) Sign(dataTobeSigned []byte) ([]byte, error) {
	hash := sha256.Sum256(dataTobeSigned)
	return rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
}

// ECDSASigner implements signing using ECDSA with SHA-256 and ASN.1 encoding.
type ECDSASigner struct {
	privateKey *ecdsa.PrivateKey
}

// NewECDSASigner creates an ECDSA signer with the provided private key.
func NewECDSASigner(privateKey *ecdsa.PrivateKey) *ECDSASigner {
	return &ECDSASigner{
		privateKey: privateKey,
	}
}

// Sign generates an ECDSA signature by hashing data with SHA-256 then signing with ASN.1 encoding.
// Returns ASN.1 DER encoded signature bytes. Unlike RSA, ECDSA includes randomness per signature.
func (s *ECDSASigner) Sign(dataTobeSigned []byte) ([]byte, error) {
	hash := sha256.Sum256(dataTobeSigned)
	return ecdsa.SignASN1(rand.Reader, s.privateKey, hash[:])
}
