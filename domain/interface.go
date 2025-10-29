package domain

import model "github.com/bayuhutajulu/signing-service/model"

type ISignatureDeviceService interface {
	CreateDevice(opts model.CreateDeviceOptions) (*model.SignatureDevice, error)
	SignData(opts model.SignDataOptions) (*model.SignDataResponse, error)
	GetDevice(id string) (*model.SignatureDevice, error)
	GetAllDevices() ([]*model.SignatureDevice, error)
}
