FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy only necessary source code
COPY cache/ ./cache/
COPY engine/ ./engine/
COPY tests/cmd/server/ ./cmd/server/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o sync-cache-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/sync-cache-server .

EXPOSE 8080

CMD ["./sync-cache-server"]
