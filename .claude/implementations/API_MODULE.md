# API Module Blueprint

## Module Overview

The API module is the core backend service for Relayooor, built with Go and the Gin framework. It handles business logic, authentication, packet clearing orchestration, and integration with external services.

## Architecture

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Gin HTTP framework
- **Database**: PostgreSQL with pgx/v5
- **Cache**: Redis
- **Authentication**: JWT with HMAC signing
- **WebSocket**: Gorilla WebSocket
- **Testing**: Go testing package + testify
- **Metrics**: Prometheus

### Directory Structure
```
relayer-middleware/api/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── pkg/
│   ├── clearing/            # Packet clearing business logic
│   │   ├── service.go      # Main clearing service
│   │   ├── token.go        # Token generation
│   │   ├── validator.go    # Request validation
│   │   └── executor.go     # Hermes integration
│   ├── handlers/           # HTTP request handlers
│   │   ├── auth.go         # Authentication endpoints
│   │   ├── clearing.go     # Clearing endpoints
│   │   ├── monitoring.go   # Monitoring endpoints
│   │   ├── packets.go      # Packet query endpoints
│   │   └── websocket.go    # WebSocket handler
│   ├── middleware/         # HTTP middleware
│   │   ├── auth.go         # JWT validation
│   │   ├── cors.go         # CORS configuration
│   │   ├── ratelimit.go    # Rate limiting
│   │   └── logging.go      # Request logging
│   ├── database/           # Database layer
│   │   ├── connection.go   # Connection pooling
│   │   ├── queries.go      # SQL queries
│   │   └── models.go       # Data models
│   ├── cache/             # Redis caching layer
│   │   ├── client.go      # Redis client
│   │   └── operations.go  # Cache operations
│   ├── services/          # External service integrations
│   │   ├── chainpulse.go  # Chainpulse client
│   │   ├── hermes.go      # Hermes client
│   │   └── pricing.go     # Price feed client
│   └── utils/             # Utility functions
│       ├── crypto.go      # Cryptographic functions
│       ├── validation.go  # Input validation
│       └── errors.go      # Error handling
├── migrations/            # Database migrations
├── config/               # Configuration files
└── tests/               # Integration tests
```

## Core Components

### 1. Server Initialization
```go
func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize database
    db, err := database.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Run migrations
    if err := database.Migrate(db); err != nil {
        log.Fatal("Failed to run migrations:", err)
    }
    
    // Initialize Redis
    cache := cache.NewClient(cfg.RedisURL)
    defer cache.Close()
    
    // Initialize services
    chainpulse := services.NewChainpulseClient(cfg.ChainpulseURL)
    hermes := services.NewHermesClient(cfg.HermesURL)
    clearing := clearing.NewService(db, cache, hermes)
    
    // Setup router
    router := setupRouter(db, cache, clearing, chainpulse)
    
    // Start server
    log.Printf("Starting server on port %s", cfg.Port)
    if err := router.Run(":" + cfg.Port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

### 2. Router Configuration
```go
func setupRouter(deps Dependencies) *gin.Engine {
    router := gin.New()
    
    // Global middleware
    router.Use(middleware.Logger())
    router.Use(middleware.Recovery())
    router.Use(middleware.CORS())
    router.Use(middleware.RateLimit())
    
    // Health check
    router.GET("/health", handlers.Health)
    
    // API routes
    api := router.Group("/api")
    {
        // Public routes
        api.GET("/chains/:id/packets/stuck", handlers.GetStuckPackets)
        api.GET("/help/glossary", handlers.GetGlossary)
        
        // Authentication
        auth := api.Group("/auth")
        {
            auth.POST("/wallet-sign", handlers.WalletAuth)
            auth.POST("/refresh", handlers.RefreshToken)
        }
        
        // Protected routes
        protected := api.Group("/")
        protected.Use(middleware.RequireAuth())
        {
            // Clearing service
            clearing := protected.Group("/clearing")
            {
                clearing.POST("/request", handlers.RequestClearing)
                clearing.POST("/verify-payment", handlers.VerifyPayment)
                clearing.GET("/status/:token", handlers.GetClearingStatus)
                clearing.POST("/execute", handlers.ExecuteClearing)
                clearing.GET("/history", handlers.GetClearingHistory)
            }
            
            // User management
            users := protected.Group("/users")
            {
                users.GET("/profile", handlers.GetProfile)
                users.GET("/statistics", handlers.GetUserStats)
                users.GET("/transactions", handlers.GetTransactions)
            }
        }
        
        // Monitoring
        api.GET("/monitoring/data", handlers.GetMonitoringData)
        api.GET("/statistics/platform", handlers.GetPlatformStats)
        
        // WebSocket
        api.GET("/ws", handlers.WebSocketHandler)
    }
    
    return router
}
```

## Authentication System

### JWT Implementation
```go
type Claims struct {
    UserID        string `json:"user_id"`
    WalletAddress string `json:"wallet_address"`
    ChainID       string `json:"chain_id"`
    jwt.StandardClaims
}

func GenerateToken(user *models.User) (string, error) {
    claims := &Claims{
        UserID:        user.ID,
        WalletAddress: user.WalletAddress,
        ChainID:       user.ChainID,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            IssuedAt:  time.Now().Unix(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.JWTSecret))
}
```

### Wallet Authentication
```go
func WalletAuth(c *gin.Context) {
    var req WalletAuthRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Verify signature
    if !crypto.VerifyWalletSignature(req.Message, req.Signature, req.Address) {
        c.JSON(401, gin.H{"error": "Invalid signature"})
        return
    }
    
    // Create or update user
    user, err := db.CreateOrUpdateUser(req.Address, req.ChainID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create user"})
        return
    }
    
    // Generate token
    token, err := GenerateToken(user)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to generate token"})
        return
    }
    
    c.JSON(200, gin.H{
        "token": token,
        "user":  user,
    })
}
```

## Packet Clearing Service

### Token Generation
```go
type ClearingToken struct {
    ID          string
    PacketID    string
    UserID      string
    Amount      string
    PaymentMemo string
    ExpiresAt   time.Time
    Signature   string
}

func (s *ClearingService) GenerateToken(packet *models.Packet, user *models.User) (*ClearingToken, error) {
    // Verify ownership
    if !s.verifyPacketOwnership(packet, user) {
        return nil, ErrUnauthorized
    }
    
    // Calculate fee
    fee := s.calculateFee(packet)
    
    // Generate unique token
    token := &ClearingToken{
        ID:          uuid.New().String(),
        PacketID:    packet.ID,
        UserID:      user.ID,
        Amount:      fee.String(),
        PaymentMemo: generatePaymentMemo(),
        ExpiresAt:   time.Now().Add(30 * time.Minute),
    }
    
    // Sign token
    token.Signature = s.signToken(token)
    
    // Store in cache
    if err := s.cache.Set(token.ID, token, 30*time.Minute); err != nil {
        return nil, err
    }
    
    return token, nil
}
```

### Payment Verification
```go
func (s *ClearingService) VerifyPayment(tokenID string, txHash string) error {
    // Get token from cache
    token, err := s.cache.Get(tokenID)
    if err != nil {
        return ErrTokenNotFound
    }
    
    // Verify token hasn't expired
    if time.Now().After(token.ExpiresAt) {
        return ErrTokenExpired
    }
    
    // Query blockchain for transaction
    tx, err := s.queryTransaction(txHash)
    if err != nil {
        return err
    }
    
    // Verify memo matches
    if tx.Memo != token.PaymentMemo {
        return ErrInvalidMemo
    }
    
    // Verify amount
    if tx.Amount.LT(token.Amount) {
        return ErrInsufficientPayment
    }
    
    // Update token status
    token.PaymentVerified = true
    token.TxHash = txHash
    
    return s.cache.Set(tokenID, token, 30*time.Minute)
}
```

### Hermes Integration
```go
type HermesClient struct {
    baseURL    string
    httpClient *http.Client
}

func (h *HermesClient) ClearPacket(packet *models.Packet) error {
    // Construct clear packet request
    req := map[string]interface{}{
        "chain_id":    packet.SrcChainID,
        "channel_id":  packet.SrcChannelID,
        "port_id":     packet.SrcPortID,
        "sequence":    packet.Sequence,
    }
    
    // Send request to Hermes
    resp, err := h.httpClient.Post(
        h.baseURL+"/clear",
        "application/json",
        bytes.NewBuffer(jsonEncode(req)),
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
```

## Database Layer

### Connection Pool
```go
func Connect(dsn string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, err
    }
    
    // Configure pool
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = time.Minute * 30
    
    // Connect
    pool, err := pgxpool.ConnectConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }
    
    // Test connection
    if err := pool.Ping(context.Background()); err != nil {
        return nil, err
    }
    
    return pool, nil
}
```

### Query Examples
```go
func GetStuckPackets(chainID string, minStuckDuration time.Duration) ([]*Packet, error) {
    query := `
        SELECT 
            p.id, p.src_chain_id, p.dst_chain_id,
            p.src_channel_id, p.dst_channel_id,
            p.sequence, p.timeout_timestamp,
            p.created_at, p.data
        FROM packets p
        WHERE p.status = 'pending'
            AND p.src_chain_id = $1
            AND p.created_at < $2
        ORDER BY p.created_at ASC
        LIMIT 100
    `
    
    cutoff := time.Now().Add(-minStuckDuration)
    rows, err := db.Query(context.Background(), query, chainID, cutoff)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var packets []*Packet
    for rows.Next() {
        var p Packet
        if err := rows.Scan(&p.ID, &p.SrcChainID, ...); err != nil {
            return nil, err
        }
        packets = append(packets, &p)
    }
    
    return packets, nil
}
```

## WebSocket Implementation

### Connection Manager
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            log.Printf("Client connected: %s", client.id)
            
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                log.Printf("Client disconnected: %s", client.id)
            }
            
        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}
```

### Message Broadcasting
```go
func BroadcastPacketUpdate(packet *Packet) {
    message := map[string]interface{}{
        "type": "packet_update",
        "data": packet,
    }
    
    jsonData, _ := json.Marshal(message)
    hub.broadcast <- jsonData
}
```

## Monitoring & Metrics

### Prometheus Metrics
```go
var (
    packetsCleared = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "relayooor_packets_cleared_total",
            Help: "Total number of packets cleared",
        },
        []string{"chain", "status"},
    )
    
    clearingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "relayooor_clearing_duration_seconds",
            Help:    "Time taken to clear a packet",
            Buckets: prometheus.DefBuckets,
        },
        []string{"chain"},
    )
    
    apiRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "relayooor_api_request_duration_seconds",
            Help:    "API request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
)

func init() {
    prometheus.MustRegister(packetsCleared)
    prometheus.MustRegister(clearingDuration)
    prometheus.MustRegister(apiRequestDuration)
}
```

## Error Handling

### Custom Error Types
```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

var (
    ErrUnauthorized        = &APIError{Code: "UNAUTHORIZED", Message: "Unauthorized access"}
    ErrPacketNotFound      = &APIError{Code: "PACKET_NOT_FOUND", Message: "Packet not found"}
    ErrTokenExpired        = &APIError{Code: "TOKEN_EXPIRED", Message: "Token has expired"}
    ErrInvalidPayment      = &APIError{Code: "INVALID_PAYMENT", Message: "Invalid payment"}
    ErrClearingFailed      = &APIError{Code: "CLEARING_FAILED", Message: "Failed to clear packet"}
)
```

### Error Middleware
```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            switch e := err.Err.(type) {
            case *APIError:
                c.JSON(400, e)
            case validation.Errors:
                c.JSON(400, &APIError{
                    Code:    "VALIDATION_ERROR",
                    Message: "Validation failed",
                    Details: e,
                })
            default:
                c.JSON(500, &APIError{
                    Code:    "INTERNAL_ERROR",
                    Message: "Internal server error",
                })
            }
        }
    }
}
```

## Testing

### Unit Tests
```go
func TestClearingService_GenerateToken(t *testing.T) {
    service := setupTestService(t)
    packet := createTestPacket()
    user := createTestUser()
    
    token, err := service.GenerateToken(packet, user)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, token.ID)
    assert.Equal(t, packet.ID, token.PacketID)
    assert.Equal(t, user.ID, token.UserID)
    assert.NotEmpty(t, token.Signature)
}
```

### Integration Tests
```go
func TestAPI_ClearingFlow(t *testing.T) {
    router := setupTestRouter()
    
    // 1. Request clearing token
    resp := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/clearing/request", 
        strings.NewReader(`{"packet_id": "test-packet"}`))
    req.Header.Set("Authorization", "Bearer "+testToken)
    router.ServeHTTP(resp, req)
    
    assert.Equal(t, 200, resp.Code)
    
    var tokenResp map[string]interface{}
    json.Unmarshal(resp.Body.Bytes(), &tokenResp)
    tokenID := tokenResp["token_id"].(string)
    
    // 2. Verify payment
    resp = httptest.NewRecorder()
    req, _ = http.NewRequest("POST", "/api/clearing/verify-payment",
        strings.NewReader(fmt.Sprintf(`{"token_id": "%s", "tx_hash": "test-tx"}`, tokenID)))
    req.Header.Set("Authorization", "Bearer "+testToken)
    router.ServeHTTP(resp, req)
    
    assert.Equal(t, 200, resp.Code)
    
    // 3. Execute clearing
    resp = httptest.NewRecorder()
    req, _ = http.NewRequest("POST", "/api/clearing/execute",
        strings.NewReader(fmt.Sprintf(`{"token_id": "%s"}`, tokenID)))
    req.Header.Set("Authorization", "Bearer "+testToken)
    router.ServeHTTP(resp, req)
    
    assert.Equal(t, 200, resp.Code)
}
```

## Configuration

### Environment Variables
```env
# Server
PORT=8080
ENV=production

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/relayooor
DATABASE_MAX_CONNECTIONS=25

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=secret

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# External Services
CHAINPULSE_URL=http://chainpulse:3000
HERMES_URL=http://hermes:5185

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
```

## Deployment

### Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./api"]
```

### Health Checks
```go
func Health(c *gin.Context) {
    checks := map[string]string{
        "database": checkDatabase(),
        "redis":    checkRedis(),
        "chainpulse": checkChainpulse(),
        "hermes":   checkHermes(),
    }
    
    healthy := true
    for _, status := range checks {
        if status != "healthy" {
            healthy = false
            break
        }
    }
    
    statusCode := 200
    if !healthy {
        statusCode = 503
    }
    
    c.JSON(statusCode, gin.H{
        "status": healthy,
        "checks": checks,
        "timestamp": time.Now().Unix(),
    })
}
```