package main

import (
	"log"

	"github.com/bayuhutajulu/signing-service/api"
	"github.com/bayuhutajulu/signing-service/domain"
	"github.com/bayuhutajulu/signing-service/persistence"
)

const (
	ListenAddress = ":8080"
)

func main() {
	storage := persistence.NewInMemoryStorage()
	service := domain.NewSignatureDeviceService(storage)
	server := api.NewServer(ListenAddress, service)

	if err := server.Run(); err != nil {
		log.Fatal("Could not start server on ", ListenAddress)
	}
}
