package domain

import (
	"encoding/base64"
	"fmt"
	"sync"

	signingcrypto "github.com/bayuhutajulu/signing-service/crypto"
	model "github.com/bayuhutajulu/signing-service/model"
)

// SignatureDeviceService orchestrates device creation, signature generation with chaining,
// and device retrieval. Uses a mutex to ensure atomic counter increments across concurrent requests.
type SignatureDeviceService struct {
	storage DeviceStorage
	mu      sync.Mutex // Serializes signing operations to prevent counter gaps
}

// NewSignatureDeviceService creates a service with the given storage implementation.
func NewSignatureDeviceService(storage DeviceStorage) *SignatureDeviceService {
	return &SignatureDeviceService{
		storage: storage,
	}
}

// CreateDevice generates a new signature device with a cryptographic key pair.
// Validates algorithm (RSA/ECC), generates keys, initializes counter to 0, and sets
// last_signature to base64(device_id) for the base case. Persists device to storage.
func (s *SignatureDeviceService) CreateDevice(opts model.CreateDeviceOptions) (*model.SignatureDevice, error) {
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
	device := &model.SignatureDevice{
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

// SignData generates a signature with chaining using format: "<counter>_<data>_<last_signature>".
// Uses the CURRENT counter value (starting from 0), signs the data, then increments counter.
// The mutex ensures strictly monotonic counter increments without gaps during concurrent access.
func (s *SignatureDeviceService) SignData(opts model.SignDataOptions) (*model.SignDataResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, err := s.storage.GetDevice(opts.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find device: %w", err)
	}

	counter := device.SignatureCounter
	dataToBeSigned := fmt.Sprintf("%d_%s_%s", counter, opts.Data, device.LastSignature)
	signature, err := device.Signer.Sign([]byte(dataToBeSigned))
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	device.SignatureCounter++

	signatureB64 := base64.StdEncoding.EncodeToString(signature)
	device.LastSignature = signatureB64

	err = s.storage.Update(device)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	resp := &model.SignDataResponse{
		Signature:  signatureB64,
		SignedData: dataToBeSigned,
	}
	return resp, nil
}

// GetDevice retrieves a device by its unique identifier.
func (s *SignatureDeviceService) GetDevice(id string) (*model.SignatureDevice, error) {
	device, err := s.storage.GetDevice(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return device, nil
}

// GetAllDevices retrieves all devices from storage.
func (s *SignatureDeviceService) GetAllDevices() ([]*model.SignatureDevice, error) {
	devices, err := s.storage.GetAllDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get all devices: %w", err)
	}
	return devices, nil
}
