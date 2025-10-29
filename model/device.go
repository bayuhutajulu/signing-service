package model

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

type CreateDeviceOptions struct {
	ID        string
	Label     string
	Algorithm string
}

type CreateDeviceRequest struct {
	ID        string
	Label     string
	Algorithm string
}

func (r *CreateDeviceRequest) ToOptions() CreateDeviceOptions {
	return CreateDeviceOptions{
		ID:        r.ID,
		Label:     r.Label,
		Algorithm: r.Algorithm,
	}
}

type DeviceResponse struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	Algorithm        string `json:"algorithm"`
	SignatureCounter int    `json:"signature_counter"`
}
