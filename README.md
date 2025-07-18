# Relayooor - IBC Packet Clearing Platform

A comprehensive platform for IBC packet clearing and monitoring that provides secure solutions for stuck IBC transfers across the Cosmos ecosystem.

## Key Features

### Packet Clearing Service
- **Secure Token-Based Authorization**: One-time tokens with cryptographic signatures and 5-minute expiry
- **On-Chain Payment Verification**: Pay service fees via standard IBC transfer with memo
- **Automated Clearing**: Our Hermes relayer automatically clears stuck packets
- **Multi-Chain Support**: Works with Osmosis, Cosmos Hub, Neutron, and more
- **User Statistics**: Track your clearing history and success rates
- **Automatic Refunds**: Failed clearings are automatically refunded (specific failure types)
- **Duplicate Payment Protection**: Prevents double-charging with multi-layer detection

### Monitoring Dashboard
- **Real-time IBC Metrics**: Powered by Chainpulse with CometBFT 0.38 support
- **Stuck Packet Detection**: Identify packets stuck for >15 minutes
- **Channel Performance**: Track success rates and congestion
- **Relayer Analytics**: Competition analysis and performance metrics
- **Multi-Network Support**: Monitor Cosmos Hub, Osmosis, Neutron, Noble, and Stride
- **Predictive Analytics**: AI-powered volume and success rate predictions

### Wallet Integration
- **Keplr Support**: Connect your wallet to view your stuck transfers
- **Direct Payment**: Pay clearing fees directly from your wallet
- **QR Code Payments**: Mobile-friendly payment options
- **Batch Operations**: Clear multiple packets at once
- **Transaction History**: View all your cleared packets with pagination
- **Secure Authentication**: Sign messages for accessing personal data
- **Real-time Updates**: WebSocket notifications for clearing status

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Vue.js Apps   │────▶│  API Gateway    │────▶│   Chainpulse    │
│ - Main Dashboard│     │  (Go + Redis)   │     │  (Monitoring)   │
│ - Packet Manager│     └─────────────────┘     └─────────────────┘
└─────────────────┘
         │                       │                        │
         │                       ▼                        ▼
         │              ┌─────────────────┐     ┌─────────────────┐
         └─────────────▶│ Enhanced Service│────▶│  IBC Networks   │
                        │  - Circuit Break │     │ (Cosmos Chains) │
                        │  - Retry Logic   │     └─────────────────┘
                        │  - Auto Refunds  │              │
                        │  - Caching       │              │
                        └─────────────────┘              │
                                 │                        │
                                 ▼                        ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │  Hermes Relayer │────▶│   PostgreSQL    │
                        │ (Clearing Exec) │     │  (Operations)   │
                        └─────────────────┘     └─────────────────┘
```

### Enhanced Features

#### Robustness and Reliability
- **Circuit Breaker Pattern**: Prevents cascading failures with Hermes
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Graceful Shutdown**: Tracks active operations with up to 5-minute grace period
- **Health Monitoring**: Component-level health checks with degraded mode support

#### Performance
- **Smart Caching**: Redis caching with stampede prevention
- **Database Indexes**: Optimized queries with materialized views
- **Pagination**: Both standard and cursor-based for large datasets
- **Connection Pooling**: Auto-scaling database connections based on load

#### User Experience
- **Payment URIs**: Direct wallet integration with cosmos:// URIs
- **Price Oracle**: USD value estimates for all fees
- **Help System**: Built-in glossary and term definitions
- **Error Messages**: User-friendly errors with actionable guidance

## Repository Structure

```
relayooor/
├── webapp/                 # Vue.js frontend application
│   ├── src/
│   │   ├── components/    # Reusable UI components
│   │   ├── views/         # Page components
│   │   ├── services/      # API services
│   │   └── stores/        # Pinia state management
│   └── public/            # Static assets
├── relayer-middleware/     # Go API backend
│   └── api/
│       └── pkg/
│           ├── clearing/  # Packet clearing logic
│           ├── handlers/  # HTTP handlers
│           └── middleware/# Auth, CORS, logging
├── monitoring/            # Chainpulse IBC monitoring
│   └── chainpulse/       # Fork with user data support
├── docs/                  # Documentation
│   ├── deployment/       # Deployment guides
│   ├── packet-clearing-* # Feature documentation
│   └── *.md             # Various docs
└── config/               # Configuration files
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Node.js 20+ and Yarn
- PostgreSQL 15+ with pg_cron extension (optional)
- Go 1.21+
- For M1/M4 Macs: Docker Desktop with proper resource allocation

### Local Development

1. **Clone the repository:**
```bash
git clone https://github.com/yourusername/relayooor.git
cd relayooor
```

2. **Set up environment:**
```bash
cp .env.example .env
# Edit .env with your configuration:
# - SERVICE_WALLET_ADDRESS (for collecting fees)
# - CLEARING_SECRET_KEY (generate a strong secret)
# - Database and Redis URLs
# - RPC endpoints for supported chains
```

3. **Start all services with Docker:**
```bash
# Quick start with all services
./start-full-stack.sh

# Or manually:
docker-compose -f docker-compose.full.yml up -d
```

4. **Alternative: Start services individually:**
```bash
# Backend services
docker-compose up -d postgres chainpulse prometheus

# API server
cd relayer-middleware/api
go run cmd/server/main.go

# Frontend (for development)
cd webapp
yarn install
yarn dev  # Note: May have issues on M1/M4 Macs, use Docker instead
```

5. **Access the application:**
- Frontend: http://localhost (via nginx)
- API: http://localhost:8080
- Chainpulse API: http://localhost:3000
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3003 (admin/admin)

## How Packet Clearing Works

1. **Connect Wallet**: User connects their Keplr wallet
2. **View Stuck Packets**: See all stuck IBC transfers associated with your addresses
3. **Select & Review**: Choose packets to clear and review fees
4. **Make Payment**: Send payment with generated memo to service address
5. **Automatic Clearing**: Our system verifies payment and clears packets
6. **Get Results**: View transaction hashes and success status

### Fee Structure
- Base service fee: 1 TOKEN
- Per-packet fee: 0.1 TOKEN
- Gas fees: Estimated based on current network conditions
- Automatic refunds for overpayments

## Tools and Access Points

### Web Applications
- **Main Dashboard** (http://localhost:8080): Full-featured IBC monitoring and packet clearing platform
- **Packet Manager** (http://localhost:5174): Simplified tool for querying and clearing stuck packets

### API Endpoints
- **Main API**: http://localhost:3000/api/v1
- **Hermes REST API**: http://localhost:5185
- **Hermes Metrics**: http://localhost:3010/metrics
- **Chainpulse Metrics**: http://localhost:3001

## Configuration

### Required Chainpulse Modifications
To support user-based packet queries, Chainpulse needs to:
1. Parse IBC packet data to extract sender/receiver addresses
2. Add database indexes for user queries
3. Implement stuck packet detection (>15 minutes pending)

See [chainpulse-required-modifications.md](docs/chainpulse-required-modifications.md) for details.

### Environment Variables
Key configuration in `.env`:
```bash
# Service Configuration
SERVICE_WALLET_ADDRESS=cosmos1...  # Your service fee collection address
CLEARING_SECRET_KEY=...            # Strong secret for token signing

# Fees (in smallest denomination)
CLEARING_SERVICE_FEE=1000000       # 1 TOKEN
CLEARING_PER_PACKET_FEE=100000     # 0.1 TOKEN

# Infrastructure
DATABASE_URL=postgresql://...
REDIS_URL=redis://localhost:6379
HERMES_REST_URL=http://localhost:5185
```

## API Endpoints

### Clearing Operations
- `POST /api/v1/clearing/request-token` - Get clearing authorization token
- `POST /api/v1/clearing/verify-payment` - Verify payment transaction
- `GET /api/v1/clearing/status/:token` - Check clearing status

### User Management
- `POST /api/v1/auth/wallet-sign` - Authenticate with wallet signature
- `GET /api/v1/users/statistics` - Get user clearing statistics

### Platform Analytics
- `GET /api/v1/statistics/platform` - Platform-wide statistics

## Production Deployment

### Fly.io Deployment
See [deployment guide](docs/deployment/README.md) for detailed instructions.

Quick deploy:
```bash
fly launch --name relayooor-app
fly secrets set CLEARING_SECRET_KEY=... SERVICE_WALLET_ADDRESS=...
fly deploy
```

### Security Checklist
- [ ] Generate strong `CLEARING_SECRET_KEY`
- [ ] Secure service wallet private key
- [ ] Enable TLS for all connections
- [ ] Configure CORS for your domain only
- [ ] Set up monitoring alerts
- [ ] Regular security audits

## Testing

```bash
# Backend tests
cd relayer-middleware/api
go test ./...

# Frontend tests
cd webapp
yarn test
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Documentation

- [Architecture Overview](docs/packet-clearing-architecture.md)
- [API Documentation](docs/API_IMPROVEMENTS.md)
- [Deployment Guide](docs/deployment/README.md)
- [Full Stack Deployment](FULL_STACK_DEPLOYMENT.md)
- [Database Optimizations](relayer-middleware/api/migrations/)

## Security

- All payment verification happens on-chain
- One-time tokens prevent replay attacks
- Cryptographic signatures ensure authenticity
- Automatic refunds for edge cases
- Regular security audits recommended

## Roadmap

### Phase 1 (Current)
- Basic packet clearing functionality
- Payment verification system
- User statistics tracking
- Multi-chain support

### Phase 2 (Next)
- [ ] Direct wallet payment integration
- [ ] Mobile app support
- [ ] Bulk clearing discounts
- [ ] Advanced analytics

### Phase 3 (Future)
- [ ] Predictive stuck packet detection
- [ ] Automated clearing options
- [ ] Cross-chain clearing
- [ ] Decentralized clearing network

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Performance Optimizations

The platform includes extensive database optimizations for handling multiple networks:
- **Connection pooling** with pgx/v5 for optimal throughput
- **Time-series data** support with TimescaleDB integration
- **Materialized views** for expensive aggregations
- **Smart indexing** on all common query patterns
- **Automatic maintenance** procedures for data cleanup

## Acknowledgments

- [Chainpulse](https://github.com/Cordtus/chainpulse) fork with CometBFT 0.38 support
- [Hermes](https://github.com/informalsystems/hermes) for reliable packet clearing
- The Cosmos ecosystem for IBC protocol
- All contributors and users of Relayooor

---

**Need Help?** 
- Check our [documentation](docs/)
- Report issues on [GitHub](https://github.com/yourusername/relayooor/issues)