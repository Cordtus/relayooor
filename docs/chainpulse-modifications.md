# Chainpulse Modifications for Packet Clearing

## Overview

These modifications enhance Chainpulse to support comprehensive packet tracking and user-centric clearing operations. The changes focus on extracting user addresses from IBC packets and maintaining detailed packet state.

## Core Modifications

### 1. Enhanced Packet Data Structure

**File: `src/types.rs`**
```rust
use serde::{Deserialize, Serialize};
use cosmrs::proto::cosmos::base::v1beta1::Coin as ProtoCoin;

/// IBC Fungible Token Transfer packet data
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FungibleTokenPacketData {
    pub denom: String,
    pub amount: String,
    pub sender: String,
    pub receiver: String,
    #[serde(default)]
    pub memo: String,
}

/// Complete packet information with user data
#[derive(Debug, Clone, Serialize)]
pub struct EnhancedPacketInfo {
    // Core identifiers
    pub chain_id: String,
    pub sequence: u64,
    pub src_channel: String,
    pub src_port: String,
    pub dst_channel: String,
    pub dst_port: String,
    
    // User data (extracted from packet)
    pub sender: Option<String>,
    pub receiver: Option<String>,
    pub denom: Option<String>,
    pub amount: Option<String>,
    pub transfer_memo: Option<String>,
    
    // Packet metadata
    pub timeout_height: Option<Height>,
    pub timeout_timestamp: Option<u64>,
    
    // Tracking data
    pub first_seen_height: i64,
    pub first_seen_at: DateTime<Utc>,
    pub state: PacketState,
    pub relay_attempts: Vec<RelayAttempt>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PacketState {
    Pending,
    Relayed,
    Acknowledged,
    TimedOut,
    Stuck { since: DateTime<Utc> },
}

#[derive(Debug, Clone, Serialize)]
pub struct RelayAttempt {
    pub relayer: String,
    pub tx_hash: String,
    pub height: i64,
    pub timestamp: DateTime<Utc>,
    pub successful: bool,
    pub gas_used: Option<u64>,
}
```

### 2. Database Schema Updates

**File: `migrations/002_enhanced_packets.sql`**
```sql
-- Enhanced packets table with user data
ALTER TABLE packets ADD COLUMN sender TEXT;
ALTER TABLE packets ADD COLUMN receiver TEXT;
ALTER TABLE packets ADD COLUMN denom TEXT;
ALTER TABLE packets ADD COLUMN amount TEXT;
ALTER TABLE packets ADD COLUMN transfer_memo TEXT;
ALTER TABLE packets ADD COLUMN packet_data_json TEXT;
ALTER TABLE packets ADD COLUMN state TEXT DEFAULT 'pending';
ALTER TABLE packets ADD COLUMN stuck_since TIMESTAMP;
ALTER TABLE packets ADD COLUMN first_seen_height INTEGER;
ALTER TABLE packets ADD COLUMN timeout_height_revision_number INTEGER;
ALTER TABLE packets ADD COLUMN timeout_height_revision_height INTEGER;
ALTER TABLE packets ADD COLUMN timeout_timestamp INTEGER;

-- Create indexes for user queries
CREATE INDEX idx_packets_sender ON packets(sender) WHERE sender IS NOT NULL;
CREATE INDEX idx_packets_receiver ON packets(receiver) WHERE receiver IS NOT NULL;
CREATE INDEX idx_packets_state ON packets(state);
CREATE INDEX idx_packets_stuck ON packets(stuck_since) WHERE state = 'stuck';

-- Add relay attempts tracking
CREATE TABLE IF NOT EXISTS relay_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    packet_id INTEGER NOT NULL,
    relayer TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    height INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    successful BOOLEAN NOT NULL,
    gas_used INTEGER,
    frontrun_by TEXT,
    
    FOREIGN KEY (packet_id) REFERENCES packets(id),
    INDEX idx_attempts_packet (packet_id),
    INDEX idx_attempts_relayer (relayer)
);

-- User-centric views
CREATE VIEW v_user_packets AS
SELECT 
    p.*,
    COUNT(ra.id) as attempt_count,
    MAX(ra.timestamp) as last_attempt
FROM packets p
LEFT JOIN relay_attempts ra ON p.id = ra.packet_id
WHERE p.sender IS NOT NULL OR p.receiver IS NOT NULL
GROUP BY p.id;

CREATE VIEW v_stuck_packets_summary AS
SELECT 
    src_channel,
    dst_channel,
    COUNT(*) as stuck_count,
    SUM(CAST(amount AS DECIMAL)) as total_amount,
    MIN(stuck_since) as oldest_stuck,
    GROUP_CONCAT(DISTINCT denom) as denoms
FROM packets
WHERE state = 'stuck'
GROUP BY src_channel, dst_channel;
```

### 3. Packet Processing Enhancement

**File: `src/processor.rs`**
```rust
use crate::types::{FungibleTokenPacketData, EnhancedPacketInfo, PacketState};

impl ChainProcessor {
    /// Process IBC RecvPacket message with user data extraction
    pub async fn process_recv_packet(
        &mut self,
        msg: &MsgRecvPacket,
        tx: &Transaction,
        effected: bool,
    ) -> Result<()> {
        // Extract packet user data if it's a fungible token transfer
        let (sender, receiver, denom, amount, transfer_memo) = 
            if msg.packet.source_port == "transfer" {
                match serde_json::from_slice::<FungibleTokenPacketData>(&msg.packet.data) {
                    Ok(ft_data) => (
                        Some(ft_data.sender),
                        Some(ft_data.receiver),
                        Some(ft_data.denom),
                        Some(ft_data.amount),
                        Some(ft_data.memo),
                    ),
                    Err(e) => {
                        debug!("Failed to parse fungible token data: {}", e);
                        (None, None, None, None, None)
                    }
                }
            } else {
                (None, None, None, None, None)
            };

        // Check if packet already exists
        let existing_packet = sqlx::query!(
            "SELECT id, state FROM packets 
             WHERE chain_id = ? AND sequence = ? AND src_channel = ? AND src_port = ?",
            self.chain_id,
            msg.packet.sequence as i64,
            msg.packet.source_channel,
            msg.packet.source_port
        )
        .fetch_optional(&self.db)
        .await?;

        let packet_id = match existing_packet {
            Some(p) => p.id,
            None => {
                // Insert new packet with user data
                let packet_id = sqlx::query!(
                    "INSERT INTO packets (
                        chain_id, sequence, src_channel, src_port, 
                        dst_channel, dst_port, sender, receiver, 
                        denom, amount, transfer_memo, packet_data_json,
                        first_seen_height, timeout_timestamp, state, created_at
                    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                    self.chain_id,
                    msg.packet.sequence as i64,
                    msg.packet.source_channel,
                    msg.packet.source_port,
                    msg.packet.destination_channel,
                    msg.packet.destination_port,
                    sender,
                    receiver,
                    denom,
                    amount,
                    transfer_memo,
                    serde_json::to_string(&msg.packet.data).ok(),
                    tx.height,
                    msg.packet.timeout_timestamp.map(|t| t as i64),
                    "pending",
                    Utc::now()
                )
                .execute(&self.db)
                .await?
                .last_insert_rowid();
                
                packet_id
            }
        };

        // Record relay attempt
        sqlx::query!(
            "INSERT INTO relay_attempts (
                packet_id, relayer, tx_hash, height, 
                timestamp, successful, gas_used
            ) VALUES (?, ?, ?, ?, ?, ?, ?)",
            packet_id,
            tx.signer,
            tx.hash,
            tx.height,
            Utc::now(),
            effected,
            tx.gas_used.map(|g| g as i64)
        )
        .execute(&self.db)
        .await?;

        // Update packet state if successfully relayed
        if effected {
            sqlx::query!(
                "UPDATE packets SET state = 'relayed' WHERE id = ?",
                packet_id
            )
            .execute(&self.db)
            .await?;
        }

        // Check if packet should be marked as stuck
        self.check_stuck_packet(packet_id).await?;

        // Update metrics
        self.update_packet_metrics(&msg.packet, effected, sender.as_deref()).await?;

        Ok(())
    }

    /// Check if a packet should be marked as stuck
    async fn check_stuck_packet(&self, packet_id: i64) -> Result<()> {
        let stuck_threshold = Duration::minutes(15);
        
        let should_mark_stuck = sqlx::query!(
            "SELECT created_at, state 
             FROM packets 
             WHERE id = ? 
               AND state = 'pending'
               AND created_at < ?",
            packet_id,
            Utc::now() - stuck_threshold
        )
        .fetch_optional(&self.db)
        .await?
        .is_some();

        if should_mark_stuck {
            sqlx::query!(
                "UPDATE packets 
                 SET state = 'stuck', stuck_since = ? 
                 WHERE id = ?",
                Utc::now(),
                packet_id
            )
            .execute(&self.db)
            .await?;
            
            info!("Marked packet {} as stuck", packet_id);
        }

        Ok(())
    }

    /// Process acknowledgment to update packet state
    pub async fn process_acknowledge_packet(
        &mut self,
        msg: &MsgAcknowledgement,
        tx: &Transaction,
    ) -> Result<()> {
        sqlx::query!(
            "UPDATE packets 
             SET state = 'acknowledged'
             WHERE chain_id = ? 
               AND sequence = ? 
               AND src_channel = ? 
               AND src_port = ?",
            self.chain_id,
            msg.packet.sequence as i64,
            msg.packet.source_channel,
            msg.packet.source_port
        )
        .execute(&self.db)
        .await?;

        Ok(())
    }

    /// Process timeout to update packet state
    pub async fn process_timeout_packet(
        &mut self,
        msg: &MsgTimeout,
        tx: &Transaction,
    ) -> Result<()> {
        sqlx::query!(
            "UPDATE packets 
             SET state = 'timed_out'
             WHERE chain_id = ? 
               AND sequence = ? 
               AND src_channel = ? 
               AND src_port = ?",
            self.chain_id,
            msg.packet.sequence as i64,
            msg.packet.source_channel,
            msg.packet.source_port
        )
        .execute(&self.db)
        .await?;

        Ok(())
    }
}
```

### 4. Enhanced Metrics

**File: `src/metrics.rs`**
```rust
use prometheus::{IntGaugeVec, GaugeVec, HistogramVec};

lazy_static! {
    /// Gauge for stuck packets by user
    pub static ref IBC_STUCK_PACKETS_BY_USER: IntGaugeVec = register_int_gauge_vec!(
        "ibc_stuck_packets_by_user",
        "Number of stuck packets by user address",
        &["address", "role"] // role = sender/receiver
    ).unwrap();
    
    /// Total value stuck by denomination
    pub static ref IBC_STUCK_VALUE: GaugeVec = register_gauge_vec!(
        "ibc_stuck_value",
        "Total value stuck in packets by denomination",
        &["denom", "src_chain", "dst_chain"]
    ).unwrap();
    
    /// Packet clearing success rate
    pub static ref IBC_CLEARING_SUCCESS_RATE: GaugeVec = register_gauge_vec!(
        "ibc_clearing_success_rate",
        "Success rate of packet clearing attempts",
        &["channel"]
    ).unwrap();
    
    /// Time packets remain stuck before clearing
    pub static ref IBC_STUCK_DURATION: HistogramVec = register_histogram_vec!(
        "ibc_stuck_duration_seconds",
        "How long packets remain stuck before being cleared",
        &["src_chain", "dst_chain", "denom"],
        exponential_buckets(60.0, 2.0, 10).unwrap() // 1m, 2m, 4m, ..., ~17h
    ).unwrap();
}

impl ChainProcessor {
    /// Update metrics with user data
    async fn update_packet_metrics(
        &self,
        packet: &Packet,
        effected: bool,
        sender: Option<&str>,
    ) -> Result<()> {
        // Existing metrics...
        
        // Update user-specific metrics if we have sender data
        if let Some(sender_addr) = sender {
            if !effected {
                IBC_STUCK_PACKETS_BY_USER
                    .with_label_values(&[sender_addr, "sender"])
                    .inc();
            }
        }
        
        // Query and update stuck value metrics
        let stuck_values = sqlx::query!(
            "SELECT denom, SUM(CAST(amount AS REAL)) as total_amount
             FROM packets
             WHERE state = 'stuck' AND denom IS NOT NULL
             GROUP BY denom"
        )
        .fetch_all(&self.db)
        .await?;
        
        for row in stuck_values {
            if let (Some(denom), Some(amount)) = (row.denom, row.total_amount) {
                IBC_STUCK_VALUE
                    .with_label_values(&[&denom, &self.chain_id, &packet.destination_channel])
                    .set(amount);
            }
        }
        
        Ok(())
    }
}
```

### 5. API Endpoints for Packet Data

**File: `src/api.rs`**
```rust
use axum::{Router, Json, extract::{Path, Query}};
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
pub struct UserPacketsQuery {
    pub address: String,
    pub role: Option<String>, // sender, receiver, both
    pub state: Option<String>, // pending, stuck, all
    pub limit: Option<i64>,
    pub offset: Option<i64>,
}

#[derive(Serialize)]
pub struct UserPacketsResponse {
    pub packets: Vec<UserPacket>,
    pub total: i64,
    pub summary: PacketSummary,
}

#[derive(Serialize)]
pub struct UserPacket {
    pub chain_id: String,
    pub sequence: i64,
    pub src_channel: String,
    pub dst_channel: String,
    pub sender: Option<String>,
    pub receiver: Option<String>,
    pub amount: Option<String>,
    pub denom: Option<String>,
    pub state: String,
    pub stuck_since: Option<DateTime<Utc>>,
    pub attempts: i64,
    pub last_attempt: Option<DateTime<Utc>>,
}

#[derive(Serialize)]
pub struct PacketSummary {
    pub total_stuck: i64,
    pub total_value: HashMap<String, String>, // denom -> amount
    pub oldest_stuck_hours: Option<f64>,
}

/// Get packets for a specific user address
pub async fn get_user_packets(
    Query(params): Query<UserPacketsQuery>,
    State(state): State<AppState>,
) -> Result<Json<UserPacketsResponse>, ApiError> {
    let role_filter = match params.role.as_deref() {
        Some("sender") => "sender = ?",
        Some("receiver") => "receiver = ?",
        _ => "(sender = ? OR receiver = ?)",
    };
    
    let state_filter = match params.state.as_deref() {
        Some("pending") => " AND state = 'pending'",
        Some("stuck") => " AND state = 'stuck'",
        _ => "",
    };
    
    let query = format!(
        "SELECT p.*, COUNT(ra.id) as attempts, MAX(ra.timestamp) as last_attempt
         FROM packets p
         LEFT JOIN relay_attempts ra ON p.id = ra.packet_id
         WHERE {} {}
         GROUP BY p.id
         ORDER BY p.created_at DESC
         LIMIT ? OFFSET ?",
        role_filter, state_filter
    );
    
    let packets = if params.role.as_deref() == Some("sender") || 
                     params.role.as_deref() == Some("receiver") {
        sqlx::query_as::<_, UserPacket>(&query)
            .bind(&params.address)
            .bind(params.limit.unwrap_or(50))
            .bind(params.offset.unwrap_or(0))
            .fetch_all(&state.db)
            .await?
    } else {
        sqlx::query_as::<_, UserPacket>(&query)
            .bind(&params.address)
            .bind(&params.address)
            .bind(params.limit.unwrap_or(50))
            .bind(params.offset.unwrap_or(0))
            .fetch_all(&state.db)
            .await?
    };
    
    // Get summary statistics
    let summary = get_packet_summary(&state.db, &params.address).await?;
    
    Ok(Json(UserPacketsResponse {
        packets,
        total: summary.total_stuck,
        summary,
    }))
}

/// Get channel congestion data
pub async fn get_channel_congestion(
    State(state): State<AppState>,
) -> Result<Json<Vec<ChannelCongestion>>, ApiError> {
    let congestion_data = sqlx::query!(
        "SELECT 
            src_channel,
            dst_channel,
            COUNT(*) as stuck_count,
            MIN(stuck_since) as oldest_stuck,
            GROUP_CONCAT(DISTINCT denom) as denoms,
            SUM(CAST(amount AS REAL)) as total_value
         FROM packets
         WHERE state = 'stuck'
         GROUP BY src_channel, dst_channel
         HAVING stuck_count > 0
         ORDER BY stuck_count DESC"
    )
    .fetch_all(&state.db)
    .await?
    .into_iter()
    .map(|row| ChannelCongestion {
        src_channel: row.src_channel,
        dst_channel: row.dst_channel,
        stuck_packets: row.stuck_count as u64,
        oldest_stuck_hours: row.oldest_stuck.map(|t| {
            (Utc::now() - t).num_hours() as f64
        }),
        denoms: row.denoms.map(|d| d.split(',').map(String::from).collect()),
        total_stuck_value: row.total_value.map(|v| v.to_string()),
    })
    .collect();
    
    Ok(Json(congestion_data))
}

/// Mount API routes
pub fn api_routes() -> Router {
    Router::new()
        .route("/packets/user", get(get_user_packets))
        .route("/packets/stuck/:channel", get(get_stuck_packets_by_channel))
        .route("/channels/congestion", get(get_channel_congestion))
        .route("/packets/:chain/:channel/:sequence", get(get_packet_details))
}
```

### 6. Background Tasks

**File: `src/tasks.rs`**
```rust
use tokio::time::{interval, Duration};

/// Background task to identify and mark stuck packets
pub async fn stuck_packet_monitor(db: SqlitePool) {
    let mut interval = interval(Duration::from_secs(60)); // Check every minute
    
    loop {
        interval.tick().await;
        
        if let Err(e) = check_for_stuck_packets(&db).await {
            error!("Error checking for stuck packets: {}", e);
        }
    }
}

async fn check_for_stuck_packets(db: &SqlitePool) -> Result<()> {
    let stuck_threshold = Duration::minutes(15);
    let cutoff = Utc::now() - stuck_threshold;
    
    // Find packets that should be marked as stuck
    let updated = sqlx::query!(
        "UPDATE packets 
         SET state = 'stuck', stuck_since = ?
         WHERE state = 'pending' 
           AND created_at < ?
           AND stuck_since IS NULL",
        Utc::now(),
        cutoff
    )
    .execute(db)
    .await?;
    
    if updated.rows_affected() > 0 {
        info!("Marked {} packets as stuck", updated.rows_affected());
        
        // Update metrics
        update_stuck_metrics(db).await?;
    }
    
    Ok(())
}

async fn update_stuck_metrics(db: &SqlitePool) -> Result<()> {
    // Update user-specific stuck metrics
    let user_stuck = sqlx::query!(
        "SELECT sender, COUNT(*) as count
         FROM packets
         WHERE state = 'stuck' AND sender IS NOT NULL
         GROUP BY sender"
    )
    .fetch_all(db)
    .await?;
    
    for row in user_stuck {
        if let Some(sender) = row.sender {
            IBC_STUCK_PACKETS_BY_USER
                .with_label_values(&[&sender, "sender"])
                .set(row.count);
        }
    }
    
    // Update value metrics
    let stuck_values = sqlx::query!(
        "SELECT denom, SUM(CAST(amount AS REAL)) as total
         FROM packets
         WHERE state = 'stuck' AND denom IS NOT NULL
         GROUP BY denom"
    )
    .fetch_all(db)
    .await?;
    
    for row in stuck_values {
        if let (Some(denom), Some(total)) = (row.denom, row.total) {
            IBC_STUCK_VALUE
                .with_label_values(&[&denom, "all", "all"])
                .set(total);
        }
    }
    
    Ok(())
}

/// Task to clean up old completed packets
pub async fn cleanup_old_packets(db: SqlitePool) {
    let mut interval = interval(Duration::from_secs(3600)); // Every hour
    
    loop {
        interval.tick().await;
        
        let cutoff = Utc::now() - Duration::days(30);
        
        match sqlx::query!(
            "DELETE FROM packets 
             WHERE state IN ('acknowledged', 'timed_out') 
               AND created_at < ?",
            cutoff
        )
        .execute(&db)
        .await {
            Ok(result) => {
                if result.rows_affected() > 0 {
                    info!("Cleaned up {} old packets", result.rows_affected());
                }
            }
            Err(e) => error!("Error cleaning up old packets: {}", e),
        }
    }
}
```

### 7. Integration with Main Application

**File: `src/main.rs` modifications**
```rust
// Add to main function
tokio::spawn(tasks::stuck_packet_monitor(db.clone()));
tokio::spawn(tasks::cleanup_old_packets(db.clone()));

// Add API routes
let api = api::api_routes()
    .with_state(AppState { db: db.clone() });

let app = Router::new()
    .route("/metrics", get(metrics_handler))
    .nest("/api/v1", api);
```

## Testing Recommendations

### 1. Unit Tests for Packet Parsing

```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_parse_fungible_token_data() {
        let data = r#"{
            "denom": "uosmo",
            "amount": "1000000",
            "sender": "osmo1sender...",
            "receiver": "cosmos1receiver...",
            "memo": "test transfer"
        }"#;
        
        let parsed: FungibleTokenPacketData = 
            serde_json::from_str(data).unwrap();
            
        assert_eq!(parsed.denom, "uosmo");
        assert_eq!(parsed.amount, "1000000");
        assert_eq!(parsed.sender, "osmo1sender...");
        assert_eq!(parsed.receiver, "cosmos1receiver...");
        assert_eq!(parsed.memo, "test transfer");
    }
    
    #[tokio::test]
    async fn test_stuck_packet_detection() {
        let db = create_test_db().await;
        
        // Insert old packet
        sqlx::query!(
            "INSERT INTO packets (chain_id, sequence, src_channel, 
             src_port, state, created_at) 
             VALUES (?, ?, ?, ?, ?, ?)",
            "test-chain",
            1,
            "channel-0",
            "transfer",
            "pending",
            Utc::now() - Duration::minutes(20)
        )
        .execute(&db)
        .await
        .unwrap();
        
        // Run stuck detection
        check_for_stuck_packets(&db).await.unwrap();
        
        // Verify packet marked as stuck
        let packet = sqlx::query!(
            "SELECT state FROM packets WHERE sequence = 1"
        )
        .fetch_one(&db)
        .await
        .unwrap();
        
        assert_eq!(packet.state, "stuck");
    }
}
```

### 2. Integration Tests

```rust
#[tokio::test]
async fn test_user_packet_query() {
    let app = create_test_app().await;
    
    // Insert test data
    insert_test_packets(&app.db).await;
    
    // Query user packets
    let response = app
        .oneshot(
            Request::builder()
                .uri("/api/v1/packets/user?address=osmo1test&role=sender")
                .body(Body::empty())
                .unwrap()
        )
        .await
        .unwrap();
        
    assert_eq!(response.status(), StatusCode::OK);
    
    let body: UserPacketsResponse = 
        serde_json::from_slice(&hyper::body::to_bytes(response).await.unwrap())
        .unwrap();
        
    assert!(!body.packets.is_empty());
}
```

## Migration Guide

1. **Backup existing database** before applying migrations
2. **Run migrations** sequentially
3. **Reindex** to populate user data for historical packets (optional)
4. **Update configuration** to enable new endpoints
5. **Monitor performance** as new indexes may affect write speed

## Performance Considerations

1. **Indexing Impact**: New indexes on sender/receiver will slow writes slightly
2. **Query Optimization**: User packet queries should use indexes effectively
3. **Background Tasks**: Stuck detection runs every minute with minimal overhead
4. **Metrics Updates**: Batch updates to Prometheus metrics to reduce overhead

## Security Notes

1. **Input Validation**: All user addresses must be validated before queries
2. **Rate Limiting**: API endpoints should implement rate limiting per IP
3. **Data Privacy**: Consider which packet data should be publicly accessible
4. **Query Limits**: Enforce reasonable limits on result set sizes