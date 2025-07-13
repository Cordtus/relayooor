# Quick Start - IBC Packet Clearing Platform

This guide will help you get the Relayooor packet clearing platform running for Cosmos Hub and Osmosis.

## Prerequisites

- Docker and Docker Compose installed
- Service wallet address for collecting clearing fees
- (Optional) Mnemonic phrases for funded accounts if you want to run a relayer

## 1. Clone and Setup

```bash
git clone https://github.com/yourusername/relayooor.git
cd relayooor
```

## 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` to configure:
- `SERVICE_WALLET_ADDRESS` - Your wallet for collecting clearing fees (required)
- `CLEARING_SECRET_KEY` - Generate a strong secret for token signing (required)
- RPC authentication credentials (if using private nodes)
- Database and Redis URLs
- Fee structure (SERVICE_FEE, PER_PACKET_FEE)

## 3. Start the Stack

```bash
./start.sh
```

This will:
- Create necessary directories
- Start all services
- Show you the service status

## 4. Access Services

- **Packet Clearing App**: http://localhost
  - Connect Keplr wallet to view stuck transfers
  - Clear stuck IBC packets with one-click payment
  - View clearing history and statistics
  
- **Grafana**: http://localhost:3003
  - Username: admin
  - Password: admin
  - Pre-configured dashboard for Cosmos <> Osmosis

- **Prometheus**: http://localhost:9090
  - Query raw metrics
  
- **API**: http://localhost:3000
  - REST API for packet clearing operations
  - WebSocket support for real-time updates
  - JWT authentication for secure access

## 5. Test Packet Clearing

1. **Connect your wallet** at http://localhost
2. **View stuck packets** associated with your addresses
3. **Select packets to clear** and review fees
4. **Make payment** with the generated memo
5. **Monitor clearing progress** in real-time

## 6. Add Relayer Keys (Optional)

If you want to run your own relayer (not required for packet clearing):

```bash
# Add Cosmos Hub key
docker exec -it relayooor-relayer-1 bash
echo "your mnemonic phrase here" > /tmp/cosmos.mnemonic
hermes keys add --chain cosmoshub-4 --mnemonic-file /tmp/cosmos.mnemonic
rm /tmp/cosmos.mnemonic

# Add Osmosis key
echo "your mnemonic phrase here" > /tmp/osmosis.mnemonic
hermes keys add --chain osmosis-1 --mnemonic-file /tmp/osmosis.mnemonic
rm /tmp/osmosis.mnemonic
exit
```

Then restart the relayer:
```bash
docker-compose restart relayer
```

## 7. Monitor Channels

The system monitors:
- **Cosmos Hub**: channel-141 (to Osmosis)
- **Osmosis**: channel-0 (to Cosmos Hub)

## Troubleshooting

### Check logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f relayer
docker-compose logs -f chainpulse
```

### Service not starting?
```bash
# Check status
docker-compose ps

# Restart specific service
docker-compose restart [service-name]
```

### No metrics showing?
1. Wait 2-3 minutes for initial data collection
2. Check chainpulse is connected: `docker-compose logs chainpulse`
3. Verify Prometheus targets: http://localhost:9090/targets

### Can't connect wallet?
1. Make sure you have Keplr wallet installed
2. Osmosis and Cosmos Hub chains should be added to Keplr
3. Check browser console for errors

### Payment verification failing?
1. Ensure memo is copied exactly as shown
2. Wait for transaction to be confirmed (6+ blocks)
3. Check service wallet received the payment

### Clearing not starting?
1. Check Hermes relayer is running: `docker-compose ps relayer`
2. Verify Hermes REST API: `curl http://localhost:5185/version`
3. Check execution worker logs: `docker-compose logs api | grep worker`

## Stopping the Stack

```bash
docker-compose down
```

To also remove volumes (data):
```bash
docker-compose down -v
```