# Monitoring - Chainpulse

Advanced monitoring tool for IBC relayers and blockchain networks.

## Features

- Real-time packet flow monitoring
- Chain status tracking
- Relayer performance metrics
- Alert system for stuck packets
- Historical data analysis
- Prometheus metrics export

## Architecture

Chainpulse is built in Rust for high performance and reliability. It continuously monitors:
- IBC packet flow across channels
- Chain health and block production
- Relayer transaction success rates
- Gas usage and fee optimization

## Configuration

Create a `chainpulse.toml` configuration file:

```toml
[general]
log_level = "info"
metrics_port = 9090

[chains]
[[chains.cosmoshub]]
rpc_url = "https://rpc.cosmos.network"
grpc_url = "https://grpc.cosmos.network"

[[chains.osmosis]]
rpc_url = "https://rpc.osmosis.zone"
grpc_url = "https://grpc.osmosis.zone"

[monitoring]
check_interval = 30
alert_threshold = 100
```

## Running

### Local Development

```bash
cd monitoring/chainpulse
cargo run -- --config chainpulse.toml
```

### Docker

```bash
docker build -t chainpulse .
docker run -p 9090:9090 -v $(pwd)/chainpulse.toml:/config.toml chainpulse
```

## Metrics

Chainpulse exports Prometheus metrics on the configured port:

- `ibc_packets_pending` - Number of pending packets per channel
- `ibc_packets_stuck` - Number of stuck packets (pending > threshold)
- `chain_latest_height` - Latest block height per chain
- `relayer_tx_success_rate` - Transaction success rate
- `relayer_balance` - Relayer account balances

## Integration

Chainpulse integrates with:
- Prometheus for metrics collection
- Grafana for visualization
- Alertmanager for notifications
- The relayer middleware API

## Development

### Prerequisites
- Rust 1.70+
- Protocol Buffers compiler

### Building

```bash
cargo build --release
```

### Testing

```bash
cargo test
```