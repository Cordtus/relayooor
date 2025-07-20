# Hermes Module Blueprint

## Module Overview

Hermes is an IBC relayer implementation that Relayooor uses to execute packet clearing operations. This module covers the integration, configuration, and API usage of Hermes within the Relayooor platform.

## Architecture

### Technology Stack
- **Core**: Hermes IBC Relayer (Rust)
- **Version**: Latest from ghcr.io/informalsystems/hermes
- **API**: REST API on port 5185
- **Telemetry**: Prometheus metrics on port 3001
- **Configuration**: TOML-based configuration

### Integration Points
1. REST API for packet clearing commands
2. WebSocket for real-time events
3. Telemetry endpoint for monitoring
4. Health check endpoint

## Configuration

### Main Configuration (config.toml)
```toml
[global]
log_level = 'info'

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 5185

[telemetry]
enabled = true
host = '0.0.0.0'
port = 3001

# Chain configurations
[[chains]]
id = 'cosmoshub-4'
type = 'CosmosSdk'
rpc_addr = 'https://cosmos-rpc.example.com:443'
grpc_addr = 'https://cosmos-grpc.example.com:9090'
websocket_addr = 'wss://cosmos-rpc.example.com:443/websocket'
rpc_timeout = '30s'
account_prefix = 'cosmos'
key_name = 'relayer'
key_store_type = 'Test'
store_prefix = 'ibc'
default_gas = 100000
max_gas = 1000000
gas_price = { price = 0.025, denom = 'uatom' }
gas_multiplier = 1.2
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
memo_prefix = 'relayooor'
sequential_batch_tx = false

[chains.trust_threshold]
numerator = '1'
denominator = '3'

[chains.packet_filter]
policy = 'allow'
list = [
    ['transfer', '*'],
]

[chains.address_type]
derivation = 'cosmos'

[[chains]]
id = 'osmosis-1'
type = 'CosmosSdk'
rpc_addr = 'https://osmosis-rpc.example.com:443'
grpc_addr = 'https://osmosis-grpc.example.com:9090'
websocket_addr = 'wss://osmosis-rpc.example.com:443/websocket'
# ... similar configuration as above
```

### Docker Entrypoint Scripts

#### Standard Entrypoint (entrypoint.sh)
```bash
#!/bin/bash
set -e

# Wait for required services
echo "Waiting for required services..."
sleep 5

# Import keys if provided
if [ -n "$RELAYER_MNEMONIC" ]; then
    echo "Importing relayer key..."
    echo "$RELAYER_MNEMONIC" | hermes keys add \
        --chain cosmoshub-4 \
        --key-name relayer \
        --mnemonic-file /dev/stdin
fi

# Start Hermes
echo "Starting Hermes..."
exec hermes start
```

#### Proxy-Enabled Entrypoint (entrypoint-with-proxy.sh)
```bash
#!/bin/bash
set -e

# Configure proxy if needed
if [ -n "$HTTP_PROXY" ]; then
    export http_proxy=$HTTP_PROXY
    export https_proxy=$HTTP_PROXY
    export HTTPS_PROXY=$HTTP_PROXY
fi

# Continue with standard startup
exec /entrypoint.sh
```

## API Integration

### REST API Endpoints

#### 1. Clear Packet Endpoint
```bash
POST /clear_packets
Content-Type: application/json

{
  "chain_id": "cosmoshub-4",
  "channel_id": "channel-141",
  "port_id": "transfer",
  "sequences": [123, 124, 125]
}
```

#### 2. Query Packet Status
```bash
GET /query/packet/unreceived
?chain=cosmoshub-4
&channel=channel-141
&port=transfer
&sequences=123,124,125
```

#### 3. Health Check
```bash
GET /health

Response:
{
  "status": "healthy",
  "chains": {
    "cosmoshub-4": "connected",
    "osmosis-1": "connected"
  }
}
```

### Go Client Implementation
```go
package hermes

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// ClearPackets triggers packet clearing for specific sequences
func (c *Client) ClearPackets(chainID, channelID, portID string, sequences []uint64) error {
    req := ClearPacketsRequest{
        ChainID:   chainID,
        ChannelID: channelID,
        PortID:    portID,
        Sequences: sequences,
    }
    
    body, err := json.Marshal(req)
    if err != nil {
        return err
    }
    
    resp, err := c.httpClient.Post(
        c.baseURL+"/clear_packets",
        "application/json",
        bytes.NewReader(body),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("hermes returned status %d", resp.StatusCode)
    }
    
    return nil
}

// QueryUnreceivedPackets gets list of unreceived packets
func (c *Client) QueryUnreceivedPackets(chainID, channelID, portID string) ([]uint64, error) {
    url := fmt.Sprintf(
        "%s/query/packet/unreceived?chain=%s&channel=%s&port=%s",
        c.baseURL, chainID, channelID, portID,
    )
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result QueryPacketsResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result.Sequences, nil
}

// Health checks Hermes health status
func (c *Client) Health() (*HealthResponse, error) {
    resp, err := c.httpClient.Get(c.baseURL + "/health")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var health HealthResponse
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        return nil, err
    }
    
    return &health, nil
}
```

## Authentication Configuration

### RPC Authentication
For chains requiring authentication:
```toml
[[chains]]
id = 'cosmoshub-4'
rpc_addr = 'https://username:password@cosmos-rpc.example.com:443'
# Note: WebSocket endpoints require auth in URL
websocket_addr = 'wss://username:password@cosmos-rpc.example.com:443/websocket'
# gRPC typically doesn't support auth in URL
grpc_addr = 'cosmos-grpc.example.com:9090'
```

## Monitoring & Telemetry

### Prometheus Metrics
Key metrics exposed by Hermes:
```
# Packet metrics
hermes_acknowledged_packets_total
hermes_receive_packets_total
hermes_timeout_packets_total

# Chain metrics
hermes_chain_status
hermes_wallet_balance

# Performance metrics
hermes_tx_latency_submitted
hermes_tx_latency_confirmed

# Error metrics
hermes_send_packet_errors_total
hermes_acknowledgment_errors_total
```

### Grafana Dashboard Configuration
```json
{
  "dashboard": {
    "title": "Hermes IBC Relayer",
    "panels": [
      {
        "title": "Packets Cleared",
        "targets": [{
          "expr": "rate(hermes_acknowledged_packets_total[5m])"
        }]
      },
      {
        "title": "Chain Status",
        "targets": [{
          "expr": "hermes_chain_status"
        }]
      },
      {
        "title": "Transaction Latency",
        "targets": [{
          "expr": "histogram_quantile(0.95, hermes_tx_latency_confirmed)"
        }]
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "rate(hermes_send_packet_errors_total[5m])"
        }]
      }
    ]
  }
}
```

## Operational Procedures

### 1. Adding New Chain
```bash
# Add chain configuration to config.toml
# Then restart Hermes
docker-compose restart hermes

# Import key for new chain
docker exec -it hermes hermes keys add \
    --chain new-chain-id \
    --key-name relayer \
    --mnemonic-file key.txt
```

### 2. Troubleshooting Commands
```bash
# Check wallet balances
docker exec hermes hermes keys balance --chain cosmoshub-4

# Query client status
docker exec hermes hermes query clients --host-chain cosmoshub-4

# Check channel status
docker exec hermes hermes query channel ends \
    --chain cosmoshub-4 \
    --channel channel-141 \
    --port transfer

# Manual packet clear
docker exec hermes hermes clear packets \
    --chain cosmoshub-4 \
    --channel channel-141 \
    --port transfer
```

### 3. Performance Tuning
```toml
# Optimize for high throughput
[[chains]]
max_msg_num = 50  # Increase batch size
max_tx_size = 4194304  # 4MB
sequential_batch_tx = true  # Sequential processing

# Gas optimization
gas_multiplier = 1.5  # Increase for congested chains
default_gas = 200000
max_gas = 2000000
```

## Error Handling

### Common Errors and Solutions

#### 1. RPC Connection Failed
```
Error: RPC error to endpoint https://cosmos-rpc.example.com:443
```
**Solution**: Check RPC endpoint authentication and availability

#### 2. Insufficient Gas
```
Error: out of gas in location: WritePerByte; gasWanted: 100000, gasUsed: 150000
```
**Solution**: Increase gas_multiplier or default_gas in config

#### 3. Account Sequence Mismatch
```
Error: account sequence mismatch, expected 123, got 125
```
**Solution**: Usually self-corrects, but can restart Hermes if persistent

#### 4. Timeout Height Exceeded
```
Error: packet timeout height exceeded
```
**Solution**: Packet has timed out, needs timeout handling not clearing

## Security Considerations

### 1. Key Management
- Use hardware security modules (HSM) in production
- Rotate keys regularly
- Never commit keys to version control

### 2. Network Security
- Use TLS for all RPC connections
- Implement firewall rules for Hermes ports
- Restrict API access to authorized services only

### 3. Operational Security
```yaml
# Docker security settings
security_opt:
  - no-new-privileges:true
  - seccomp:unconfined
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE
read_only: true
tmpfs:
  - /tmp
  - /var/run
```

## Integration with Relayooor

### Clearing Service Flow
```go
// In clearing service
func (s *ClearingService) ExecuteClearing(packet *Packet) error {
    // 1. Verify packet is still stuck
    unreceived, err := s.hermes.QueryUnreceivedPackets(
        packet.SrcChainID,
        packet.SrcChannelID,
        packet.SrcPortID,
    )
    if err != nil {
        return err
    }
    
    if !contains(unreceived, packet.Sequence) {
        return ErrPacketAlreadyCleared
    }
    
    // 2. Execute clearing
    err = s.hermes.ClearPackets(
        packet.SrcChainID,
        packet.SrcChannelID,
        packet.SrcPortID,
        []uint64{packet.Sequence},
    )
    if err != nil {
        return err
    }
    
    // 3. Wait for confirmation
    confirmed := false
    for i := 0; i < 30; i++ {
        time.Sleep(2 * time.Second)
        
        unreceived, err = s.hermes.QueryUnreceivedPackets(
            packet.SrcChainID,
            packet.SrcChannelID,
            packet.SrcPortID,
        )
        if err != nil {
            continue
        }
        
        if !contains(unreceived, packet.Sequence) {
            confirmed = true
            break
        }
    }
    
    if !confirmed {
        return ErrClearingTimeout
    }
    
    return nil
}
```

## Testing

### Integration Test Setup
```go
func TestHermesIntegration(t *testing.T) {
    // Start test containers
    ctx := context.Background()
    
    hermesContainer, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image: "ghcr.io/informalsystems/hermes:latest",
                ExposedPorts: []string{"5185/tcp", "3001/tcp"},
                Mounts: []testcontainers.ContainerMount{
                    {
                        Source: "./test-config.toml",
                        Target: "/app/config.toml",
                    },
                },
            },
            Started: true,
        },
    )
    require.NoError(t, err)
    defer hermesContainer.Terminate(ctx)
    
    // Get endpoint
    endpoint, err := hermesContainer.Endpoint(ctx, "5185")
    require.NoError(t, err)
    
    // Create client and test
    client := hermes.NewClient("http://" + endpoint)
    
    health, err := client.Health()
    require.NoError(t, err)
    assert.Equal(t, "healthy", health.Status)
}
```

## Deployment

### Docker Compose Configuration
```yaml
hermes:
  image: ghcr.io/informalsystems/hermes:latest
  container_name: hermes
  ports:
    - "5185:5185"  # REST API
    - "3001:3001"  # Telemetry
  volumes:
    - ./config/hermes/config.toml:/app/config.toml:ro
    - ./config/hermes/entrypoint.sh:/entrypoint.sh:ro
  environment:
    - RUST_LOG=info
    - RELAYER_MNEMONIC=${RELAYER_MNEMONIC}
  entrypoint: ["/entrypoint.sh"]
  restart: unless-stopped
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:5185/health"]
    interval: 30s
    timeout: 10s
    retries: 3
```

### Production Considerations
1. Use separate Hermes instances per chain pair for isolation
2. Implement monitoring and alerting for stuck packets
3. Regular key balance monitoring
4. Automated key rotation procedures
5. Backup configuration and keys securely