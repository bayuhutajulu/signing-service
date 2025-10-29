package persistence

import (
	"fmt"
	"sync"

	"github.com/bayuhutajulu/signing-service/domain"
)

type InMemoryStorage struct {
	mu      sync.RWMutex
	devices map[string]*domain.SignatureDevice
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		devices: make(map[string]*domain.SignatureDevice),
	}
}

var _ domain.DeviceStorage = (*InMemoryStorage)(nil)

func (s *InMemoryStorage) Save(device *domain.SignatureDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
	return nil
}

func (s *InMemoryStorage) Update(device *domain.SignatureDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
	return nil
}

func (s *InMemoryStorage) GetDevice(id string) (*domain.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	device, exists := s.devices[id]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	return device, nil
}

func (s *InMemoryStorage) GetAllDevices() ([]*domain.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	devices := make([]*domain.SignatureDevice, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices, nil
}
