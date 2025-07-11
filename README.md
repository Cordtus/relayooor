# Relayooor - IBC Relayer Monorepo

A comprehensive monorepo for IBC relayer operations, monitoring, and management. This project consists of three main components:

1. **Relayer Middleware**: Dockerized setup for running multiple IBC relayers (Hermes and Go Relayer)
2. **Monitoring**: Chainpulse-based monitoring and metrics collection for IBC networks
3. **Web App**: User-friendly dashboard for viewing metrics and clearing stuck packets via wallet connection

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
├── scripts/             # Utility scripts
├── .env.example         # Example environment variables
└── .gitignore           # Git ignore file (includes .env)
```

## Projects

### 1. Hermes (`/hermes`)
The Rust implementation of the IBC relayer by Informal Systems.

### 2. Go Relayer (`/relayer`)
The official Go implementation of the IBC relayer.

### 3. Relayer Middleware (`/relayer-middleware`)
Dockerized setup containing:
- Hermes relayer (with support for legacy versions)
- Go relayer (cosmos/relayer)
- Configuration management
- Automated deployment scripts

### 4. Monitoring (`/monitoring`)
Chainpulse-based monitoring system:
- Real-time IBC packet tracking
- Stuck packet detection
- Channel performance metrics
- Prometheus metrics export
- Extensible for custom metrics

### 5. Web Application (`/webapp`)
React-based dashboard for:
- User-friendly IBC metrics visualization
- Wallet integration for packet clearing
- Individual packet or entire channel clearing
- View user's stuck IBC transfers
- Clear stuck transfers with one click
- Real-time updates via WebSocket
- Authentication for access control

## Key Features

- **Multiple Relayer Support**: Run both Hermes and Go relayer with legacy version support
- **Comprehensive Monitoring**: Real-time IBC packet flow and stuck packet detection
- **Wallet Integration**: Connect wallet to clear packets relevant to your addresses
- **Flexible Packet Clearing**: Clear individual packets or entire channels
- **Production Ready**: Optimized for Fly.io deployment
- **Extensible Architecture**: Easy to add new monitoring metrics or relayer types
- **Secure Node Access**: Support for username/password authenticated RPC nodes

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Web App       │────▶│   Monitoring    │────▶│  IBC Networks   │
│ (React + Vite)  │     │  (Chainpulse)   │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                                                │
         │                                                │
         ▼                                                ▼
┌─────────────────┐                              ┌─────────────────┐
│Relayer Middleware│                             │    Relayers     │
│  (Docker Setup)  │────────────────────────────▶│ (Hermes + Rly)  │
└─────────────────┘                              └─────────────────┘
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

**IMPORTANT**: The `.env` file contains sensitive credentials. It is already included in `.gitignore` and will NOT be committed to the repository. Keep your credentials secure!

3. Start the services:
```bash
docker-compose up -d
```

4. Access the dashboard at http://localhost:8080

Default credentials:
- Username: `admin`
- Password: `admin123`

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

- `RPC_USERNAME` / `RPC_PASSWORD`: Authentication for RPC endpoints (if required)
- `JWT_SECRET`: Secret key for API authentication
- `GF_ADMIN_USER` / `GF_ADMIN_PASSWORD`: Grafana admin credentials
- `ACTIVE_RELAYER`: Choose between `hermes` or `go-relayer`

### Hermes Configuration

Place your Hermes configuration at `config/hermes/config.toml`

### Go Relayer Configuration

Place your Go relayer configuration at `config/relayer/config.yaml`

### Chainpulse Configuration

Chainpulse uses the forked version with authentication support:
- Repository: https://github.com/cordtus/chainpulse.git
- Config: `monitoring/config/chainpulse-cosmos-osmosis.toml`
- Authentication is automatically injected from environment variables

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