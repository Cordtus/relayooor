# Relayooor - IBC Relayer Monorepo

A comprehensive monorepo for IBC relayer operations, monitoring, and management. This project provides tools for running, monitoring, and managing IBC relayers (Hermes and Go Relayer) with a focus on operational excellence.

## Repository Structure

```
relayooor/
├── hermes/              # Hermes IBC relayer (Rust)
├── relayer/             # Go IBC relayer
├── relayer-middleware/  # API backend for relayer management
├── monitoring/          # Chainpulse monitoring tool
├── webapp/              # React dashboard frontend
├── config/              # Shared configuration files
├── docker/              # Docker configurations
└── scripts/             # Utility scripts
```

## Projects

### 1. Hermes (`/hermes`)
The Rust implementation of the IBC relayer by Informal Systems.

### 2. Go Relayer (`/relayer`)
The official Go implementation of the IBC relayer.

### 3. Relayer Middleware (`/relayer-middleware`)
Go-based API server that provides:
- Unified interface for managing both relayers
- WebSocket support for real-time updates
- Authentication and authorization
- Metrics collection and aggregation

### 4. Monitoring (`/monitoring`)
Chainpulse monitoring tool for tracking relayer performance and IBC network health.

### 5. Web Application (`/webapp`)
React-based dashboard for:
- Real-time packet flow visualization
- Stuck packet detection and clearing
- Channel and chain management
- Configuration editing
- Performance metrics

## Features

- **Dual Relayer Support**: Manage both Hermes and Go relayer from a single interface
- **Real-time Monitoring**: WebSocket-based live updates for packet flow and relayer status
- **Stuck Packet Detection**: Identify and clear stuck IBC packets across channels
- **Authentication**: JWT-based access control for secure operations
- **Metrics Dashboard**: Visualize packet flow, channel status, and relayer performance
- **Configuration Management**: Edit and update relayer configurations through the UI
- **Production Ready**: Optimized for deployment on various platforms including Fly.io

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     Webapp      │────▶│Relayer Middleware│────▶│    Relayers     │
│  (React + Vite) │     │   (Go API)      │     │ (Hermes + Rly)  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                        │
         │                       ▼                        │
         │              ┌─────────────────┐               │
         └─────────────▶│   Monitoring    │               │
                        │  (Chainpulse)   │               │
                        └─────────────────┘               │
                                                          │
                                                          ▼
                                                 ┌─────────────────┐
                                                 │   IBC Networks  │
                                                 └─────────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Node.js 20+ (for local development)
- Go 1.21+ (for local development)

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/yourusername/relayooor.git
cd relayooor
```

2. Copy environment configuration:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start the services:
```bash
docker-compose up -d
```

4. Access the dashboard at http://localhost:8080

Default credentials:
- Username: `admin`
- Password: `admin123`

## Configuration

### Hermes Configuration

Place your Hermes configuration at `config/hermes/config.toml`

### Go Relayer Configuration

Place your Go relayer configuration at `config/relayer/config.yaml`

## API Endpoints

### Authentication
- `POST /auth/login` - Login with username/password
- `POST /auth/refresh` - Refresh JWT token

### IBC Operations
- `GET /ibc/chains` - List all configured chains
- `GET /ibc/channels` - List all channels
- `GET /ibc/packets/pending` - Get pending packets
- `POST /ibc/packets/clear` - Clear stuck packets

### Relayer Management
- `GET /relayer/status` - Get status of both relayers
- `POST /relayer/hermes/start` - Start Hermes
- `POST /relayer/rly/start` - Start Go relayer

## Deployment on Fly.io

1. Install Fly CLI:
```bash
curl -L https://fly.io/install.sh | sh
```

2. Create a new Fly app:
```bash
fly launch --name your-relayer-dashboard
```

3. Set secrets:
```bash
fly secrets set JWT_SECRET=your-secret-key
fly secrets set DB_PASSWORD=your-db-password
```

4. Deploy:
```bash
fly deploy
```

## Security Considerations

- Always use strong JWT secrets in production
- Configure CORS properly for your domain
- Use HTTPS in production
- Regularly update relayer binaries
- Monitor access logs

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.