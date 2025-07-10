# Quick Start - Cosmos Hub <> Osmosis Monitoring

This guide will help you get the IBC monitoring stack running for Cosmos Hub and Osmosis.

## Prerequisites

- Docker and Docker Compose installed
- (Optional) Mnemonic phrases for funded accounts on both chains if you want to enable packet clearing

## 1. Clone and Setup

```bash
git clone https://github.com/yourusername/relayooor.git
cd relayooor
```

## 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` if you need to:
- Add RPC authentication credentials (if using private nodes)
- Change default passwords
- Modify JWT secret

## 3. Start the Stack

```bash
./start.sh
```

This will:
- Create necessary directories
- Start all services
- Show you the service status

## 4. Access Services

- **Web Dashboard**: http://localhost
  - View IBC metrics
  - Connect wallet to clear stuck packets
  
- **Grafana**: http://localhost:3003
  - Username: admin
  - Password: admin
  - Pre-configured dashboard for Cosmos <> Osmosis

- **Prometheus**: http://localhost:9090
  - Query raw metrics
  
- **API**: http://localhost:8080
  - REST API for programmatic access

## 5. Add Relayer Keys (Optional)

If you want the relayer to actually relay packets, you need to add keys:

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

## 6. Monitor Channels

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

## Stopping the Stack

```bash
docker-compose down
```

To also remove volumes (data):
```bash
docker-compose down -v
```