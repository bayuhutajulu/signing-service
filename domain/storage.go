package domain

import model "github.com/bayuhutajulu/signing-service/model"

type DeviceStorage interface {
	Save(device *model.SignatureDevice) error
	Update(device *model.SignatureDevice) error
	GetDevice(id string) (*model.SignatureDevice, error)
	GetAllDevices() ([]*model.SignatureDevice, error)
}
