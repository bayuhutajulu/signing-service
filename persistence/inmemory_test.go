package persistence

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bayuhutajulu/signing-service/crypto"
	model "github.com/bayuhutajulu/signing-service/model"
)

func createTestDevice(id, label, algorithm string) *model.SignatureDevice {
	var signer crypto.Signer
	var privateKey, publicKey interface{}

	if algorithm == "RSA" {
		generator := &crypto.RSAGenerator{}
		keyPair, _ := generator.Generate()
		signer = crypto.NewRSASigner(keyPair.Private)
		privateKey = keyPair.Private
		publicKey = keyPair.Public
	} else {
		generator := &crypto.ECCGenerator{}
		keyPair, _ := generator.Generate()
		signer = crypto.NewECDSASigner(keyPair.Private)
		privateKey = keyPair.Private
		publicKey = keyPair.Public
	}

	return &model.SignatureDevice{
		ID:               id,
		Label:            label,
		Algorithm:        algorithm,
		SignatureCounter: 0,
		LastSignature:    "initial",
		PrivateKey:       privateKey,
		PublicKey:        publicKey,
		Signer:           signer,
	}
}

func TestNewInMemoryStorage(t *testing.T) {
	t.Run("creates storage with initialized map", func(t *testing.T) {
		storage := NewInMemoryStorage()

		if storage == nil {
			t.Fatal("expected storage to be initialized")
		}
		if storage.devices == nil {
			t.Fatal("expected devices map to be initialized")
		}
		if len(storage.devices) != 0 {
			t.Errorf("expected empty map, got %d devices", len(storage.devices))
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("successfully saves device", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-001", "Test Device", "RSA")

		err := storage.Save(device)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(storage.devices) != 1 {
			t.Errorf("expected 1 device in storage, got %d", len(storage.devices))
		}

		saved := storage.devices[device.ID]
		if saved == nil {
			t.Fatal("expected device to be in storage")
		}
		if saved.ID != device.ID {
			t.Errorf("expected ID %s, got %s", device.ID, saved.ID)
		}
	})

	t.Run("overwrites existing device with same ID", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device1 := createTestDevice("device-002", "Label 1", "RSA")
		device2 := createTestDevice("device-002", "Label 2", "ECC")

		storage.Save(device1)
		storage.Save(device2)

		if len(storage.devices) != 1 {
			t.Errorf("expected 1 device in storage, got %d", len(storage.devices))
		}

		saved := storage.devices["device-002"]
		if saved.Label != "Label 2" {
			t.Errorf("expected label 'Label 2', got '%s'", saved.Label)
		}
		if saved.Algorithm != "ECC" {
			t.Errorf("expected algorithm 'ECC', got '%s'", saved.Algorithm)
		}
	})

	t.Run("saves multiple different devices", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device1 := createTestDevice("device-003", "Device 1", "RSA")
		device2 := createTestDevice("device-004", "Device 2", "ECC")
		device3 := createTestDevice("device-005", "Device 3", "RSA")

		storage.Save(device1)
		storage.Save(device2)
		storage.Save(device3)

		if len(storage.devices) != 3 {
			t.Errorf("expected 3 devices in storage, got %d", len(storage.devices))
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("successfully updates existing device", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-006", "Original Label", "RSA")

		storage.Save(device)

		device.Label = "Updated Label"
		device.SignatureCounter = 5

		err := storage.Update(device)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updated := storage.devices[device.ID]
		if updated.Label != "Updated Label" {
			t.Errorf("expected label 'Updated Label', got '%s'", updated.Label)
		}
		if updated.SignatureCounter != 5 {
			t.Errorf("expected counter 5, got %d", updated.SignatureCounter)
		}
	})

	t.Run("creates device if not exists", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-007", "New Device", "ECC")

		err := storage.Update(device)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(storage.devices) != 1 {
			t.Errorf("expected 1 device in storage, got %d", len(storage.devices))
		}
	})

	t.Run("updates device counter", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-008", "Counter Test", "RSA")
		device.SignatureCounter = 0

		storage.Save(device)

		for i := 1; i <= 10; i++ {
			device.SignatureCounter = i
			err := storage.Update(device)
			if err != nil {
				t.Fatalf("iteration %d: expected no error, got %v", i, err)
			}

			updated := storage.devices[device.ID]
			if updated.SignatureCounter != i {
				t.Errorf("iteration %d: expected counter %d, got %d", i, i, updated.SignatureCounter)
			}
		}
	})
}

func TestGetDevice(t *testing.T) {
	t.Run("successfully retrieves existing device", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-009", "Test Device", "RSA")

		storage.Save(device)

		retrieved, err := storage.GetDevice(device.ID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected device, got nil")
		}
		if retrieved.ID != device.ID {
			t.Errorf("expected ID %s, got %s", device.ID, retrieved.ID)
		}
		if retrieved.Label != device.Label {
			t.Errorf("expected label %s, got %s", device.Label, retrieved.Label)
		}
	})

	t.Run("returns error for non-existent device", func(t *testing.T) {
		storage := NewInMemoryStorage()

		device, err := storage.GetDevice("non-existent-id")

		if err == nil {
			t.Fatal("expected error for non-existent device, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device, got %v", device)
		}
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		storage := NewInMemoryStorage()

		device, err := storage.GetDevice("")

		if err == nil {
			t.Fatal("expected error for empty ID, got nil")
		}
		if device != nil {
			t.Errorf("expected nil device, got %v", device)
		}
	})
}

func TestGetAllDevices(t *testing.T) {
	t.Run("returns all devices", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device1 := createTestDevice("device-010", "Device 1", "RSA")
		device2 := createTestDevice("device-011", "Device 2", "ECC")
		device3 := createTestDevice("device-012", "Device 3", "RSA")

		storage.Save(device1)
		storage.Save(device2)
		storage.Save(device3)

		devices, err := storage.GetAllDevices()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(devices) != 3 {
			t.Errorf("expected 3 devices, got %d", len(devices))
		}

		ids := make(map[string]bool)
		for _, device := range devices {
			ids[device.ID] = true
		}

		if !ids["device-010"] || !ids["device-011"] || !ids["device-012"] {
			t.Error("expected all three device IDs to be present")
		}
	})

	t.Run("returns empty slice for no devices", func(t *testing.T) {
		storage := NewInMemoryStorage()

		devices, err := storage.GetAllDevices()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if devices == nil {
			t.Fatal("expected empty slice, got nil")
		}
		if len(devices) != 0 {
			t.Errorf("expected 0 devices, got %d", len(devices))
		}
	})

	t.Run("returns independent slice", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-013", "Test Device", "RSA")
		storage.Save(device)

		devices1, _ := storage.GetAllDevices()
		devices2, _ := storage.GetAllDevices()

		if len(devices1) != len(devices2) {
			t.Error("expected same length for both calls")
		}

		devices1 = append(devices1, createTestDevice("extra", "Extra", "RSA"))

		if len(devices1) == len(devices2) {
			t.Error("expected devices1 to be independent from devices2")
		}
	})
}

func TestConcurrentOperations(t *testing.T) {
	t.Run("concurrent saves", func(t *testing.T) {
		storage := NewInMemoryStorage()
		concurrency := 100
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				device := createTestDevice(
					fmt.Sprintf("device-concurrent-save-%d", index),
					fmt.Sprintf("Device %d", index),
					"RSA",
				)
				storage.Save(device)
			}(i)
		}

		wg.Wait()

		devices, _ := storage.GetAllDevices()
		if len(devices) != concurrency {
			t.Errorf("expected %d devices, got %d", concurrency, len(devices))
		}
	})

	t.Run("concurrent updates", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-concurrent-update", "Test", "RSA")
		storage.Save(device)

		concurrency := 100
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				device := createTestDevice("device-concurrent-update", fmt.Sprintf("Label %d", index), "RSA")
				device.SignatureCounter = index
				storage.Update(device)
			}(i)
		}

		wg.Wait()

		updated, _ := storage.GetDevice("device-concurrent-update")
		if updated == nil {
			t.Fatal("expected device to exist")
		}
	})

	t.Run("concurrent reads", func(t *testing.T) {
		storage := NewInMemoryStorage()
		device := createTestDevice("device-concurrent-read", "Test", "RSA")
		storage.Save(device)

		concurrency := 100
		var wg sync.WaitGroup
		errorsChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := storage.GetDevice("device-concurrent-read")
				if err != nil {
					errorsChan <- err
				}
			}()
		}

		wg.Wait()
		close(errorsChan)

		for err := range errorsChan {
			t.Errorf("unexpected error during concurrent read: %v", err)
		}
	})

	t.Run("concurrent read all", func(t *testing.T) {
		storage := NewInMemoryStorage()
		for i := 0; i < 10; i++ {
			device := createTestDevice(fmt.Sprintf("device-read-all-%d", i), "Test", "RSA")
			storage.Save(device)
		}

		concurrency := 50
		var wg sync.WaitGroup
		errorsChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				devices, err := storage.GetAllDevices()
				if err != nil {
					errorsChan <- err
					return
				}
				if len(devices) != 10 {
					errorsChan <- fmt.Errorf("expected 10 devices, got %d", len(devices))
				}
			}()
		}

		wg.Wait()
		close(errorsChan)

		for err := range errorsChan {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("concurrent mixed operations", func(t *testing.T) {
		storage := NewInMemoryStorage()
		concurrency := 100
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				if index%3 == 0 {
					device := createTestDevice(fmt.Sprintf("device-mixed-%d", index), "Test", "RSA")
					storage.Save(device)
				} else if index%3 == 1 {
					storage.GetDevice(fmt.Sprintf("device-mixed-%d", index-1))
				} else {
					storage.GetAllDevices()
				}
			}(i)
		}

		wg.Wait()
	})
}
