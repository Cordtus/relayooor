# Quick Start Guide

Get started with Relayooor in 5 minutes!

## What is Relayooor?

Relayooor helps you clear stuck IBC (Inter-Blockchain Communication) transfers. If you have tokens stuck in transit between Cosmos chains, Relayooor can help get them moving again.

## Installation

### Requirements
- Docker and Docker Compose
- Node.js 18+ and Yarn
- 8GB RAM minimum

### Setup Steps

1. **Clone and Navigate**
   ```bash
   git clone <repository-url>
   cd relayooor
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   ```
   Edit `.env` and add your RPC endpoints for the chains you want to support.

3. **Build and Launch**
   ```bash
   cd webapp && yarn install && yarn build && cd ..
   ./scripts/setup-and-launch.sh
   ```

4. **Access the Application**
   Open http://localhost in your browser

## Using Relayooor

1. **Connect Your Wallet**
   - Click "Connect Wallet"
   - Approve the Keplr connection

2. **View Stuck Packets**
   - Your stuck transfers will appear automatically
   - Each packet shows the source, destination, and stuck amount

3. **Clear a Packet**
   - Click "Clear Packet" on any stuck transfer
   - Follow the payment instructions
   - Your packet will be cleared automatically

## Checking Service Status

```bash
# View all services
docker-compose ps

# Check logs
docker-compose logs -f

# Test if services are healthy
curl http://localhost/health
```

## Stopping the Application

```bash
docker-compose down
```

## Getting Help

- Check the [full documentation](./README.md)
- View [deployment guide](./DEPLOYMENT.md)
- Open an issue on GitHub

## Common Issues

**Services won't start?**
- Make sure Docker is running
- Check if ports 80, 8080, 3001 are available

**Can't see stuck packets?**
- Ensure your wallet is connected
- Check that you have stuck IBC transfers
- Verify RPC endpoints in `.env` are correct

**Frontend not updating?**
- Always run `yarn build` after making changes
- Clear your browser cache