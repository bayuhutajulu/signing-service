package domain

import (
	"encoding/base64"
	"fmt"
	"sync"

	signingcrypto "github.com/bayuhutajulu/signing-service/crypto"
)

type SignatureDeviceService struct {
	storage DeviceStorage
	mu      sync.Mutex
}

type CreateDeviceOptions struct {
	ID        string
	Label     string
	Algorithm string
}

func NewSignatureDeviceService(storage DeviceStorage) *SignatureDeviceService {
	return &SignatureDeviceService{
		storage: storage,
	}
}

func (s *SignatureDeviceService) CreateDevice(opts CreateDeviceOptions) (*SignatureDevice, error) {
	if opts.Algorithm != "RSA" && opts.Algorithm != "ECC" {
		return nil, fmt.Errorf("invalid algorithm: %s", opts.Algorithm)
	}

	var signer signingcrypto.Signer
	var privateKey, publicKey interface{}

	switch opts.Algorithm {
	case "RSA":
		generator := &signingcrypto.RSAGenerator{}
		keyPair, err := generator.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		privateKey = keyPair.Private
		publicKey = keyPair.Public
		signer = signingcrypto.NewRSASigner(keyPair.Private)
	case "ECC":
		generator := &signingcrypto.ECCGenerator{}
		keyPair, err := generator.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECC key pair: %w", err)
		}
		privateKey = keyPair.Private
		publicKey = keyPair.Public
		signer = signingcrypto.NewECDSASigner(keyPair.Private)
	}

	initialSignature := base64.StdEncoding.EncodeToString([]byte(opts.ID))
	device := &SignatureDevice{
		ID:               opts.ID,
		Label:            opts.Label,
		Algorithm:        opts.Algorithm,
		SignatureCounter: 0,
		LastSignature:    initialSignature,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		Signer:           signer,
	}

	err := s.storage.Save(device)
	if err != nil {
		return nil, fmt.Errorf("failed to save device: %w", err)
	}

	return device, nil
}

func (s *SignatureDeviceService) SignData(deviceID string, data string) (*SignatureDeviceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, err := s.storage.GetDevice(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find device: %w", err)
	}

	device.SignatureCounter++
	counter := device.SignatureCounter

	dataToBeSigned := fmt.Sprintf("%d_%s_%s", counter, data, device.LastSignature)
	signature, err := device.Signer.Sign([]byte(dataToBeSigned))
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	signatureB64 := base64.StdEncoding.EncodeToString(signature)
	device.LastSignature = signatureB64

	err = s.storage.Update(device)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	resp := &SignatureDeviceResponse{
		Signature:  signatureB64,
		SignedData: dataToBeSigned,
	}
	return resp, nil
}

func (s *SignatureDeviceService) GetDevice(id string) (*SignatureDevice, error) {
	device, err := s.storage.GetDevice(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return device, nil
}

func (s *SignatureDeviceService) GetAllDevices() ([]*SignatureDevice, error) {
	devices, err := s.storage.GetAllDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get all devices: %w", err)
	}
	return devices, nil
}
