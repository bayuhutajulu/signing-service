package model

type SignDataOptions struct {
	DeviceID string
	Data     string
}

type SignDataRequest struct {
	Data string
}

func (r *SignDataRequest) ToOptions() SignDataOptions {
	return SignDataOptions{
		Data: r.Data,
	}
}

type SignDataResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}
