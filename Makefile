.PHONY: run build test test-race test-verbose coverage clean help test-health-check test-create-device test-create-device-ecc test-get-device test-sign-data test-get-all-devices lint fmt tidy

run:
	go run main.go

build:
	go build -o bin/signing-service .

test:
	go test ./... -coverprofile=coverage.out

tidy:
	go mod tidy

test-health-check:
	curl http://localhost:8080/api/v0/health

test-create-device-rsa:
	curl -X POST http://localhost:8080/api/v0/devices \
		-H "Content-Type: application/json" \
		-d '{"id": "device-001", "label": "Test Device", "algorithm": "RSA"}'

test-create-device-ecc:
	curl -X POST http://localhost:8080/api/v0/devices \
		-H "Content-Type: application/json" \
		-d '{"id": "device-002", "label": "ECC Device", "algorithm": "ECC"}'

test-get-device:
	curl http://localhost:8080/api/v0/devices/device-001

test-sign-data:
	curl -X POST http://localhost:8080/api/v0/devices/device-001/sign \
		-H "Content-Type: application/json" \
		-d '{"data": "test transaction data"}'

test-get-all-devices:
	curl http://localhost:8080/api/v0/devices

# Workflow test targets
test-workflow-create-rsa-1:
	@echo "Creating RSA device 1..."
	@curl -X POST http://localhost:8080/api/v0/devices \
		-H "Content-Type: application/json" \
		-d '{"id": "workflow-rsa-1", "label": "Workflow RSA Device 1", "algorithm": "RSA"}' && echo ""

test-workflow-create-rsa-2:
	@echo "Creating RSA device 2..."
	@curl -X POST http://localhost:8080/api/v0/devices \
		-H "Content-Type: application/json" \
		-d '{"id": "workflow-rsa-2", "label": "Workflow RSA Device 2", "algorithm": "RSA"}' && echo ""

test-workflow-create-ecc-1:
	@echo "Creating ECC device 1..."
	@curl -X POST http://localhost:8080/api/v0/devices \
		-H "Content-Type: application/json" \
		-d '{"id": "workflow-ecc-1", "label": "Workflow ECC Device 1", "algorithm": "ECC"}' && echo ""

test-workflow-create-ecc-2:
	@echo "Creating ECC device 2..."
	@curl -X POST http://localhost:8080/api/v0/devices \
		-H "Content-Type: application/json" \
		-d '{"id": "workflow-ecc-2", "label": "Workflow ECC Device 2", "algorithm": "ECC"}' && echo ""

test-workflow-create-devices: test-workflow-create-rsa-1 test-workflow-create-rsa-2 test-workflow-create-ecc-1 test-workflow-create-ecc-2
	@echo "All devices created successfully!"

test-workflow-sign-rsa-1: test-workflow-create-devices
	@echo "Signing data with RSA device 1..."
	@curl -X POST http://localhost:8080/api/v0/devices/workflow-rsa-1/sign \
		-H "Content-Type: application/json" \
		-d '{"data": "transaction from RSA device 1"}' && echo ""

test-workflow-sign-rsa-2: test-workflow-create-devices
	@echo "Signing data with RSA device 2..."
	@curl -X POST http://localhost:8080/api/v0/devices/workflow-rsa-2/sign \
		-H "Content-Type: application/json" \
		-d '{"data": "transaction from RSA device 2"}' && echo ""

test-workflow-sign-ecc-1: test-workflow-create-devices
	@echo "Signing data with ECC device 1..."
	@curl -X POST http://localhost:8080/api/v0/devices/workflow-ecc-1/sign \
		-H "Content-Type: application/json" \
		-d '{"data": "transaction from ECC device 1"}' && echo ""

test-workflow-sign-ecc-2: test-workflow-create-devices
	@echo "Signing data with ECC device 2..."
	@curl -X POST http://localhost:8080/api/v0/devices/workflow-ecc-2/sign \
		-H "Content-Type: application/json" \
		-d '{"data": "transaction from ECC device 2"}' && echo ""

test-workflow-sign-all: test-workflow-sign-rsa-1 test-workflow-sign-rsa-2 test-workflow-sign-ecc-1 test-workflow-sign-ecc-2
	@echo "All devices have signed data successfully!"

test-workflow-list-devices: test-workflow-sign-all
	@echo "Listing all devices..."
	@curl http://localhost:8080/api/v0/devices && echo ""

test-workflow-complete: test-workflow-list-devices
	@echo ""
	@echo "========================================="
	@echo "Workflow test completed successfully!"
	@echo "- Created 4 devices (2 RSA, 2 ECC)"
	@echo "- Signed data with each device"
	@echo "- Listed all devices"
	@echo "========================================="

help:
	@echo "Available targets:"
	@echo "  run                    - Run the application"
	@echo "  build                  - Build the application binary"
	@echo "  test                   - Run all tests with coverage"
	@echo "  tidy                   - Tidy Go modules"
	@echo "  test-health-check      - Test health endpoint"
	@echo "  test-create-device-rsa - Test device creation (RSA)"
	@echo "  test-create-device-ecc - Test device creation (ECC)"
	@echo "  test-get-device        - Test get device endpoint"
	@echo "  test-sign-data         - Test sign data endpoint"
	@echo "  test-get-all-devices   - Test list all devices endpoint"
	@echo ""
	@echo "Workflow tests:"
	@echo "  test-workflow-complete - Run complete workflow (create, sign, list)"