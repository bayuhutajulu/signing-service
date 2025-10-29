package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bayuhutajulu/signing-service/model"
	"github.com/gorilla/mux"
)

// CreateDevice handles POST /api/v0/devices to create a new signature device.
// Validates the request, creates the device with key pair generation, and returns
// device info (hiding private keys). Returns 409 if device ID already exists.
func (s *Server) CreateDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	var req model.CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{
			"Invalid request body",
		})
		return
	}

	device, err := s.signDeviceService.CreateDevice(req.ToOptions())
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			WriteErrorResponse(w, http.StatusConflict, []string{err.Error()})
		} else {
			WriteErrorResponse(w, http.StatusInternalServerError, []string{err.Error()})
		}
		return
	}

	response := model.DeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        device.Algorithm,
		SignatureCounter: device.SignatureCounter,
	}
	WriteAPIResponse(w, http.StatusCreated, response)
}

// SignData handles POST /api/v0/devices/{id}/sign to create a signature with chaining.
// Extracts device ID from URL path, signs the data using signature chaining format,
// and returns the signature with signed data string.
func (s *Server) SignData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	var req model.SignDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{
			"Invalid request body",
		})
		return
	}

	opt := req.ToOptions()
	opt.DeviceID = mux.Vars(r)["id"]
	resp, err := s.signDeviceService.SignData(opt)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{
			"Failed to sign data",
		})
		return
	}

	WriteAPIResponse(w, http.StatusOK, resp)
}

// GetDevice handles GET /api/v0/devices/{id} to retrieve a single device by ID.
// Returns device info (without private keys). Returns 500 if device not found.
func (s *Server) GetDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	deviceID := mux.Vars(r)["id"]
	if deviceID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{
			"Device ID is required",
		})
		return
	}

	device, err := s.signDeviceService.GetDevice(deviceID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{
			"Failed to get device",
		})
		return
	}

	response := model.DeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        device.Algorithm,
		SignatureCounter: device.SignatureCounter,
	}
	WriteAPIResponse(w, http.StatusOK, response)
}

// GetAllDevices handles GET /api/v0/devices to list all signature devices.
// Returns array of device info (without private keys). Returns empty array if no devices exist.
func (s *Server) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	devices, err := s.signDeviceService.GetAllDevices()
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{
			"Failed to get all devices",
		})
		return
	}

	responses := make([]model.DeviceResponse, len(devices))
	for i, device := range devices {
		responses[i] = model.DeviceResponse{
			ID:               device.ID,
			Label:            device.Label,
			Algorithm:        device.Algorithm,
			SignatureCounter: device.SignatureCounter,
		}
	}
	WriteAPIResponse(w, http.StatusOK, responses)
}
