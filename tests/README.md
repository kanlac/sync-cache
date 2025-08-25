# Sync-Cache Integration Tests

This directory contains integration tests for the sync-cache library using Docker Compose to simulate real microservice scenarios.

## Directory Structure

```
tests/
├── go.mod                    # Independent test module dependencies
├── docker-compose.yml        # Multi-instance test environment
├── Dockerfile               # Container image for cache instances
├── cmd/server/main.go       # HTTP API server for testing
├── integration/             # Ginkgo BDD integration tests
└── README.md               # This file
```

## Dependencies

The test module uses separate dependencies from the main library:
- **Ginkgo/Gomega**: BDD testing framework
- **testcontainers-go**: Container management for tests
- **go-redis**: Redis client for test utilities

## Quick Start

1. **Start test environment:**
   ```bash
   cd tests/
   docker-compose up -d
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

## Test Environment

- **Redis**: Single Redis instance with persistence
- **Instance-A**: Cache instance on port 8080
- **Instance-B**: Cache instance on port 8081

Each instance has HTTP API endpoints:
- `GET /health` - Health check
- `POST /set` - Set key-value pair
- `GET /get?key=<key>` - Get value by key
- `DELETE /delete?key=<key>` - Delete key

## Manual Testing Example

```bash
# Set a value in instance-a
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"user:123","value":"John Doe"}'

# Wait for sync (2-3 seconds)
sleep 3

# Get the value from instance-b
curl "http://localhost:8081/get?key=user:123"
```

The value should be synchronized between instances via Redis streams.
