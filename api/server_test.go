package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/bayuhutajulu/signing-service/domain"
	"github.com/bayuhutajulu/signing-service/model"
	"github.com/bayuhutajulu/signing-service/persistence"
	"github.com/gorilla/mux"
)

func setupTestServer() (*Server, *domain.SignatureDeviceService) {
	storage := persistence.NewInMemoryStorage()
	service := domain.NewSignatureDeviceService(storage)
	server := NewServer(":8080", service)
	return server, service
}

func TestCreateDevice(t *testing.T) {
	t.Run("successful device creation with RSA", func(t *testing.T) {
		server, _ := setupTestServer()

		reqBody := model.CreateDeviceRequest{
			ID:        "device-001",
			Label:     "Test Device",
			Algorithm: "RSA",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response Response
		json.NewDecoder(w.Body).Decode(&response)

		if response.Data == nil {
			t.Error("expected response data, got nil")
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}
	})

	t.Run("successful device creation with ECC", func(t *testing.T) {
		server, _ := setupTestServer()

		reqBody := model.CreateDeviceRequest{
			ID:        "device-002",
			Label:     "ECC Device",
			Algorithm: "ECC",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("invalid algorithm", func(t *testing.T) {
		server, _ := setupTestServer()

		reqBody := model.CreateDeviceRequest{
			ID:        "device-003",
			Label:     "Invalid Device",
			Algorithm: "INVALID",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var response ErrorResponse
		json.NewDecoder(w.Body).Decode(&response)

		if len(response.Errors) == 0 {
			t.Error("expected error message")
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices", nil)
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestSignData(t *testing.T) {
	t.Run("successful signature creation", func(t *testing.T) {
		server, service := setupTestServer()

		device, _ := service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-sign-001",
			Label:     "Sign Test",
			Algorithm: "RSA",
		})

		reqBody := model.SignDataRequest{
			Data: "transaction-data",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/"+device.ID+"/sign", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": device.ID})
		w := httptest.NewRecorder()

		server.SignData(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response Response
		json.NewDecoder(w.Body).Decode(&response)

		if response.Data == nil {
			t.Error("expected signature response, got nil")
		}
	})

	t.Run("device not found", func(t *testing.T) {
		server, _ := setupTestServer()

		reqBody := model.SignDataRequest{
			Data: "transaction-data",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/non-existent/sign", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "non-existent"})
		w := httptest.NewRecorder()

		server.SignData(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/device-001/sign", bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"id": "device-001"})
		w := httptest.NewRecorder()

		server.SignData(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices/device-001/sign", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "device-001"})
		w := httptest.NewRecorder()

		server.SignData(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("signature counter increments correctly", func(t *testing.T) {
		server, service := setupTestServer()

		device, _ := service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-counter-test",
			Label:     "Counter Test",
			Algorithm: "RSA",
		})

		for i := 1; i <= 5; i++ {
			reqBody := model.SignDataRequest{
				Data: fmt.Sprintf("transaction-%d", i),
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/"+device.ID+"/sign", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": device.ID})
			w := httptest.NewRecorder()

			server.SignData(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("iteration %d: expected status %d, got %d", i, http.StatusOK, w.Code)
			}
		}

		updatedDevice, _ := service.GetDevice(device.ID)
		if updatedDevice.SignatureCounter != 5 {
			t.Errorf("expected counter 5, got %d", updatedDevice.SignatureCounter)
		}
	})
}

func TestGetDevice(t *testing.T) {
	t.Run("successful device retrieval", func(t *testing.T) {
		server, service := setupTestServer()

		device, _ := service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-get-001",
			Label:     "Get Test",
			Algorithm: "RSA",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices/"+device.ID, nil)
		req = mux.SetURLVars(req, map[string]string{"id": device.ID})
		w := httptest.NewRecorder()

		server.GetDevice(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response Response
		json.NewDecoder(w.Body).Decode(&response)

		if response.Data == nil {
			t.Error("expected device data, got nil")
		}
	})

	t.Run("device not found", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices/non-existent", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "non-existent"})
		w := httptest.NewRecorder()

		server.GetDevice(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices/", nil)
		req = mux.SetURLVars(req, map[string]string{"id": ""})
		w := httptest.NewRecorder()

		server.GetDevice(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/device-001", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "device-001"})
		w := httptest.NewRecorder()

		server.GetDevice(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestGetAllDevices(t *testing.T) {
	t.Run("returns all devices", func(t *testing.T) {
		server, service := setupTestServer()

		service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-all-001",
			Label:     "Device 1",
			Algorithm: "RSA",
		})
		service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-all-002",
			Label:     "Device 2",
			Algorithm: "ECC",
		})
		service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-all-003",
			Label:     "Device 3",
			Algorithm: "RSA",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices", nil)
		w := httptest.NewRecorder()

		server.GetAllDevices(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response Response
		json.NewDecoder(w.Body).Decode(&response)

		if response.Data == nil {
			t.Error("expected devices data, got nil")
		}
	})

	t.Run("returns empty list", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodGet, "/api/v0/devices", nil)
		w := httptest.NewRecorder()

		server.GetAllDevices(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodDelete, "/api/v0/devices", nil)
		w := httptest.NewRecorder()

		server.GetAllDevices(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestConcurrentAPIRequests(t *testing.T) {
	t.Run("concurrent device creation", func(t *testing.T) {
		server, service := setupTestServer()

		concurrency := 50
		var wg sync.WaitGroup
		errorsChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				reqBody := model.CreateDeviceRequest{
					ID:        fmt.Sprintf("device-concurrent-%d", index),
					Label:     fmt.Sprintf("Device %d", index),
					Algorithm: "RSA",
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(body))
				w := httptest.NewRecorder()

				server.CreateDevice(w, req)

				if w.Code != http.StatusCreated {
					errorsChan <- fmt.Errorf("request %d failed with status %d", index, w.Code)
				}
			}(i)
		}

		wg.Wait()
		close(errorsChan)

		for err := range errorsChan {
			t.Error(err)
		}

		devices, _ := service.GetAllDevices()
		if len(devices) != concurrency {
			t.Errorf("expected %d devices, got %d", concurrency, len(devices))
		}
	})

	t.Run("concurrent signing with same device", func(t *testing.T) {
		server, service := setupTestServer()

		device, _ := service.CreateDevice(model.CreateDeviceOptions{
			ID:        "device-concurrent-sign",
			Label:     "Concurrent Sign Test",
			Algorithm: "RSA",
		})

		concurrency := 100
		var wg sync.WaitGroup
		errorsChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				reqBody := model.SignDataRequest{
					Data: fmt.Sprintf("transaction-%d", index),
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/"+device.ID+"/sign", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"id": device.ID})
				w := httptest.NewRecorder()

				server.SignData(w, req)

				if w.Code != http.StatusOK {
					errorsChan <- fmt.Errorf("request %d failed with status %d", index, w.Code)
				}
			}(i)
		}

		wg.Wait()
		close(errorsChan)

		for err := range errorsChan {
			t.Error(err)
		}

		updatedDevice, _ := service.GetDevice(device.ID)
		if updatedDevice.SignatureCounter != concurrency {
			t.Errorf("expected counter %d, got %d", concurrency, updatedDevice.SignatureCounter)
		}
	})
}

func TestResponseFormats(t *testing.T) {
	t.Run("success response format", func(t *testing.T) {
		server, _ := setupTestServer()

		reqBody := model.CreateDeviceRequest{
			ID:        "device-format-001",
			Label:     "Format Test",
			Algorithm: "RSA",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Data == nil {
			t.Error("expected data field in response")
		}
	})

	t.Run("error response format", func(t *testing.T) {
		server, _ := setupTestServer()

		req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer([]byte("invalid")))
		w := httptest.NewRecorder()

		server.CreateDevice(w, req)

		var response ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		if err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if len(response.Errors) == 0 {
			t.Error("expected errors field in error response")
		}
	})
}
