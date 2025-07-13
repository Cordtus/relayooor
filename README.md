# Relayooor - IBC Packet Clearing Platform

A comprehensive platform for IBC packet clearing and monitoring, providing secure and user-friendly solutions for stuck IBC transfers across the Cosmos ecosystem.

## Key Features

### Packet Clearing Service
- **Secure Token-Based Authorization**: One-time tokens with cryptographic signatures
- **On-Chain Payment Verification**: Pay service fees via standard IBC transfer with memo
- **Automated Clearing**: Our Hermes relayer automatically clears stuck packets
- **Multi-Chain Support**: Works with Osmosis, Cosmos Hub, Neutron, and more
- **User Statistics**: Track your clearing history and success rates

### Monitoring Dashboard
- **Real-time IBC Metrics**: Powered by Chainpulse
- **Stuck Packet Detection**: Identify packets stuck for >15 minutes
- **Channel Performance**: Track success rates and congestion
- **Relayer Analytics**: Competition analysis and performance metrics

### Wallet Integration
- **Keplr Support**: Connect your wallet to view your stuck transfers
- **Batch Operations**: Clear multiple packets at once
- **Transaction History**: View all your cleared packets
- **Secure Authentication**: Sign messages for accessing personal data

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Vue.js App    │────▶│  API Gateway    │────▶│   Chainpulse    │
│ (User Interface)│     │  (Go + Redis)   │     │  (Monitoring)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                        │
         │                       ▼                        ▼
         │              ┌─────────────────┐     ┌─────────────────┐
         └─────────────▶│ Packet Clearing │────▶│  IBC Networks   │
                        │    Service      │     │ (Cosmos Chains) │
                        └─────────────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Hermes Relayer │
                        │ (Clearing Exec) │
                        └─────────────────┘
```

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
- Node.js 16+ and Yarn
- PostgreSQL (or SQLite for development)
- Redis
- Hermes relayer instance

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

3. **Start backend services:**
```bash
docker-compose up -d postgres redis chainpulse
cd relayer-middleware/api
go run cmd/server/main.go
```

4. **Start frontend:**
```bash
cd webapp
yarn install
yarn dev
```

5. **Access the application:**
- Frontend: http://localhost:5173
- API: http://localhost:3000
- Chainpulse metrics: http://localhost:3001/metrics

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
- [Implementation Plan](docs/packet-clearing-implementation-plan.md)
- [Edge Cases & Improvements](docs/packet-clearing-edge-cases.md)
- [Operator Guide](docs/operator-review.md)
- [User Experience Review](docs/user-experience-review.md)
- [Deployment Guide](docs/deployment/README.md)

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

## Acknowledgments

- [Chainpulse](https://github.com/informalsystems/chainpulse) for IBC monitoring
- [Hermes](https://github.com/informalsystems/hermes) for reliable packet clearing
- The Cosmos ecosystem for IBC protocol
- All contributors and users of Relayooor

---

**Need Help?** 
- Check our [documentation](docs/)
- Report issues on [GitHub](https://github.com/yourusername/relayooor/issues)
- Join our [Discord](https://discord.gg/relayooor) community