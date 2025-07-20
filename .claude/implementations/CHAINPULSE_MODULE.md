# Chainpulse Module Blueprint

**Last Updated**: 2025-01-19
**Version**: 0.4.0-pre (Enhanced Fork)
**Repository**: https://github.com/cordtus/chainpulse (main branch)

## Module Overview

Chainpulse is a forked IBC monitoring service that collects real-time packet data from Cosmos chains. The Relayooor fork includes major enhancements for complete packet data collection including sender/receiver addresses, amounts, denominations, and timeout information.

## Architecture

### Technology Stack
- **Language**: Rust
- **Async Runtime**: Tokio
- **Web Framework**: Axum (for REST API and metrics)
- **Database**: SQLite (changed from PostgreSQL for simplicity)
- **Metrics**: Prometheus
- **Chain Integration**: ibc-proto, tendermint-rpc

### Key Modifications from Original v0.3.x
1. **Complete packet data parsing** - extracts sender, receiver, amount, denom from packet data
2. **REST API endpoints** - comprehensive JSON API for packet queries
3. **Enhanced database schema** - stores full transfer details and timeout info
4. **User-based packet queries** - find packets by sender/receiver address
5. **Timeout tracking** - monitors packets approaching timeout
6. **Data integrity** - SHA256 hashing for duplicate detection
7. **CometBFT 0.37/0.38 compatibility**

## Core Components

### 1. Chain Monitor
```rust
pub struct ChainMonitor {
    chain_id: String,
    rpc_client: TendermintClient,
    db_pool: PgPool,
    metrics: MetricsCollector,
}

impl ChainMonitor {
    pub async fn start(&self) -> Result<()> {
        // Subscribe to new blocks
        let mut block_stream = self.rpc_client
            .subscribe_blocks()
            .await?;
        
        while let Some(block) = block_stream.next().await {
            self.process_block(block).await?;
        }
        
        Ok(())
    }
    
    async fn process_block(&self, block: Block) -> Result<()> {
        // Extract IBC events
        let ibc_events = self.extract_ibc_events(&block)?;
        
        // Process packets
        for event in ibc_events {
            match event {
                IbcEvent::SendPacket(packet) => {
                    self.handle_send_packet(packet).await?;
                }
                IbcEvent::RecvPacket(packet) => {
                    self.handle_recv_packet(packet).await?;
                }
                IbcEvent::AckPacket(packet) => {
                    self.handle_ack_packet(packet).await?;
                }
                IbcEvent::TimeoutPacket(packet) => {
                    self.handle_timeout_packet(packet).await?;
                }
            }
        }
        
        // Update metrics
        self.metrics.update_block_height(
            &self.chain_id, 
            block.header.height
        );
        
        Ok(())
    }
}
```

### 2. Packet Processor
```rust
pub struct PacketProcessor {
    db: PgPool,
    packet_cache: Arc<RwLock<HashMap<PacketKey, PacketInfo>>>,
}

impl PacketProcessor {
    pub async fn handle_send_packet(&self, packet: SendPacket) -> Result<()> {
        let packet_info = PacketInfo {
            id: generate_packet_id(&packet),
            src_chain_id: packet.source_chain,
            dst_chain_id: packet.destination_chain,
            src_channel_id: packet.source_channel,
            dst_channel_id: packet.destination_channel,
            src_port_id: packet.source_port,
            dst_port_id: packet.destination_port,
            sequence: packet.sequence,
            timeout_height: packet.timeout_height,
            timeout_timestamp: packet.timeout_timestamp,
            status: PacketStatus::Pending,
            created_at: Utc::now(),
            // Custom fields for Relayooor
            sender: extract_sender(&packet.data),
            receiver: extract_receiver(&packet.data),
            amount: extract_amount(&packet.data),
            denom: extract_denom(&packet.data),
        };
        
        // Store in database
        sqlx::query!(
            r#"
            INSERT INTO packets (
                id, src_chain_id, dst_chain_id, 
                src_channel_id, dst_channel_id,
                src_port_id, dst_port_id,
                sequence, timeout_timestamp,
                status, created_at, sender, receiver,
                amount, denom, data
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
            ON CONFLICT (id) DO NOTHING
            "#,
            packet_info.id,
            packet_info.src_chain_id,
            packet_info.dst_chain_id,
            packet_info.src_channel_id,
            packet_info.dst_channel_id,
            packet_info.src_port_id,
            packet_info.dst_port_id,
            packet_info.sequence as i64,
            packet_info.timeout_timestamp as i64,
            packet_info.status.to_string(),
            packet_info.created_at,
            packet_info.sender,
            packet_info.receiver,
            packet_info.amount,
            packet_info.denom,
            serde_json::to_value(&packet.data)?
        )
        .execute(&self.db)
        .await?;
        
        // Update cache
        self.packet_cache.write().await.insert(
            PacketKey::from(&packet_info),
            packet_info
        );
        
        Ok(())
    }
    
    pub async fn handle_recv_packet(&self, packet: RecvPacket) -> Result<()> {
        // Update packet status to received
        sqlx::query!(
            r#"
            UPDATE packets 
            SET status = 'received', 
                received_at = $1
            WHERE src_chain_id = $2 
                AND src_channel_id = $3 
                AND sequence = $4
            "#,
            Utc::now(),
            packet.source_chain,
            packet.source_channel,
            packet.sequence as i64
        )
        .execute(&self.db)
        .await?;
        
        Ok(())
    }
}
```

### 3. REST API Endpoints (NEW)

The enhanced fork provides a comprehensive REST API at `http://localhost:3001/api/v1`:

#### Available Endpoints

1. **Get Stuck Packets**
   ```
   GET /api/v1/packets/stuck?min_age_seconds=900&limit=100
   ```
   Returns undelivered packets with full transfer details.

2. **Find Packets by User**
   ```
   GET /api/v1/packets/by-user?address={address}&role={sender|receiver}
   ```
   Query packets for a specific user address.

3. **Get Packet Details**
   ```
   GET /api/v1/packets/{chain_id}/{channel}/{sequence}
   ```
   Retrieve complete packet information.

4. **Get Expiring Packets**
   ```
   GET /api/v1/packets/expiring?minutes=60
   ```
   Find packets approaching timeout.

5. **Get Channel Congestion**
   ```
   GET /api/v1/channels/congestion
   ```
   Returns congestion statistics with stuck packet counts.

6. **Get User Transfer History**
   ```
   GET /api/v1/users/{address}/transfers
   ```
   Complete transfer history for a user.

7. **Get Packet Analytics**
   ```
   GET /api/v1/packets/analytics?chain_id={chain}&period={24h|7d|30d}
   ```
   Analytics including delivery rates and timeout statistics.

#### Example Response
```json
{
  "packets": [
    {
      "chain_id": "osmosis-1",
      "sequence": 895396,
      "src_channel": "channel-750",
      "dst_channel": "channel-1",
      "sender": "osmo1m8wg4vxkefhs374qxmmqpyusgz289wmulex5qdwpfx7jnrxzer5s9cv83q",
      "receiver": "noble1ejmfwh88dxrehv345kj4743uznwpzkaz5tpv8z",
      "amount": "10832264",
      "denom": "transfer/channel-750/uusdc",
      "age_seconds": 330516,
      "relay_attempts": 3,
      "last_attempt_by": "osmo1j6swju2q7zywxmpcttcw4k98j7fphx5nu4scjy",
      "timeout_timestamp": 1234567890000000000,
      "ibc_version": "v1"
    }
  ],
  "total": 150,
  "api_version": "1.0"
}
```

### 4. API Implementation
```rust
pub fn configure_routes(cfg: &mut web::ServiceConfig) {
    cfg
        // Prometheus metrics endpoint
        .route("/metrics", web::get().to(metrics_handler))
        
        // REST API endpoints
        .service(
            web::scope("/api/v1")
                .route("/packets/stuck", web::get().to(get_stuck_packets))
                .route("/packets/by-user", web::get().to(get_user_packets))
                .route("/packets/expiring", web::get().to(get_expiring_packets))
                .route("/packets/{chain_id}/{channel}/{sequence}", web::get().to(get_packet))
                .route("/channels/congestion", web::get().to(get_channel_congestion))
                .route("/users/{address}/transfers", web::get().to(get_user_transfers))
                .route("/packets/analytics", web::get().to(get_analytics))
        );
}

async fn get_stuck_packets(
    query: web::Query<StuckPacketsQuery>,
    db: web::Data<PgPool>,
) -> Result<HttpResponse> {
    let min_stuck_duration = query.min_stuck_hours.unwrap_or(1);
    let cutoff = Utc::now() - Duration::hours(min_stuck_duration);
    
    let packets = sqlx::query_as!(
        PacketInfo,
        r#"
        SELECT 
            id, src_chain_id, dst_chain_id,
            src_channel_id, dst_channel_id,
            src_port_id, dst_port_id,
            sequence, timeout_timestamp,
            status as "status: PacketStatus",
            created_at, sender, receiver,
            amount, denom
        FROM packets
        WHERE status = 'pending'
            AND created_at < $1
            AND ($2::text IS NULL OR src_chain_id = $2)
            AND ($3::text IS NULL OR dst_chain_id = $3)
        ORDER BY created_at ASC
        LIMIT 1000
        "#,
        cutoff,
        query.src_chain,
        query.dst_chain
    )
    .fetch_all(db.get_ref())
    .await?;
    
    Ok(HttpResponse::Ok().json(packets))
}

async fn get_user_packets(
    path: web::Path<String>,
    db: web::Data<PgPool>,
) -> Result<HttpResponse> {
    let address = path.into_inner();
    
    let packets = sqlx::query_as!(
        PacketInfo,
        r#"
        SELECT 
            id, src_chain_id, dst_chain_id,
            src_channel_id, dst_channel_id,
            src_port_id, dst_port_id,
            sequence, timeout_timestamp,
            status as "status: PacketStatus",
            created_at, sender, receiver,
            amount, denom
        FROM packets
        WHERE sender = $1 OR receiver = $1
        ORDER BY created_at DESC
        LIMIT 100
        "#,
        address
    )
    .fetch_all(db.get_ref())
    .await?;
    
    Ok(HttpResponse::Ok().json(packets))
}
```

### 4. Metrics Collection
```rust
lazy_static! {
    static ref PACKETS_TOTAL: IntCounterVec = register_int_counter_vec!(
        "chainpulse_packets_total",
        "Total number of packets by status",
        &["chain", "status"]
    ).unwrap();
    
    static ref STUCK_PACKETS: IntGaugeVec = register_int_gauge_vec!(
        "chainpulse_stuck_packets",
        "Current number of stuck packets",
        &["src_chain", "dst_chain", "channel"]
    ).unwrap();
    
    static ref PACKET_PROCESSING_TIME: HistogramVec = register_histogram_vec!(
        "chainpulse_packet_processing_seconds",
        "Time to process packet events",
        &["event_type"]
    ).unwrap();
    
    static ref CHAIN_HEIGHT: IntGaugeVec = register_int_gauge_vec!(
        "chainpulse_chain_height",
        "Current block height by chain",
        &["chain"]
    ).unwrap();
}

pub async fn metrics_handler() -> Result<HttpResponse> {
    let encoder = TextEncoder::new();
    let metric_families = prometheus::gather();
    let mut buffer = vec![];
    encoder.encode(&metric_families, &mut buffer)?;
    
    Ok(HttpResponse::Ok()
        .content_type("text/plain; version=0.0.4")
        .body(buffer))
}
```

### 5. Chain Compatibility
```rust
pub enum ChainType {
    Standard,
    Neutron, // ABCI++ with vote extensions
}

pub struct ChainConfig {
    pub chain_id: String,
    pub chain_type: ChainType,
    pub rpc_endpoint: String,
    pub grpc_endpoint: Option<String>,
    pub ws_endpoint: Option<String>,
}

impl ChainConfig {
    pub fn create_client(&self) -> Result<Box<dyn ChainClient>> {
        match self.chain_type {
            ChainType::Standard => {
                Ok(Box::new(StandardChainClient::new(&self.rpc_endpoint)?))
            }
            ChainType::Neutron => {
                // Special handling for Neutron
                warn!("Neutron chain support is limited - ABCI++ not fully supported");
                Ok(Box::new(NeutronChainClient::new(&self.rpc_endpoint)?))
            }
        }
    }
}
```

## Database Schema (SQLite)

### Tables

#### txs table
```sql
CREATE TABLE txs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    chain        TEXT    NOT NULL,      -- Chain ID (e.g., 'osmosis-1')
    height       INTEGER NOT NULL,      -- Block height
    hash         TEXT    NOT NULL,      -- Transaction hash (hex)
    memo         TEXT    NOT NULL,      -- Transaction memo field
    created_at   TEXT    NOT NULL       -- Timestamp when recorded
);

-- Indexes
CREATE UNIQUE INDEX txs_unique ON txs (chain, hash);
CREATE INDEX txs_chain ON txs (chain);
CREATE INDEX txs_hash ON txs (hash);
CREATE INDEX txs_memo ON txs (memo);
CREATE INDEX txs_height ON txs (height);
CREATE INDEX txs_created_at ON txs (created_at);
```

#### packets table (Enhanced)
```sql
CREATE TABLE packets (
    id                              INTEGER PRIMARY KEY AUTOINCREMENT,
    tx_id                          INTEGER NOT NULL REFERENCES txs (id),
    sequence                       INTEGER NOT NULL,

    -- Channel/Port Information
    src_channel                    TEXT    NOT NULL,
    src_port                       TEXT    NOT NULL,
    dst_channel                    TEXT    NOT NULL,
    dst_port                       TEXT    NOT NULL,

    -- Message Information
    msg_type_url                   TEXT    NOT NULL,
    signer                         TEXT,                  -- Relayer address

    -- Relay Status
    effected                       BOOL    NOT NULL,
    effected_signer                TEXT,
    effected_tx                    INTEGER REFERENCES txs (id),

    -- User Transfer Data (NEW)
    sender                         TEXT,                  -- Original sender
    receiver                       TEXT,                  -- Destination address
    denom                          TEXT,                  -- Token denomination
    amount                         TEXT,                  -- Transfer amount

    -- Timeout Information (NEW)
    timeout_timestamp              INTEGER,               -- Unix nanoseconds
    timeout_height_revision_number INTEGER,
    timeout_height_revision_height INTEGER,

    -- Metadata (NEW)
    ibc_version                    TEXT DEFAULT 'v1',
    data_hash                      TEXT,                  -- SHA256 of packet data
    created_at                     TEXT    NOT NULL
);

-- Indexes
CREATE INDEX packets_tx_id ON packets(tx_id);
CREATE INDEX packets_signer ON packets (signer);
CREATE INDEX packets_src_channel ON packets (src_channel);
CREATE INDEX packets_dst_channel ON packets (dst_channel);
CREATE INDEX packets_effected ON packets (effected);
CREATE INDEX packets_effected_tx ON packets (effected_tx);
CREATE INDEX packets_sender ON packets (sender) WHERE sender IS NOT NULL;
CREATE INDEX packets_receiver ON packets (receiver) WHERE receiver IS NOT NULL;
CREATE INDEX packets_pending_sender ON packets (sender, effected) 
    WHERE effected = 0 AND sender IS NOT NULL;
CREATE INDEX packets_pending_receiver ON packets (receiver, effected) 
    WHERE effected = 0 AND receiver IS NOT NULL;
CREATE INDEX packets_stuck ON packets (src_channel, dst_channel, effected, created_at) 
    WHERE effected = 0;
CREATE INDEX packets_timeout_ts ON packets (timeout_timestamp) 
    WHERE timeout_timestamp IS NOT NULL;
CREATE INDEX packets_timeout_pending ON packets (timeout_timestamp, effected) 
    WHERE effected = 0 AND timeout_timestamp IS NOT NULL;
CREATE INDEX packets_data_hash ON packets (data_hash) 
    WHERE data_hash IS NOT NULL;
```

## Configuration

### TOML Configuration
```toml
[database]
path = "/data/chainpulse.db"

[metrics]
enabled = true
port = 3001

[api]
enabled = true
port = 3001

# Chains are configured dynamically via environment variables
[chains.osmosis-1]
url = "${OSMOSIS_WS_URL}"
comet_version = "0.37"

[chains.cosmoshub-4]
url = "${COSMOS_WS_URL}"
comet_version = "0.37"

[chains.noble-1]
url = "${NOBLE_WS_URL}"
comet_version = "0.37"

[chains.neutron-1]
url = "${NEUTRON_WS_URL}"
comet_version = "0.38"  # ABCI++ support limited

[chains.stride-1]
url = "${STRIDE_WS_URL}"
comet_version = "0.37"

[chains.axelar-dojo-1]
url = "${AXELAR_WS_URL}"
comet_version = "0.37"

[chains.jackal-1]
url = "${JACKAL_WS_URL}"
comet_version = "0.37"
```

## Deployment

### Dockerfile
```dockerfile
# Build stage - use latest Rust for compatibility
FROM rust:1-slim-bullseye as builder

# Install git to clone the repository
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

# Clone the forked chainpulse from main branch
RUN git clone -b main https://github.com/cordtus/chainpulse.git
WORKDIR /usr/src/chainpulse

# Show commit info for debugging
RUN git log -1 --oneline

# Build the binary
RUN cargo build --release

# Runtime stage
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /usr/src/chainpulse/target/release/chainpulse /usr/local/bin/chainpulse

# Create data directory
RUN mkdir -p /data

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose metrics and API port
EXPOSE 3001

VOLUME ["/data", "/config"]

ENTRYPOINT ["/entrypoint.sh"]
CMD ["chainpulse"]
```

### Health Check
```rust
async fn health_check(db: web::Data<PgPool>) -> Result<HttpResponse> {
    // Check database connection
    let db_healthy = sqlx::query("SELECT 1")
        .fetch_one(db.get_ref())
        .await
        .is_ok();
    
    // Check chain monitors
    let monitors_healthy = CHAIN_MONITORS
        .read()
        .await
        .values()
        .all(|m| m.is_healthy());
    
    let status = if db_healthy && monitors_healthy {
        "healthy"
    } else {
        "unhealthy"
    };
    
    Ok(HttpResponse::Ok().json(json!({
        "status": status,
        "database": db_healthy,
        "monitors": monitors_healthy,
        "version": env!("CARGO_PKG_VERSION"),
    })))
}
```

## Known Issues & Recent Improvements

### 1. Neutron Chain Support
**Issue**: Cannot decode ABCI++ vote extensions
**Impact**: Chain shows as "degraded", limited packet tracking
**Status**: Monitoring continues but with occasional errors

### 2. Chain Connection Stability
**Current Status**: 
- osmosis-1, cosmoshub-4, noble-1, axelar-dojo-1: Connected and stable
- stride-1, jackal-1: Showing reconnection issues
**Solution**: Automatic reconnection with exponential backoff

### 3. Recent Enhancements (v0.4.0-pre)
**Completed**:
- ✅ Full packet data extraction (sender, receiver, amount, denom)
- ✅ REST API for packet queries
- ✅ Timeout tracking and expiration monitoring
- ✅ User-based packet searches
- ✅ Enhanced database schema with comprehensive indexes
- ✅ Data integrity via SHA256 hashing

**In Progress**:
- 🔄 WebSocket API for real-time updates
- 🔄 Packet archival strategy for long-term storage
- 🔄 GraphQL API for flexible queries

## Performance Optimization

### 1. Batch Processing
```rust
// Process packets in batches
let mut packet_batch = Vec::with_capacity(BATCH_SIZE);

for event in events {
    packet_batch.push(event);
    
    if packet_batch.len() >= BATCH_SIZE {
        process_packet_batch(&packet_batch).await?;
        packet_batch.clear();
    }
}

// Process remaining packets
if !packet_batch.is_empty() {
    process_packet_batch(&packet_batch).await?;
}
```

### 2. Connection Pooling
```rust
// Reuse RPC connections
lazy_static! {
    static ref RPC_POOL: Arc<RwLock<HashMap<String, TendermintClient>>> = 
        Arc::new(RwLock::new(HashMap::new()));
}

async fn get_rpc_client(endpoint: &str) -> Result<TendermintClient> {
    let pool = RPC_POOL.read().await;
    
    if let Some(client) = pool.get(endpoint) {
        return Ok(client.clone());
    }
    
    drop(pool);
    
    let client = TendermintClient::new(endpoint).await?;
    RPC_POOL.write().await.insert(endpoint.to_string(), client.clone());
    
    Ok(client)
}
```

### 3. Caching Strategy
```rust
// Cache frequently accessed data
pub struct PacketCache {
    packets: Arc<RwLock<LruCache<String, PacketInfo>>>,
    ttl: Duration,
}

impl PacketCache {
    pub async fn get(&self, id: &str) -> Option<PacketInfo> {
        self.packets.read().await.get(id).cloned()
    }
    
    pub async fn set(&self, id: String, packet: PacketInfo) {
        self.packets.write().await.put(id, packet);
    }
}
```

## Monitoring Integration

### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "Chainpulse IBC Monitoring",
    "panels": [
      {
        "title": "Packets per Hour",
        "targets": [
          {
            "expr": "rate(chainpulse_packets_total[1h])"
          }
        ]
      },
      {
        "title": "Stuck Packets by Chain",
        "targets": [
          {
            "expr": "chainpulse_stuck_packets"
          }
        ]
      },
      {
        "title": "Processing Time",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, chainpulse_packet_processing_seconds)"
          }
        ]
      }
    ]
  }
}
```