# Sync-Cache Integration Tests

This directory contains integration tests for the sync-cache library using Docker Compose to simulate real microservice scenarios.

## Dependencies

The test module uses separate dependencies from the main library:
- **Ginkgo/Gomega**: BDD testing framework
- **testcontainers-go**: Container management for tests
- **go-redis**: Redis client for test utilities

## Quick Start

1. **Start test environment:**
   ```bash
   cd tests/
   docker compose up -d
   ```

2. **Verify services are healthy:**
   ```bash
   curl http://localhost:8080/health  # instance-a
   curl http://localhost:8081/health  # instance-b
   ```

3. **Run integration tests:**
   ```bash
   go test ./integration/
   ```

## Manual Testing Example

```bash
# Set a value in instance-a
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"user:123","value":"John Doe"}'

# Get the value from instance-b
curl "http://localhost:8081/get?key=user:123"

# Update in instance-a
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"user:123","value":"Alice Doe"}'

# Get synchronized value from instance-b
curl "http://localhost:8081/get?key=user:123"
```

The value should be synchronized between instances via Redis streams.
