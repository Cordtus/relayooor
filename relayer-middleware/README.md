# Relayer Middleware

Dockerized setup for running multiple IBC relayers with support for both current and legacy versions.

## Features

- **Multiple Relayer Support**: Hermes and Go relayer (rly)
- **Legacy Version Support**: Older versions for chains requiring them
- **RPC Authentication**: Built-in support for username/password authenticated nodes
- **Supervisor Management**: Reliable process management
- **Configurable**: Easy switching between relayers via environment variables

## Available Relayers

1. **hermes**: Latest Hermes relayer
2. **hermes-legacy**: Hermes v1.4.1 for older chains
3. **rly**: Latest Go relayer
4. **rly-legacy**: Go relayer v2.3.1 for compatibility

## Quick Start

1. Create configuration directories:
```bash
mkdir -p config/{hermes,hermes-legacy,relayer,relayer-legacy,keys}
```

2. Add your relayer configurations:
   - Hermes: `config/hermes/config.toml`
   - Hermes Legacy: `config/hermes-legacy/config.toml`
   - Go Relayer: `config/relayer/config.yaml`
   - Go Relayer Legacy: `config/relayer-legacy/config.yaml`

3. Set environment variables:
```bash
export ACTIVE_RELAYER=hermes  # or hermes-legacy, rly, rly-legacy
export RPC_USERNAME=your_username  # Optional
export RPC_PASSWORD=your_password  # Optional
```

4. Start the middleware:
```bash
docker-compose up -d
```

## Configuration

### Environment Variables

- `ACTIVE_RELAYER`: Which relayer to run (hermes, hermes-legacy, rly, rly-legacy)
- `RPC_USERNAME`: Username for RPC authentication (optional)
- `RPC_PASSWORD`: Password for RPC authentication (optional)

### RPC Authentication

The middleware automatically injects authentication credentials into RPC URLs when `RPC_USERNAME` and `RPC_PASSWORD` are provided.

### Metrics Endpoints

- Hermes: http://localhost:3001/metrics
- Hermes Legacy: http://localhost:3002/metrics
- Go Relayer: http://localhost:3003/metrics
- Go Relayer Legacy: http://localhost:3004/metrics

## Managing Relayers

Access the supervisor console:
```bash
docker exec -it relayer-middleware supervisorctl
```

Commands:
- `status`: Show all relayer statuses
- `start <relayer>`: Start a specific relayer
- `stop <relayer>`: Stop a specific relayer
- `restart <relayer>`: Restart a specific relayer

## Logs

Logs are stored in the `./logs` directory:
- `hermes.out.log` / `hermes.err.log`
- `hermes-legacy.out.log` / `hermes-legacy.err.log`
- `rly.out.log` / `rly.err.log`
- `rly-legacy.out.log` / `rly-legacy.err.log`

## Building from Source

```bash
docker build -t relayer-middleware .
```

## Deployment

For production deployment, ensure you:
1. Use secure credentials for RPC authentication
2. Mount configurations as read-only volumes
3. Set up proper log rotation
4. Configure monitoring for the metrics endpoints