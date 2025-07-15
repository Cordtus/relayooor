# Configuration Directory

This directory contains configuration files for the Relayooor platform.

## Node Endpoints Configuration

### Security Notice ⚠️

**NEVER commit actual node endpoints or credentials to the repository!**

Private node endpoints should only be configured through environment variables.

### Setup Instructions

1. Copy `.env.example` to `.env` in the project root
2. Fill in your private node endpoints in the `.env` file:
   ```bash
   COSMOS_RPC_URL=https://username:password@your-private-cosmos-rpc.com
   COSMOS_API_URL=https://username:password@your-private-cosmos-api.com
   # ... etc
   ```

3. The application will automatically use these endpoints when available
4. If private endpoints are not configured, the application will fall back to public endpoints

### For Developers

- Frontend code should ONLY use public endpoints
- Backend services load private endpoints from environment variables
- All endpoint configuration is centralized in `/config/nodes.ts`
- The `nodes.toml.example` file shows the expected format if you need to generate a TOML config

### Chainpulse Configuration

Chainpulse configurations can be auto-generated from your node endpoints:

```bash
# Generate chainpulse config for configured chains
node scripts/generate-chainpulse-config.js > config/chainpulse-generated.toml
```

## Other Configuration Files

- `chainpulse*.toml` - Chainpulse monitoring configurations
- `hermes/` - Hermes relayer configurations
- `grafana/` - Grafana dashboard configurations
- `prometheus/` - Prometheus monitoring configurations