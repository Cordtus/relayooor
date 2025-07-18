# IBC Packet Manager

A simplified web application for managing stuck IBC packets across Cosmos chains.

## Features

- **Query Stuck Packets**: Find non-acknowledged or non-relayed packets on any IBC channel
- **Cross-Validation**: Compare data from both Chainpulse monitoring and Hermes metrics
- **Individual Clearing**: Clear specific packets by sequence number
- **Bulk Operations**: Clear all stuck packets on a channel at once
- **Real-time Status**: See packet counts and clearing results instantly

## Usage

### Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

The app will be available at http://localhost:5174

### Production

The packet manager is included in the main docker-compose setup:

```bash
docker-compose up packet-manager
```

Access the packet manager at http://localhost:5174

## How It Works

1. **Select Chain and Channel**: Choose from available IBC chains and their channels
2. **Query Packets**: Click "Query Packets" to fetch stuck packets from both data sources
3. **Review Data**: 
   - Chainpulse shows packets stuck for >15 minutes
   - Hermes metrics show total pending packet counts
4. **Clear Packets**: 
   - Clear individual packets with the "Clear" button
   - Clear all packets at once with "Clear All Packets"

## Data Sources

### Chainpulse
- Monitors IBC activity in real-time
- Identifies packets stuck for more than 15 minutes
- Provides detailed packet information including sequences and timeouts

### Hermes Metrics
- Shows pending packet counts per channel
- Provides validation of stuck packet detection
- Updated every time packets are cleared

## API Integration

The packet manager uses the Relayooor API for all operations:

- `GET /api/v1/ibc/chains` - List available chains
- `GET /api/v1/ibc/chains/{chain}/channels` - List channels for a chain
- `GET /api/v1/ibc/packets/stuck` - Get stuck packets from Chainpulse
- `POST /api/v1/relayer/hermes/clear` - Clear packets via Hermes

## Security

- No authentication required (webapp is authorized by default)
- All clearing operations are logged
- Confirmation required for bulk operations

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Vue.js App     │────▶│   API Gateway   │────▶│   Chainpulse    │
│ (Packet Manager)│     │                 │     │   Monitoring    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │     Hermes      │
                        │  (IBC Relayer)  │
                        └─────────────────┘
```

## Troubleshooting

### No packets showing
- Ensure the selected chain and channel are correct
- Check that Chainpulse and Hermes are running
- Verify there are actually stuck packets on the channel

### Clearing fails
- Check Hermes logs for errors
- Ensure relayer wallets have sufficient funds
- Verify the channel is not closed

### Metrics not loading
- Hermes metrics are served on port 3010
- Check if Hermes telemetry is enabled
- Verify network connectivity between services