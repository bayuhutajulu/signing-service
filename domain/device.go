package domain

import signingcrypto "github.com/bayuhutajulu/signing-service/crypto"

type SignatureDevice struct {
	ID               string
	Label            string
	Algorithm        string
	SignatureCounter int
	LastSignature    string
	PublicKey        interface{}
	PrivateKey       interface{}
	Signer           signingcrypto.Signer
}

type SignatureDeviceResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}
