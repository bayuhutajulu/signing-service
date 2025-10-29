package persistence

import (
	"fmt"
	"sync"

	"github.com/bayuhutajulu/signing-service/domain"
	model "github.com/bayuhutajulu/signing-service/model"
)

// InMemoryStorage provides thread-safe in-memory storage for signature devices.
// Uses RWMutex to allow concurrent reads while ensuring exclusive writes.
type InMemoryStorage struct {
	mu      sync.RWMutex
	devices map[string]*model.SignatureDevice
}

// NewInMemoryStorage creates an empty in-memory storage instance.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		devices: make(map[string]*model.SignatureDevice),
	}
}

// Compile-time check that InMemoryStorage implements DeviceStorage interface.
var _ domain.DeviceStorage = (*InMemoryStorage)(nil)

// Save persists a new device to storage. Returns an error if device ID already exists.
func (s *InMemoryStorage) Save(device *model.SignatureDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[device.ID]; exists {
		return fmt.Errorf("device %s already exists", device.ID)
	}

	s.devices[device.ID] = device
	return nil
}

// Update overwrites an existing device in storage. Creates device if it doesn't exist.
func (s *InMemoryStorage) Update(device *model.SignatureDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
	return nil
}

// GetDevice retrieves a device by ID. Returns error if device not found.
func (s *InMemoryStorage) GetDevice(id string) (*model.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	device, exists := s.devices[id]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	return device, nil
}

// GetAllDevices returns all devices in storage. Returns empty slice if no devices exist.
func (s *InMemoryStorage) GetAllDevices() ([]*model.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	devices := make([]*model.SignatureDevice, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices, nil
}
