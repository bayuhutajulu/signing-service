package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bayuhutajulu/signing-service/domain"
	"github.com/gorilla/mux"
)

// Response is the generic API response container.
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	listenAddress     string
	signDeviceService domain.ISignatureDeviceService
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string, signDeviceService *domain.SignatureDeviceService) *Server {
	return &Server{
		listenAddress:     listenAddress,
		signDeviceService: signDeviceService,
	}
}

// Run registers all HandlerFuncs for the existing HTTP routes and starts the Server.
func (s *Server) Run() error {
	router := mux.NewRouter()

	router.HandleFunc("/api/v0/health", s.Health).Methods(http.MethodGet)
	router.HandleFunc("/api/v0/devices", s.CreateDevice).Methods(http.MethodPost)
	router.HandleFunc("/api/v0/devices", s.GetAllDevices).Methods(http.MethodGet)
	router.HandleFunc("/api/v0/devices/{id}", s.GetDevice).Methods(http.MethodGet)
	router.HandleFunc("/api/v0/devices/{id}/sign", s.SignData).Methods(http.MethodPost)

	log.Printf("Server is starting on %s", s.listenAddress)
	return http.ListenAndServe(s.listenAddress, router)
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
		return
	}

	w.Write(bytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
		return
	}

	w.Write(bytes)
}
