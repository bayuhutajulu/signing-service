# Signature Service - Implementation

This is an implementation of the fiskaly signature service coding challenge. The service provides RESTful API endpoints for managing cryptographic signature devices with signature chaining capabilities.

## Prerequisites

- Go 1.20+
- Make

## Project Setup

1. **Install dependencies:**
   ```bash
   make tidy
   ```

2. **Run the service:**
   ```bash
   make run
   ```
   The service will start on `http://localhost:8080`

3. **Run unit tests:**
   ```bash
   make test
   ```
   This runs all unit tests with coverage reporting.

4. **Test the workflow (in a separate terminal):**
   ```bash
   make test-workflow-complete
   ```
   This runs a complete integration workflow that:
   - Creates 4 devices (2 RSA, 2 ECC)
   - Signs data with each device
   - Lists all devices to verify

## API Endpoints

### Create Device
```bash
POST /api/v0/devices
Content-Type: application/json

{
  "id": "device-001",
  "label": "My Device",
  "algorithm": "RSA"  // or "ECC"
}
```

### Sign Data
```bash
POST /api/v0/devices/{id}/sign
Content-Type: application/json

{
  "data": "transaction data to sign"
}
```

### Get Device
```bash
GET /api/v0/devices/{id}
```

### List All Devices
```bash
GET /api/v0/devices
```

### Health Check
```bash
GET /api/v0/health
```

## Architecture

The implementation follows Clean Architecture principles with clear separation of concerns:

```
├── crypto/          # Cryptographic signing implementations (RSA, ECDSA)
├── model/           # Domain models and DTOs
├── domain/          # Core business logic and interfaces
├── persistence/     # Storage implementations (in-memory)
├── api/             # HTTP handlers and routing
└── main.go          # Dependency injection and server bootstrap
```

### Key Design Decisions

1. **Domain-Centric Architecture**: The domain layer defines interfaces (like `DeviceStorage`) that persistence must implement, following dependency inversion principle.

2. **Concurrency Safety**:
   - `sync.Mutex` at service level serializes signing operations to ensure atomic counter increments
   - `sync.RWMutex` at storage level allows concurrent reads while ensuring exclusive writes

3. **Interface-Based Design**: The `Signer` interface allows easy addition of new algorithms without modifying domain logic.

4. **Security**: Private keys are never exposed in API responses. Only safe fields (ID, label, algorithm, counter) are returned.

## Implementation Notes & Assumptions

### Signature Counter Behavior

The signature counter starts at **0**. When a signature is created:
1. The **current** counter value is used in the signed data format: `<counter>_<data>_<last_signature>`
2. The signature is generated
3. The counter is incremented **after** successful signature creation

This means:
- First signature uses counter=0
- Second signature uses counter=1
- And so on...

### Base Case (First Signature)

When `signature_counter == 0` (first signature), there is no previous signature. Per the spec, `last_signature` is set to `base64(device.id)`.

Example for first signature:
```
counter = 0
last_signature = base64("device-001")
signed_data = "0_transaction_data_base64(device-001)"
```

### Concurrency Model

**Why Mutex Over Channels?**

Mutex was chosen for signature counter protection because:
- Simple read-modify-write operation fits mutex semantics
- Lower overhead than channel-based coordination
- Clear critical section boundaries
- Easier to reason about in this use case

The mutex ensures strictly monotonic counter increments without gaps, which is critical for compliance requirements.

### Error Handling

- **409 Conflict**: Returned when attempting to create a device with an existing ID
- **400 Bad Request**: Invalid request body or missing parameters
- **404 Not Found**: Would require service-level distinction (currently returns 500)
- **500 Internal Server Error**: Device not found, signing failure, or storage errors

## Testing Strategy

The implementation includes comprehensive testing:

1. **Unit Tests** (`make test`):
   - Crypto signers (RSA, ECDSA)
   - Domain service with mock storage
   - In-memory storage with concurrency tests
   - API handlers with httptest

2. **Integration Tests**:
   - Full stack tests in `api/server_test.go`
   - Tests entire request/response cycle

3. **Concurrency Tests**:
   - 50 concurrent device creations
   - 100 concurrent signatures on same device
   - Verifies counter correctness under load

**Test Coverage**: 94%+ across all packages

## Future Production Considerations

### Scaling & Performance

If this service were to go to production with high throughput requirements:

1. **Redis for Counter Management**:
   - Use Redis `INCR` for atomic counter increments with lower latency
   - Reduces lock contention across service instances
   - Enables horizontal scaling of stateless service instances

2. **Database Persistence**:
   - Store devices in PostgreSQL/MySQL for durability
   - Use Redis as write-through cache for hot devices
   - Async write pattern: increment counter in Redis, async persist to DB

3. **Architecture**:
   ```
   Client → Load Balancer → Service Instances (stateless)
                                ↓
                            Redis (counter)
                                ↓
                         PostgreSQL (persistence)
   ```

### Database Migration

The current design already supports easy database migration:
- `DeviceStorage` interface abstracts storage implementation
- Swap `InMemoryStorage` with `PostgresStorage` implementing same interface
- No domain logic changes required

## Design Decisions & Trade-offs

### 1. Mutex vs Channels for Concurrency
**Decision**: Use `sync.Mutex` at service level for counter operations.

**Rationale**:
- Simple read-modify-write pattern fits mutex semantics naturally
- Lower overhead compared to channel-based coordination
- Clear critical section boundaries make reasoning easier
- Signing operations are fast, so lock contention is minimal

**Trade-off**: In extremely high-throughput scenarios, channels might offer better scalability, but for this use case mutex provides better clarity and performance.

### 2. Domain Defines Storage Interface
**Decision**: Place `DeviceStorage` interface in domain package, not persistence.

**Rationale**:
- Follows dependency inversion principle
- Domain doesn't depend on infrastructure
- Easy to swap storage implementations (in-memory → database)
- Testability: can mock storage for domain tests

**Trade-off**: Slightly less intuitive initially, but better long-term maintainability.

### 3. RWMutex at Storage Level
**Decision**: Use `sync.RWMutex` in `InMemoryStorage` instead of regular mutex.

**Rationale**:
- Allows concurrent reads (GetDevice, GetAllDevices)
- Only writes (Save, Update) need exclusive access
- Better performance for read-heavy workloads

**Trade-off**: Slightly more complex than regular mutex, but performance gains justify it.

### 4. In-Memory Storage Only
**Decision**: Implement only in-memory storage without database persistence.

**Rationale**:
- Challenge explicitly states "it is enough to store signature devices in memory"
- Keeps implementation focused on core requirements
- Interface design already supports future database migration

**Trade-off**: Data lost on restart, but acceptable for challenge scope.

## Assumptions & Limitations

### Assumptions
1. **Device IDs are unique**: Client provides unique device IDs. No UUID generation server-side.
2. **Single instance deployment**: In-memory storage assumes single server instance. For multi-instance, Redis/DB required.
3. **No authentication**: Challenge states "pretend there is only one user/organization".
4. **No device deletion**: Only create and read operations implemented.
5. **Error granularity**: Device-not-found returns 500 instead of 404 (would require service-level error distinction).

### Known Limitations
1. **No persistence**: Data lost on server restart.
2. **No device key rotation**: Keys generated once during device creation.
3. **No signature verification**: Only signature generation implemented.
4. **Limited error context**: Some errors return generic 500 status.
5. **No pagination**: GetAllDevices returns all devices (fine for in-memory, but would need pagination for DB).

## Time Spent

**Total**: Approximately 1 day

Breakdown:
- Architecture design and planning: 2 hours
- Core implementation (crypto, domain, persistence, API): 3 hours
- Testing (unit + integration + concurrency): 2 hours
- Documentation and refinement: 1 hour

## Challenge Feedback

**Difficulty**: Moderate

**Why**:
- **Straightforward aspects**:
  - Core requirements are clear and well-specified
  - Cryptographic primitives provided by Go stdlib
  - Signature chaining logic is elegant once understood

- **Challenging aspects**:
  - Ensuring strictly monotonic counter without gaps requires careful concurrency handling
  - Understanding the base case (counter=0 with base64(device_id)) needs careful reading
  - Deciding between different concurrency patterns (mutex vs channels) requires experience
  - Balancing simplicity with production-readiness considerations

**Overall**: Well-designed challenge that tests practical skills (concurrency, API design, testing) while remaining focused and completable in reasonable time.

## AI Tools Usage

This implementation was developed with assistance from **Claude Code (Anthropic)** for:

1. **Unit Test Generation**: I used AI to write comprehensive unit tests because it's more rewarding to focus on the core logic and domain implementation while letting AI systematically test edge cases. Humans can miss subtle edge cases, and AI excels at generating exhaustive test scenarios, especially for concurrency and error handling paths.

2. **README Documentation**: I believe good documentation should not only be clear but also engaging and enjoyable to read. To increase creativity and presentation quality, I used AI to help structure and style the README, making it more accessible and professionally formatted.

## Available Make Targets

```bash
make run                      # Run the application
make build                    # Build the application binary
make test                     # Run all tests with coverage
make tidy                     # Tidy Go modules
make test-health-check        # Test health endpoint
make test-create-device-rsa   # Test RSA device creation
make test-create-device-ecc   # Test ECC device creation
make test-get-device          # Test get device endpoint
make test-sign-data           # Test sign data endpoint
make test-get-all-devices     # Test list all devices endpoint
make test-workflow-complete   # Run complete workflow test
make help                     # Show all available targets
```
