package domain

type DeviceStorage interface {
	Save(device *SignatureDevice) error
	Update(device *SignatureDevice) error
	GetDevice(id string) (*SignatureDevice, error)
	GetAllDevices() ([]*SignatureDevice, error)
}
