package domain

import (
	"fmt"
	"sync"
	"testing"
)

type mockStorage struct {
	mu      sync.RWMutex
	devices map[string]*SignatureDevice
	saveErr error
	updateErr error
	getErr error
	getAllErr error
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		devices: make(map[string]*SignatureDevice),
	}
}

func (m *mockStorage) Save(device *SignatureDevice) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[device.ID] = device
	return nil
}

func (m *mockStorage) Update(device *SignatureDevice) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.devices[device.ID]; !exists {
		return fmt.Errorf("device not found")
	}
	m.devices[device.ID] = device
	return nil
}

func (m *mockStorage) GetDevice(id string) (*SignatureDevice, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	device, exists := m.devices[id]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	return device, nil
}

func (m *mockStorage) GetAllDevices() ([]*SignatureDevice, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	devices := make([]*SignatureDevice, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

func TestCreateDevice(t *testing.T) {
	t.Run("successful RSA device creation", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "device-rsa-001",
			Label:     "RSA Test Device",
			Algorithm: "RSA",
		}

		device, err := service.CreateDevice(opts)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if device == nil {
			t.Fatal("expected device, got nil")
		}
		if device.ID != opts.ID {
			t.Errorf("expected ID %s, got %s", opts.ID, device.ID)
		}
		if device.Label != opts.Label {
			t.Errorf("expected label %s, got %s", opts.Label, device.Label)
		}
		if device.Algorithm != "RSA" {
			t.Errorf("expected algorithm RSA, got %s", device.Algorithm)
		}
		if device.SignatureCounter != 0 {
			t.Errorf("expected counter 0, got %d", device.SignatureCounter)
		}
		if device.Signer == nil {
			t.Error("expected signer to be initialized")
		}
		if device.PrivateKey == nil {
			t.Error("expected private key to be initialized")
		}
		if device.PublicKey == nil {
			t.Error("expected public key to be initialized")
		}
		if device.LastSignature == "" {
			t.Error("expected last signature to be initialized")
		}
	})

	t.Run("successful ECC device creation", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "device-ecc-001",
			Label:     "ECC Test Device",
			Algorithm: "ECC",
		}

		device, err := service.CreateDevice(opts)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if device == nil {
			t.Fatal("expected device, got nil")
		}
		if device.Algorithm != "ECC" {
			t.Errorf("expected algorithm ECC, got %s", device.Algorithm)
		}
		if device.Signer == nil {
			t.Error("expected signer to be initialized")
		}
	})

	t.Run("invalid algorithm", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "device-invalid-001",
			Label:     "Invalid Device",
			Algorithm: "INVALID",
		}

		device, err := service.CreateDevice(opts)

		if err == nil {
			t.Fatal("expected error for invalid algorithm, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device, got %v", device)
		}
	})

	t.Run("empty algorithm", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "device-empty-001",
			Label:     "Empty Algorithm Device",
			Algorithm: "",
		}

		device, err := service.CreateDevice(opts)

		if err == nil {
			t.Fatal("expected error for empty algorithm, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device, got %v", device)
		}
	})

	t.Run("storage save error", func(t *testing.T) {
		storage := newMockStorage()
		storage.saveErr = fmt.Errorf("storage error")
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "device-error-001",
			Label:     "Error Device",
			Algorithm: "RSA",
		}

		device, err := service.CreateDevice(opts)

		if err == nil {
			t.Fatal("expected error from storage, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device on error, got %v", device)
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "",
			Label:     "Empty ID Device",
			Algorithm: "RSA",
		}

		device, err := service.CreateDevice(opts)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if device == nil {
			t.Fatal("expected device, got nil")
		}
		if device.ID != "" {
			t.Errorf("expected empty ID to be preserved, got %s", device.ID)
		}
	})

	t.Run("empty label", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		opts := CreateDeviceOptions{
			ID:        "device-empty-label-001",
			Label:     "",
			Algorithm: "ECC",
		}

		device, err := service.CreateDevice(opts)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if device == nil {
			t.Fatal("expected device, got nil")
		}
		if device.Label != "" {
			t.Errorf("expected empty label to be preserved, got %s", device.Label)
		}
	})
}

func TestSignData(t *testing.T) {
	t.Run("successful signature creation", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		device, _ := service.CreateDevice(CreateDeviceOptions{
			ID:        "device-sign-001",
			Label:     "Sign Test",
			Algorithm: "RSA",
		})

		resp, err := service.SignData(device.ID, "test-data")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("expected response, got nil")
		}
		if resp.Signature == "" {
			t.Error("expected signature to be set")
		}
		if resp.SignedData == "" {
			t.Error("expected signed data to be set")
		}

		updatedDevice, _ := storage.GetDevice(device.ID)
		if updatedDevice.SignatureCounter != 1 {
			t.Errorf("expected counter 1, got %d", updatedDevice.SignatureCounter)
		}
		if updatedDevice.LastSignature != resp.Signature {
			t.Error("expected last signature to be updated")
		}
	})

	t.Run("signature counter increments correctly", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		device, _ := service.CreateDevice(CreateDeviceOptions{
			ID:        "device-counter-001",
			Label:     "Counter Test",
			Algorithm: "RSA",
		})

		for i := 1; i <= 5; i++ {
			resp, err := service.SignData(device.ID, fmt.Sprintf("data-%d", i))
			if err != nil {
				t.Fatalf("iteration %d: expected no error, got %v", i, err)
			}

			updatedDevice, _ := storage.GetDevice(device.ID)
			if updatedDevice.SignatureCounter != i {
				t.Errorf("iteration %d: expected counter %d, got %d", i, i, updatedDevice.SignatureCounter)
			}
			if updatedDevice.LastSignature != resp.Signature {
				t.Errorf("iteration %d: last signature not updated correctly", i)
			}
		}
	})

	t.Run("device not found", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		resp, err := service.SignData("non-existent-device", "test-data")

		if err == nil {
			t.Fatal("expected error for non-existent device, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response, got %v", resp)
		}
	})

	t.Run("storage get error", func(t *testing.T) {
		storage := newMockStorage()
		storage.getErr = fmt.Errorf("storage get error")
		service := NewSignatureDeviceService(storage)

		resp, err := service.SignData("device-001", "test-data")

		if err == nil {
			t.Fatal("expected error from storage, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response, got %v", resp)
		}
	})

	t.Run("storage update error", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		device, _ := service.CreateDevice(CreateDeviceOptions{
			ID:        "device-update-error-001",
			Label:     "Update Error Test",
			Algorithm: "RSA",
		})

		storage.updateErr = fmt.Errorf("update error")

		resp, err := service.SignData(device.ID, "test-data")

		if err == nil {
			t.Fatal("expected error from storage update, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response on update error, got %v", resp)
		}
	})

	t.Run("signature format verification", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		device, _ := service.CreateDevice(CreateDeviceOptions{
			ID:        "device-format-001",
			Label:     "Format Test",
			Algorithm: "RSA",
		})

		data := "transaction-data"
		resp, err := service.SignData(device.ID, data)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedPrefix := "1_" + data + "_"
		if len(resp.SignedData) < len(expectedPrefix) {
			t.Error("signed data format incorrect")
		}
		if resp.SignedData[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected signed data to start with %s", expectedPrefix)
		}
	})
}

func TestGetDevice(t *testing.T) {
	t.Run("successful device retrieval", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		created, _ := service.CreateDevice(CreateDeviceOptions{
			ID:        "device-get-001",
			Label:     "Get Test",
			Algorithm: "RSA",
		})

		retrieved, err := service.GetDevice(created.ID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected device, got nil")
		}
		if retrieved.ID != created.ID {
			t.Errorf("expected ID %s, got %s", created.ID, retrieved.ID)
		}
	})

	t.Run("device not found", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		device, err := service.GetDevice("non-existent-device")

		if err == nil {
			t.Fatal("expected error for non-existent device, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device, got %v", device)
		}
	})

	t.Run("storage error", func(t *testing.T) {
		storage := newMockStorage()
		storage.getErr = fmt.Errorf("storage error")
		service := NewSignatureDeviceService(storage)

		device, err := service.GetDevice("device-001")

		if err == nil {
			t.Fatal("expected error from storage, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device, got %v", device)
		}
	})
}

func TestGetAllDevices(t *testing.T) {
	t.Run("successful retrieval of multiple devices", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		service.CreateDevice(CreateDeviceOptions{
			ID:        "device-all-001",
			Label:     "Device 1",
			Algorithm: "RSA",
		})
		service.CreateDevice(CreateDeviceOptions{
			ID:        "device-all-002",
			Label:     "Device 2",
			Algorithm: "ECC",
		})

		devices, err := service.GetAllDevices()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(devices) != 2 {
			t.Errorf("expected 2 devices, got %d", len(devices))
		}
	})

	t.Run("empty device list", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		devices, err := service.GetAllDevices()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(devices))
		}
	})

	t.Run("storage error", func(t *testing.T) {
		storage := newMockStorage()
		storage.getAllErr = fmt.Errorf("storage error")
		service := NewSignatureDeviceService(storage)

		devices, err := service.GetAllDevices()

		if err == nil {
			t.Fatal("expected error from storage, got nil")
		}
		if devices != nil {
			t.Errorf("expected nil devices, got %v", devices)
		}
	})
}

func TestConcurrentSignData(t *testing.T) {
	t.Run("concurrent signing maintains counter integrity", func(t *testing.T) {
		storage := newMockStorage()
		service := NewSignatureDeviceService(storage)

		device, _ := service.CreateDevice(CreateDeviceOptions{
			ID:        "device-concurrent-001",
			Label:     "Concurrent Test",
			Algorithm: "RSA",
		})

		concurrency := 100
		var wg sync.WaitGroup
		errorsChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				_, err := service.SignData(device.ID, fmt.Sprintf("data-%d", index))
				if err != nil {
					errorsChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errorsChan)

		for err := range errorsChan {
			t.Errorf("unexpected error: %v", err)
		}

		finalDevice, _ := storage.GetDevice(device.ID)
		if finalDevice.SignatureCounter != concurrency {
			t.Errorf("expected final counter %d, got %d", concurrency, finalDevice.SignatureCounter)
		}
	})
}
