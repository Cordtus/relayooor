# Relayer Middleware

API backend service that provides a unified interface for managing IBC relayers (Hermes and Go Relayer).

## Features

- RESTful API for relayer operations
- WebSocket support for real-time updates
- JWT-based authentication
- Prometheus metrics integration
- Support for both Hermes and Go relayer

## API Endpoints

### Authentication
- `POST /auth/login` - User authentication
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/logout` - Logout and invalidate token

### IBC Operations
- `GET /ibc/chains` - List all configured chains
- `GET /ibc/channels` - List all channels
- `GET /ibc/packets/pending` - Get pending packets
- `POST /ibc/packets/clear` - Clear stuck packets

### Relayer Management
- `GET /relayer/status` - Get status of all relayers
- `POST /relayer/hermes/start` - Start Hermes relayer
- `POST /relayer/hermes/stop` - Stop Hermes relayer
- `POST /relayer/rly/start` - Start Go relayer
- `POST /relayer/rly/stop` - Stop Go relayer
- `GET /relayer/config` - Get current configuration
- `PUT /relayer/config` - Update configuration

### Metrics
- `GET /metrics` - Prometheus metrics endpoint
- `GET /metrics/summary` - Human-readable metrics summary

## Development

### Prerequisites
- Go 1.21+
- Redis (for caching and session management)

### Running Locally

```bash
cd relayer-middleware
go mod download
go run api/cmd/server/main.go
```

### Configuration

Set the following environment variables:
- `JWT_SECRET` - Secret key for JWT signing
- `REDIS_URL` - Redis connection URL
- `HERMES_CONFIG` - Path to Hermes config file
- `RLY_CONFIG` - Path to Go relayer config file

### Testing

```bash
go test ./...
```

## Docker

Build the Docker image:
```bash
docker build -t relayer-middleware .
```

Run with Docker:
```bash
docker run -p 8080:8080 relayer-middleware
```