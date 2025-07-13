# Performance Optimizations Implementation Plan

## 1. Caching for Packet Data

### Redis Cache Implementation
```go
// relayer-middleware/api/pkg/cache/packet_cache.go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

type PacketCache struct {
    redis  *redis.Client
    ttl    time.Duration
    logger *zap.Logger
}

func NewPacketCache(redis *redis.Client) *PacketCache {
    return &PacketCache{
        redis:  redis,
        ttl:    5 * time.Minute, // Cache packets for 5 minutes
        logger: logging.With(zap.String("component", "packet_cache")),
    }
}

// Cache user's stuck packets
func (c *PacketCache) SetUserPackets(ctx context.Context, walletAddress string, packets []StuckPacket) error {
    key := c.userPacketsKey(walletAddress)
    
    data, err := json.Marshal(packets)
    if err != nil {
        return fmt.Errorf("failed to marshal packets: %w", err)
    }
    
    if err := c.redis.Set(ctx, key, data, c.ttl).Err(); err != nil {
        return fmt.Errorf("failed to cache packets: %w", err)
    }
    
    c.logger.Debug("Cached user packets",
        zap.String("wallet", walletAddress),
        zap.Int("count", len(packets)),
    )
    
    return nil
}

// Get cached user packets with cache stampede prevention
func (c *PacketCache) GetUserPackets(ctx context.Context, walletAddress string) ([]StuckPacket, bool, error) {
    key := c.userPacketsKey(walletAddress)
    lockKey := key + ":lock"
    
    // Try to get from cache
    data, err := c.redis.Get(ctx, key).Bytes()
    if err != nil && err != redis.Nil {
        return nil, false, fmt.Errorf("failed to get cached packets: %w", err)
    }
    
    if err == redis.Nil {
        // Cache miss - try to acquire lock to prevent stampede
        locked, err := c.redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
        if err != nil {
            return nil, false, fmt.Errorf("failed to acquire lock: %w", err)
        }
        
        if !locked {
            // Another process is fetching, wait briefly and retry
            time.Sleep(100 * time.Millisecond)
            
            // Try cache again
            data, err = c.redis.Get(ctx, key).Bytes()
            if err != nil {
                return nil, false, nil // Let caller fetch
            }
        } else {
            // We have the lock, caller should fetch
            defer c.redis.Del(ctx, lockKey)
            return nil, false, nil
        }
    }
    
    var packets []StuckPacket
    if err := json.Unmarshal(data, &packets); err != nil {
        return nil, false, fmt.Errorf("failed to unmarshal packets: %w", err)
    }
    
    // Implement probabilistic early expiration
    ttl, _ := c.redis.TTL(ctx, key).Result()
    if c.shouldRefreshCache(ttl) {
        // Trigger async refresh
        go func() {
            refreshCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            c.triggerCacheRefresh(refreshCtx, walletAddress)
        }()
    }
    
    c.logger.Debug("Retrieved cached packets",
        zap.String("wallet", walletAddress),
        zap.Int("count", len(packets)),
    )
    
    return packets, true, nil
}

func (c *PacketCache) shouldRefreshCache(ttl time.Duration) bool {
    // Refresh cache probabilistically when TTL < 1 minute
    if ttl < 1*time.Minute && ttl > 0 {
        // Probability increases as TTL decreases
        probability := 1.0 - (float64(ttl) / float64(1*time.Minute))
        return rand.Float64() < probability
    }
    return false
}

// Invalidate user's packet cache
func (c *PacketCache) InvalidateUserPackets(ctx context.Context, walletAddress string) error {
    key := c.userPacketsKey(walletAddress)
    
    if err := c.redis.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("failed to invalidate cache: %w", err)
    }
    
    return nil
}

// Cache channel statistics
func (c *PacketCache) SetChannelStats(ctx context.Context, channelID string, stats *ChannelStats) error {
    key := c.channelStatsKey(channelID)
    
    // Use longer TTL for channel stats as they change less frequently
    ttl := 15 * time.Minute
    
    return c.redis.Set(ctx, key, stats, ttl).Err()
}

func (c *PacketCache) GetChannelStats(ctx context.Context, channelID string) (*ChannelStats, bool, error) {
    key := c.channelStatsKey(channelID)
    
    var stats ChannelStats
    err := c.redis.Get(ctx, key).Scan(&stats)
    if err != nil {
        if err == redis.Nil {
            return nil, false, nil
        }
        return nil, false, err
    }
    
    return &stats, true, nil
}

// Cache gas estimates
func (c *PacketCache) SetGasEstimate(ctx context.Context, chainID string, estimate *GasEstimate) error {
    key := c.gasEstimateKey(chainID)
    
    // Gas estimates valid for 30 minutes
    ttl := 30 * time.Minute
    
    return c.redis.Set(ctx, key, estimate, ttl).Err()
}

func (c *PacketCache) GetGasEstimate(ctx context.Context, chainID string) (*GasEstimate, bool, error) {
    key := c.gasEstimateKey(chainID)
    
    var estimate GasEstimate
    err := c.redis.Get(ctx, key).Scan(&estimate)
    if err != nil {
        if err == redis.Nil {
            return nil, false, nil
        }
        return nil, false, err
    }
    
    return &estimate, true, nil
}

// Multi-get for efficiency
func (c *PacketCache) GetMultipleUsers(ctx context.Context, walletAddresses []string) (map[string][]StuckPacket, error) {
    if len(walletAddresses) == 0 {
        return make(map[string][]StuckPacket), nil
    }
    
    // Use pipeline for efficiency
    pipe := c.redis.Pipeline()
    
    keys := make([]string, len(walletAddresses))
    for i, wallet := range walletAddresses {
        keys[i] = c.userPacketsKey(wallet)
        pipe.Get(ctx, keys[i])
    }
    
    results, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, fmt.Errorf("pipeline exec failed: %w", err)
    }
    
    // Process results
    userPackets := make(map[string][]StuckPacket)
    for i, result := range results {
        if i >= len(walletAddresses) {
            break
        }
        
        wallet := walletAddresses[i]
        
        // Skip if not found
        if result.Err() == redis.Nil {
            continue
        }
        
        // Get string result
        cmd, ok := result.(*redis.StringCmd)
        if !ok {
            continue
        }
        
        data, err := cmd.Bytes()
        if err != nil {
            continue
        }
        
        var packets []StuckPacket
        if err := json.Unmarshal(data, &packets); err != nil {
            c.logger.Error("Failed to unmarshal cached packets",
                zap.String("wallet", wallet),
                zap.Error(err),
            )
            continue
        }
        
        userPackets[wallet] = packets
    }
    
    return userPackets, nil
}

// Key generation helpers
func (c *PacketCache) userPacketsKey(wallet string) string {
    return fmt.Sprintf("packets:user:%s", wallet)
}

func (c *PacketCache) channelStatsKey(channelID string) string {
    return fmt.Sprintf("stats:channel:%s", channelID)
}

func (c *PacketCache) gasEstimateKey(chainID string) string {
    return fmt.Sprintf("gas:chain:%s", chainID)
}
```

### Cache-Aware Service Layer
```go
// Update clearing service to use cache
func (s *Service) GetUserStuckPackets(ctx context.Context, walletAddress string) ([]StuckPacket, error) {
    // Try cache first
    packets, found, err := s.cache.GetUserPackets(ctx, walletAddress)
    if err != nil {
        s.logger.Error("Cache error, falling back to database", zap.Error(err))
    }
    
    if found {
        return packets, nil
    }
    
    // Cache miss - fetch from Chainpulse
    packets, err = s.fetchStuckPacketsFromChainpulse(ctx, walletAddress)
    if err != nil {
        return nil, err
    }
    
    // Cache the results asynchronously
    go func() {
        cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := s.cache.SetUserPackets(cacheCtx, walletAddress, packets); err != nil {
            s.logger.Error("Failed to cache user packets", zap.Error(err))
        }
    }()
    
    return packets, nil
}

// Invalidate cache when packets are cleared
func (s *Service) onPacketsCleared(ctx context.Context, walletAddress string, clearedPackets []string) {
    // Invalidate user's packet cache
    if err := s.cache.InvalidateUserPackets(ctx, walletAddress); err != nil {
        s.logger.Error("Failed to invalidate packet cache",
            zap.String("wallet", walletAddress),
            zap.Error(err),
        )
    }
}
```

## 2. Pagination for Large Results

### Pagination Types
```go
// relayer-middleware/api/pkg/types/pagination.go
package types

type PaginationRequest struct {
    Page     int    `json:"page" form:"page" binding:"min=1"`
    PageSize int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
    SortBy   string `json:"sort_by" form:"sort_by"`
    SortDir  string `json:"sort_dir" form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

type PaginationResponse struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalItems int64 `json:"total_items"`
    TotalPages int   `json:"total_pages"`
    HasNext    bool  `json:"has_next"`
    HasPrev    bool  `json:"has_prev"`
}

func NewPaginationRequest() PaginationRequest {
    return PaginationRequest{
        Page:     1,
        PageSize: 20,
        SortBy:   "created_at",
        SortDir:  "desc",
    }
}

func (p PaginationRequest) Offset() int {
    return (p.Page - 1) * p.PageSize
}

func (p PaginationRequest) Validate() error {
    if p.Page < 1 {
        p.Page = 1
    }
    if p.PageSize < 1 {
        p.PageSize = 20
    }
    if p.PageSize > 100 {
        p.PageSize = 100
    }
    if p.SortDir == "" {
        p.SortDir = "desc"
    }
    return nil
}

func CalculatePaginationResponse(page, pageSize int, totalItems int64) PaginationResponse {
    totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
    
    return PaginationResponse{
        Page:       page,
        PageSize:   pageSize,
        TotalItems: totalItems,
        TotalPages: totalPages,
        HasNext:    page < totalPages,
        HasPrev:    page > 1,
    }
}
```

### Paginated Endpoints
```go
// relayer-middleware/api/pkg/handlers/statistics.go
func (h *StatisticsHandler) GetUserOperations(c *gin.Context) {
    walletAddress := c.Param("wallet")
    
    // Parse pagination
    var pagination types.PaginationRequest
    if err := c.ShouldBindQuery(&pagination); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
        return
    }
    
    // Validate pagination
    if err := pagination.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Get total count
    var totalCount int64
    if err := h.db.Model(&ClearingOperation{}).
        Where("wallet_address = ?", walletAddress).
        Count(&totalCount).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count operations"})
        return
    }
    
    // Get paginated results
    var operations []ClearingOperation
    query := h.db.Where("wallet_address = ?", walletAddress)
    
    // Apply sorting
    sortColumn := h.sanitizeSortColumn(pagination.SortBy)
    query = query.Order(fmt.Sprintf("%s %s", sortColumn, pagination.SortDir))
    
    // Apply pagination
    if err := query.
        Offset(pagination.Offset()).
        Limit(pagination.PageSize).
        Find(&operations).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch operations"})
        return
    }
    
    // Build response
    paginationResp := types.CalculatePaginationResponse(
        pagination.Page,
        pagination.PageSize,
        totalCount,
    )
    
    c.JSON(http.StatusOK, gin.H{
        "operations": operations,
        "pagination": paginationResp,
    })
}

// Prevent SQL injection in ORDER BY
func (h *StatisticsHandler) sanitizeSortColumn(column string) string {
    allowedColumns := map[string]bool{
        "created_at":     true,
        "completed_at":   true,
        "service_fee":    true,
        "packets_cleared": true,
        "success":        true,
    }
    
    if !allowedColumns[column] {
        return "created_at"
    }
    
    return column
}

// Cursor-based pagination for real-time data
func (h *StatisticsHandler) GetStuckPacketsStream(c *gin.Context) {
    walletAddress := c.Param("wallet")
    cursor := c.Query("cursor")
    limit := 50
    
    // Parse limit
    if limitStr := c.Query("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
            limit = l
        }
    }
    
    packets, nextCursor, err := h.getPacketsWithCursor(walletAddress, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch packets"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "packets":     packets,
        "next_cursor": nextCursor,
        "has_more":    nextCursor != "",
    })
}

func (h *StatisticsHandler) getPacketsWithCursor(wallet, cursor string, limit int) ([]StuckPacket, string, error) {
    // Create snapshot timestamp for consistent results
    snapshotTime := time.Now()
    
    query := h.db.Where("wallet_address = ? AND created_at <= ?", wallet, snapshotTime)
    
    // Apply cursor if provided
    if cursor != "" {
        // Decode cursor (e.g., base64 encoded timestamp + ID + snapshot)
        cursorTime, cursorID, cursorSnapshot, err := decodeCursorWithSnapshot(cursor)
        if err != nil {
            return nil, "", err
        }
        
        // Use cursor's snapshot time for consistency
        snapshotTime = cursorSnapshot
        query = h.db.Where("wallet_address = ? AND created_at <= ?", wallet, snapshotTime)
        query = query.Where("(created_at, id) > (?, ?)", cursorTime, cursorID)
    }
    
    // Generate ETag for caching
    etag := h.generateETag(wallet, snapshotTime)
    
    // Check if client has cached version
    if clientETag := h.ctx.GetHeader("If-None-Match"); clientETag == etag {
        h.ctx.Status(http.StatusNotModified)
        return nil, "", nil
    }
    
    // Fetch limit + 1 to determine if there are more results
    var packets []StuckPacket
    if err := query.
        Order("created_at ASC, id ASC").
        Limit(limit + 1).
        Find(&packets).Error; err != nil {
        return nil, "", err
    }
    
    // Set ETag header
    h.ctx.Header("ETag", etag)
    h.ctx.Header("Cache-Control", "private, max-age=60")
    
    // Check if there are more results
    var nextCursor string
    if len(packets) > limit {
        // Remove extra item
        packets = packets[:limit]
        
        // Create cursor from last item with snapshot
        lastPacket := packets[len(packets)-1]
        nextCursor = encodeCursorWithSnapshot(lastPacket.CreatedAt, lastPacket.ID, snapshotTime)
    }
    
    return packets, nextCursor, nil
}

func (h *StatisticsHandler) generateETag(wallet string, snapshot time.Time) string {
    data := fmt.Sprintf("%s:%d", wallet, snapshot.Unix())
    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf(`"%x"`, hash[:8])
}
```

## 3. Database Indexes

### Index Definitions
```sql
-- relayer-middleware/api/migrations/002_add_performance_indexes.sql

-- User query indexes
CREATE INDEX CONCURRENTLY idx_clearing_operations_wallet_created 
ON clearing_operations(wallet_address, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY idx_clearing_operations_wallet_success 
ON clearing_operations(wallet_address, success, created_at DESC)
WHERE deleted_at IS NULL;

-- Token lookup indexes
CREATE INDEX CONCURRENTLY idx_clearing_tokens_token_hash 
ON clearing_tokens(token_hash);

CREATE INDEX CONCURRENTLY idx_clearing_tokens_expires_unused 
ON clearing_tokens(expires_at) 
WHERE used_at IS NULL;

-- Payment verification indexes
CREATE INDEX CONCURRENTLY idx_clearing_operations_payment_tx 
ON clearing_operations(payment_tx_hash)
WHERE payment_tx_hash IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_clearing_operations_clearing_tx 
ON clearing_operations(clearing_tx_hash)
WHERE clearing_tx_hash IS NOT NULL;

-- Statistics aggregation indexes
CREATE INDEX CONCURRENTLY idx_clearing_operations_created_date 
ON clearing_operations(DATE(created_at), success)
WHERE deleted_at IS NULL;

-- Packet tracking indexes (if storing locally)
CREATE INDEX CONCURRENTLY idx_stuck_packets_wallet_status 
ON stuck_packets(wallet_address, status, stuck_since)
WHERE status = 'stuck';

CREATE INDEX CONCURRENTLY idx_stuck_packets_channel_sequence 
ON stuck_packets(source_channel, destination_channel, sequence);

-- Refund processing indexes
CREATE INDEX CONCURRENTLY idx_refunds_status_created 
ON refunds(refund_status, created_at) 
WHERE refund_status = 'pending';

-- Composite indexes for common queries
CREATE INDEX CONCURRENTLY idx_clearing_operations_wallet_date_success 
ON clearing_operations(wallet_address, DATE(created_at), success)
WHERE deleted_at IS NULL;

-- Maintenance script
CREATE OR REPLACE FUNCTION maintain_indexes() RETURNS void AS $$
BEGIN
    -- Update index statistics
    ANALYZE clearing_operations;
    ANALYZE clearing_tokens;
    ANALYZE stuck_packets;
    ANALYZE refunds;
    
    -- Reindex if bloat > 30%
    IF (SELECT pg_relation_size('idx_clearing_operations_wallet_created') > 
        pg_relation_size('clearing_operations') * 0.3) THEN
        REINDEX INDEX CONCURRENTLY idx_clearing_operations_wallet_created;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Schedule maintenance
CREATE EXTENSION IF NOT EXISTS pg_cron;
SELECT cron.schedule('index-maintenance', '0 3 * * 0', 'SELECT maintain_indexes()');
```

### Index Usage Monitoring
```go
// relayer-middleware/api/pkg/database/index_monitor.go
package database

import (
    "context"
    "time"
    
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type IndexMonitor struct {
    db     *gorm.DB
    logger *zap.Logger
}

func NewIndexMonitor(db *gorm.DB) *IndexMonitor {
    return &IndexMonitor{
        db:     db,
        logger: logging.With(zap.String("component", "index_monitor")),
    }
}

func (m *IndexMonitor) CheckIndexUsage(ctx context.Context) {
    // PostgreSQL specific query to check index usage
    query := `
        SELECT 
            schemaname,
            tablename,
            indexname,
            idx_scan,
            idx_tup_read,
            idx_tup_fetch,
            pg_size_pretty(pg_relation_size(indexrelid)) as index_size
        FROM pg_stat_user_indexes
        WHERE schemaname = 'public'
        ORDER BY idx_scan DESC
    `
    
    type IndexStats struct {
        SchemaName   string
        TableName    string
        IndexName    string
        IdxScan      int64
        IdxTupRead   int64
        IdxTupFetch  int64
        IndexSize    string
    }
    
    var stats []IndexStats
    if err := m.db.Raw(query).Scan(&stats).Error; err != nil {
        m.logger.Error("Failed to fetch index stats", zap.Error(err))
        return
    }
    
    // Log unused indexes
    for _, stat := range stats {
        if stat.IdxScan == 0 {
            m.logger.Warn("Unused index detected",
                zap.String("index", stat.IndexName),
                zap.String("table", stat.TableName),
                zap.String("size", stat.IndexSize),
            )
        }
    }
}

func (m *IndexMonitor) CheckSlowQueries(ctx context.Context) {
    // Check for slow queries that might benefit from indexes
    query := `
        SELECT 
            query,
            calls,
            mean_exec_time,
            total_exec_time
        FROM pg_stat_statements
        WHERE mean_exec_time > 100 -- queries taking > 100ms
        ORDER BY mean_exec_time DESC
        LIMIT 10
    `
    
    type SlowQuery struct {
        Query         string
        Calls         int64
        MeanExecTime  float64
        TotalExecTime float64
    }
    
    var slowQueries []SlowQuery
    if err := m.db.Raw(query).Scan(&slowQueries).Error; err != nil {
        // pg_stat_statements might not be enabled
        m.logger.Debug("Could not fetch slow queries", zap.Error(err))
        return
    }
    
    for _, sq := range slowQueries {
        m.logger.Info("Slow query detected",
            zap.String("query", sq.Query),
            zap.Int64("calls", sq.Calls),
            zap.Float64("mean_time_ms", sq.MeanExecTime),
        )
    }
}
```

## 4. WebSocket for Real-time Updates

### WebSocket Manager
```go
// relayer-middleware/api/pkg/websocket/manager.go
package websocket

import (
    "context"
    "encoding/json"
    "sync"
    "time"
    
    "github.com/gorilla/websocket"
    "go.uber.org/zap"
)

type Manager struct {
    clients    map[string]*Client
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
    mu         sync.RWMutex
    logger     *zap.Logger
}

type Client struct {
    id       string
    conn     *websocket.Conn
    send     chan []byte
    topics   map[string]bool
    mu       sync.RWMutex
}

type Message struct {
    Topic string      `json:"topic"`
    Event string      `json:"event"`
    Data  interface{} `json:"data"`
}

func NewManager() *Manager {
    return &Manager{
        clients:    make(map[string]*Client),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan Message, 100),
        logger:     logging.With(zap.String("component", "websocket")),
    }
}

func (m *Manager) Run(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            m.closeAllClients()
            return
            
        case client := <-m.register:
            m.mu.Lock()
            m.clients[client.id] = client
            m.mu.Unlock()
            m.logger.Info("Client connected", zap.String("id", client.id))
            
        case client := <-m.unregister:
            m.mu.Lock()
            if _, ok := m.clients[client.id]; ok {
                delete(m.clients, client.id)
                close(client.send)
                m.mu.Unlock()
                m.logger.Info("Client disconnected", zap.String("id", client.id))
            } else {
                m.mu.Unlock()
            }
            
        case message := <-m.broadcast:
            m.broadcastToTopic(message)
            
        case <-ticker.C:
            m.pingClients()
        }
    }
}

type ConnectionLimiter struct {
    mu           sync.Mutex
    connections  map[string]int
    maxPerIP     int
    window       time.Duration
    lastCleanup  time.Time
}

func NewConnectionLimiter(maxPerIP int) *ConnectionLimiter {
    return &ConnectionLimiter{
        connections: make(map[string]int),
        maxPerIP:    maxPerIP,
        window:      1 * time.Hour,
        lastCleanup: time.Now(),
    }
}

func (cl *ConnectionLimiter) AllowConnection(ip string) bool {
    cl.mu.Lock()
    defer cl.mu.Unlock()
    
    // Cleanup old entries periodically
    if time.Since(cl.lastCleanup) > cl.window {
        cl.connections = make(map[string]int)
        cl.lastCleanup = time.Now()
    }
    
    current := cl.connections[ip]
    if current >= cl.maxPerIP {
        return false
    }
    
    cl.connections[ip]++
    return true
}

func (cl *ConnectionLimiter) ReleaseConnection(ip string) {
    cl.mu.Lock()
    defer cl.mu.Unlock()
    
    if count, ok := cl.connections[ip]; ok && count > 0 {
        cl.connections[ip]--
        if cl.connections[ip] == 0 {
            delete(cl.connections, ip)
        }
    }
}

func (m *Manager) broadcastToTopic(message Message) {
    data, err := json.Marshal(message)
    if err != nil {
        m.logger.Error("Failed to marshal message", zap.Error(err))
        return
    }
    
    m.mu.RLock()
    subscribers := make([]*Client, 0)
    for _, client := range m.clients {
        if client.isSubscribed(message.Topic) {
            subscribers = append(subscribers, client)
        }
    }
    m.mu.RUnlock()
    
    // Broadcast without holding lock
    for _, client := range subscribers {
        select {
        case client.send <- data:
        default:
            // Client's send channel is full
            m.logger.Warn("Client send buffer full", 
                zap.String("id", client.id),
                zap.String("topic", message.Topic),
            )
            
            // Mark client for disconnection
            go func(c *Client) {
                m.unregister <- c
            }(client)
        }
    }
}

// Client-side reconnection with exponential backoff
class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.reconnectDelay = 1000; // Start with 1 second
        this.maxReconnectDelay = 30000; // Max 30 seconds
        this.reconnectAttempts = 0;
        this.shouldReconnect = true;
    }
    
    connect() {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectDelay = 1000; // Reset delay
            this.reconnectAttempts = 0;
        };
        
        this.ws.onclose = (event) => {
            if (this.shouldReconnect && !event.wasClean) {
                this.scheduleReconnect();
            }
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    scheduleReconnect() {
        this.reconnectAttempts++;
        
        // Calculate delay with exponential backoff and jitter
        let delay = Math.min(
            this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
            this.maxReconnectDelay
        );
        
        // Add jitter (±25%)
        const jitter = delay * 0.25;
        delay = delay + (Math.random() * 2 - 1) * jitter;
        
        console.log(`Reconnecting in ${Math.round(delay)}ms (attempt ${this.reconnectAttempts})`);
        
        setTimeout(() => {
            this.connect();
        }, delay);
    }
    
    disconnect() {
        this.shouldReconnect = false;
        if (this.ws) {
            this.ws.close();
        }
    }
}

func (m *Manager) pingClients() {
    m.mu.RLock()
    clients := make([]*Client, 0, len(m.clients))
    for _, client := range m.clients {
        clients = append(clients, client)
    }
    m.mu.RUnlock()
    
    for _, client := range clients {
        if err := client.ping(); err != nil {
            m.unregister <- client
        }
    }
}

// Client methods
func (c *Client) isSubscribed(topic string) bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.topics[topic]
}

func (c *Client) subscribe(topic string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.topics[topic] = true
}

func (c *Client) unsubscribe(topic string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.topics, topic)
}

func (c *Client) ping() error {
    c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
    return c.conn.WriteMessage(websocket.PingMessage, nil)
}
```

### WebSocket Handler
```go
// relayer-middleware/api/pkg/handlers/websocket.go
package handlers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // Configure based on environment
        return true
    },
}

type WebSocketHandler struct {
    manager *websocket.Manager
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
        return
    }
    
    client := &Client{
        id:     uuid.New().String(),
        conn:   conn,
        send:   make(chan []byte, 256),
        topics: make(map[string]bool),
    }
    
    h.manager.register <- client
    
    // Start goroutines for reading and writing
    go client.writePump()
    go client.readPump(h.manager)
}

func (c *Client) readPump(manager *websocket.Manager) {
    defer func() {
        manager.unregister <- c
        c.conn.Close()
    }()
    
    c.conn.SetReadLimit(512)
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        var msg ClientMessage
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                manager.logger.Error("WebSocket error", zap.Error(err))
            }
            break
        }
        
        // Handle client messages
        switch msg.Type {
        case "subscribe":
            if topic, ok := msg.Data["topic"].(string); ok {
                c.subscribe(topic)
                c.sendMessage("subscribed", map[string]interface{}{
                    "topic": topic,
                })
            }
            
        case "unsubscribe":
            if topic, ok := msg.Data["topic"].(string); ok {
                c.unsubscribe(topic)
                c.sendMessage("unsubscribed", map[string]interface{}{
                    "topic": topic,
                })
            }
        }
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }
            
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

### Integration with Clearing Service
```go
// Broadcast clearing status updates
func (s *Service) broadcastClearingUpdate(token string, status ClearingStatus) {
    message := websocket.Message{
        Topic: fmt.Sprintf("clearing:%s", token),
        Event: "status_update",
        Data: map[string]interface{}{
            "status":      status.Status,
            "message":     status.Message,
            "progress":    status.Progress,
            "updated_at":  status.UpdatedAt,
            "tx_hashes":   status.TxHashes,
        },
    }
    
    s.wsManager.broadcast <- message
}

// Broadcast platform statistics updates
func (s *StatisticsService) broadcastPlatformStats() {
    stats, err := s.getPlatformStatistics()
    if err != nil {
        s.logger.Error("Failed to get platform stats", zap.Error(err))
        return
    }
    
    message := websocket.Message{
        Topic: "platform:stats",
        Event: "update",
        Data:  stats,
    }
    
    s.wsManager.broadcast <- message
}
```

## Testing Considerations

1. **Cache Tests**
   - Test cache hit/miss scenarios
   - Verify TTL expiration
   - Test cache invalidation
   - Check concurrent access

2. **Pagination Tests**
   - Test boundary conditions
   - Verify cursor encoding/decoding
   - Test large result sets
   - Check sort order consistency

3. **Index Tests**
   - Verify query performance improvements
   - Test index usage with EXPLAIN
   - Check index maintenance overhead
   - Monitor index bloat

4. **WebSocket Tests**
   - Test connection handling
   - Verify message broadcasting
   - Test subscription management
   - Check reconnection logic

5. **Load Tests**
   - Test cache under high load
   - Verify pagination performance
   - Test WebSocket with many clients
   - Check database connection pooling