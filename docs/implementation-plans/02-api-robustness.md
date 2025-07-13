# API Robustness Implementation Plan

## 1. Health Check Endpoint

### Implementation
```go
// relayer-middleware/api/pkg/handlers/health.go
package handlers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
)

type HealthStatus struct {
    Status    string                 `json:"status"`
    Timestamp time.Time             `json:"timestamp"`
    Version   string                `json:"version"`
    Checks    map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
    Status  string        `json:"status"`
    Latency time.Duration `json:"latency_ms"`
    Error   string        `json:"error,omitempty"`
}

type HealthHandler struct {
    db     *gorm.DB
    redis  *redis.Client
    hermes HermesClient
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
    health := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now().UTC(),
        Version:   getVersion(),
        Checks:    make(map[string]CheckResult),
    }
    
    // Check database
    dbCheck := h.checkDatabase()
    health.Checks["database"] = dbCheck
    
    // Check Redis
    redisCheck := h.checkRedis()
    health.Checks["redis"] = redisCheck
    
    // Check Hermes
    hermesCheck := h.checkHermes()
    health.Checks["hermes"] = hermesCheck
    
    // Determine overall status with degraded mode support
    criticalFailure := false
    degradedServices := []string{}
    
    // Database is critical
    if dbCheck.Status != "healthy" {
        criticalFailure = true
    }
    
    // Redis is non-critical (can use fallbacks)
    if redisCheck.Status != "healthy" {
        degradedServices = append(degradedServices, "redis")
    }
    
    // Hermes is critical for clearing
    if hermesCheck.Status != "healthy" {
        criticalFailure = true
    }
    
    if criticalFailure {
        health.Status = "unhealthy"
        c.JSON(http.StatusServiceUnavailable, health)
    } else if len(degradedServices) > 0 {
        health.Status = "degraded"
        health.DegradedServices = degradedServices
        c.JSON(http.StatusOK, health) // Still return 200 for degraded
    } else {
        c.JSON(http.StatusOK, health)
    }
}

func (h *HealthHandler) GetReadiness(c *gin.Context) {
    // Lighter check for k8s readiness probe
    if err := h.db.Exec("SELECT 1").Error; err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not ready",
            "error":  "database unavailable",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *HealthHandler) GetLiveness(c *gin.Context) {
    // Simple liveness check for k8s
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
```

### Route Registration
```go
// relayer-middleware/api/cmd/server/main.go
// Add to route setup
health := handlers.NewHealthHandler(db, redisClient, hermesClient)
router.GET("/health", health.GetHealth)
router.GET("/health/ready", health.GetReadiness)
router.GET("/health/live", health.GetLiveness)
```

## 2. Structured Logging

### Logger Setup
```go
// relayer-middleware/api/pkg/logging/logger.go
package logging

import (
    "os"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(env string) error {
    var config zap.Config
    
    if env == "production" {
        config = zap.NewProductionConfig()
        config.EncoderConfig.TimeKey = "timestamp"
        config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    
    // Set log level from environment
    logLevel := os.Getenv("LOG_LEVEL")
    if logLevel != "" {
        var level zapcore.Level
        if err := level.UnmarshalText([]byte(logLevel)); err == nil {
            config.Level = zap.NewAtomicLevelAt(level)
        }
    }
    
    logger, err := config.Build(
        zap.AddCaller(),
        zap.AddStacktrace(zapcore.ErrorLevel),
    )
    if err != nil {
        return err
    }
    
    Logger = logger
    return nil
}

// Helper functions for structured logging
func With(fields ...zap.Field) *zap.Logger {
    return Logger.With(fields...)
}

func Info(msg string, fields ...zap.Field) {
    Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
    Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
    Logger.Fatal(msg, fields...)
}
```

### Logging Middleware
```go
// relayer-middleware/api/pkg/middleware/logging.go
package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "relayooor/api/pkg/logging"
)

// Sensitive fields that should not be logged
var sensitiveFields = map[string]bool{
    "password":     true,
    "private_key":  true,
    "mnemonic":     true,
    "secret":       true,
    "token":        true,
    "signature":    true,
}

func sanitizePath(path string) string {
    // Remove potential secrets from URLs
    if strings.Contains(path, "/tokens/") {
        // Replace token IDs with placeholder
        re := regexp.MustCompile(`/tokens/[^/]+`)
        path = re.ReplaceAllString(path, "/tokens/[REDACTED]")
    }
    return path
}

func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := sanitizePath(c.Request.URL.Path)
        raw := c.Request.URL.RawQuery
        
        // Process request
        c.Next()
        
        // Log request details
        latency := time.Since(start)
        clientIP := c.ClientIP()
        method := c.Request.Method
        statusCode := c.Writer.Status()
        errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
        
        if raw != "" {
            // Sanitize query parameters
            sanitizedQuery := sanitizeQueryParams(raw)
            path = path + "?" + sanitizedQuery
        }
        
        fields := []zap.Field{
            zap.String("client_ip", clientIP),
            zap.String("method", method),
            zap.String("path", path),
            zap.Int("status", statusCode),
            zap.Duration("latency", latency),
            zap.String("user_agent", c.Request.UserAgent()),
        }
        
        if errorMessage != "" {
            // Sanitize error messages
            fields = append(fields, zap.String("error", sanitizeError(errorMessage)))
        }
        
        // Add request ID if present
        if requestID := c.GetString("request_id"); requestID != "" {
            fields = append(fields, zap.String("request_id", requestID))
        }
        
        // Add user context if authenticated (hash wallet for privacy)
        if walletAddress := c.GetString("wallet_address"); walletAddress != "" {
            fields = append(fields, zap.String("wallet_hash", hashWallet(walletAddress)))
        }
        
        switch {
        case statusCode >= 500:
            logging.Error("Server error", fields...)
        case statusCode >= 400:
            logging.Warn("Client error", fields...)
        case statusCode >= 300:
            logging.Info("Redirection", fields...)
        default:
            logging.Info("Request completed", fields...)
        }
    }
}

func sanitizeQueryParams(query string) string {
    params, _ := url.ParseQuery(query)
    for key := range params {
        if sensitiveFields[strings.ToLower(key)] {
            params[key] = []string{"[REDACTED]"}
        }
    }
    return params.Encode()
}
```

### Integration in Services
```go
// Update clearing service with structured logging
func (s *Service) GenerateToken(ctx context.Context, request ClearingRequest) (*TokenResponse, error) {
    logger := logging.With(
        zap.String("operation", "generate_token"),
        zap.String("wallet", request.WalletAddress),
        zap.Int("packet_count", len(request.PacketIdentifiers)),
    )
    
    logger.Info("Generating clearing token")
    
    // ... existing logic ...
    
    if err != nil {
        logger.Error("Failed to generate token", zap.Error(err))
        return nil, err
    }
    
    logger.Info("Token generated successfully", 
        zap.String("token_id", token.Token),
        zap.Time("expires_at", token.ExpiresAt),
    )
    
    return response, nil
}
```

## 3. Database Connection Pooling

### Database Configuration
```go
// relayer-middleware/api/pkg/database/config.go
package database

import (
    "fmt"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Config struct {
    Host            string
    Port            int
    User            string
    Password        string
    Database        string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewConnection(cfg Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
        PrepareStmt: true, // Prepare statements for better performance
    })
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }
    
    // Configure connection pool with auto-scaling
    baseConns := cfg.MaxOpenConns
    sqlDB.SetMaxOpenConns(baseConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
    
    // Start connection pool monitor
    go monitorAndScalePool(sqlDB, baseConns)
}

func monitorAndScalePool(db *sql.DB, baseConns int) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    consecutiveHighLoad := 0
    consecutiveLowLoad := 0
    
    for range ticker.C {
        stats := db.Stats()
        utilization := float64(stats.InUse) / float64(stats.MaxOpenConnections)
        
        if utilization > 0.8 && stats.WaitCount > 0 {
            consecutiveHighLoad++
            consecutiveLowLoad = 0
            
            if consecutiveHighLoad >= 3 {
                newMax := int(float64(stats.MaxOpenConnections) * 1.5)
                if newMax > baseConns*3 {
                    newMax = baseConns * 3 // Cap at 3x base
                }
                
                db.SetMaxOpenConns(newMax)
                logging.Info("Scaled up connection pool",
                    zap.Int("new_max", newMax),
                    zap.Float64("utilization", utilization))
                consecutiveHighLoad = 0
            }
        } else if utilization < 0.3 {
            consecutiveLowLoad++
            consecutiveHighLoad = 0
            
            if consecutiveLowLoad >= 5 {
                newMax := int(float64(stats.MaxOpenConnections) * 0.7)
                if newMax < baseConns {
                    newMax = baseConns // Don't go below base
                }
                
                db.SetMaxOpenConns(newMax)
                logging.Info("Scaled down connection pool",
                    zap.Int("new_max", newMax),
                    zap.Float64("utilization", utilization))
                consecutiveLowLoad = 0
            }
        } else {
            consecutiveHighLoad = 0
            consecutiveLowLoad = 0
        }
    }
    
    // Test connection
    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return db, nil
}

func DefaultConfig() Config {
    return Config{
        Host:            getEnvOrDefault("DB_HOST", "localhost"),
        Port:            getEnvAsIntOrDefault("DB_PORT", 5432),
        User:            getEnvOrDefault("DB_USER", "postgres"),
        Password:        getEnvOrDefault("DB_PASSWORD", ""),
        Database:        getEnvOrDefault("DB_NAME", "relayooor"),
        SSLMode:         getEnvOrDefault("DB_SSLMODE", "disable"),
        MaxOpenConns:    getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 25),
        MaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 5),
        ConnMaxLifetime: getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
        ConnMaxIdleTime: getEnvAsDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
    }
}
```

### Connection Monitoring
```go
// relayer-middleware/api/pkg/database/monitor.go
package database

import (
    "context"
    "database/sql"
    "time"
    
    "go.uber.org/zap"
    "relayooor/api/pkg/logging"
)

type ConnectionMonitor struct {
    db       *sql.DB
    interval time.Duration
    logger   *zap.Logger
}

func NewConnectionMonitor(db *sql.DB, interval time.Duration) *ConnectionMonitor {
    return &ConnectionMonitor{
        db:       db,
        interval: interval,
        logger:   logging.With(zap.String("component", "db_monitor")),
    }
}

func (m *ConnectionMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            m.logger.Info("Stopping connection monitor")
            return
        case <-ticker.C:
            stats := m.db.Stats()
            m.logger.Info("Database connection stats",
                zap.Int("open_connections", stats.OpenConnections),
                zap.Int("in_use", stats.InUse),
                zap.Int("idle", stats.Idle),
                zap.Int64("wait_count", stats.WaitCount),
                zap.Duration("wait_duration", stats.WaitDuration),
                zap.Int64("max_idle_closed", stats.MaxIdleClosed),
                zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
            )
            
            // Alert if connection pool is exhausted
            if stats.OpenConnections == stats.MaxOpenConnections && stats.WaitCount > 0 {
                m.logger.Warn("Database connection pool exhausted",
                    zap.Int64("waiting_connections", stats.WaitCount),
                )
            }
        }
    }
}
```

## 4. Graceful Shutdown

### Shutdown Handler
```go
// relayer-middleware/api/pkg/server/shutdown.go
package server

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "go.uber.org/zap"
    "relayooor/api/pkg/logging"
)

type GracefulServer struct {
    server          *http.Server
    shutdownTimeout time.Duration
    logger          *zap.Logger
    cleanup         []func() error
}

func NewGracefulServer(server *http.Server, timeout time.Duration) *GracefulServer {
    return &GracefulServer{
        server:          server,
        shutdownTimeout: timeout,
        logger:          logging.With(zap.String("component", "server")),
        cleanup:         make([]func() error, 0),
    }
}

func (s *GracefulServer) RegisterCleanup(fn func() error) {
    s.cleanup = append(s.cleanup, fn)
}

func (s *GracefulServer) ListenAndServe() error {
    // Channel to listen for interrupt signals
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
    
    // Channel to track server errors
    serverErrors := make(chan error, 1)
    
    // Start server in goroutine
    go func() {
        s.logger.Info("Starting server", zap.String("addr", s.server.Addr))
        if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            serverErrors <- err
        }
    }()
    
    // Wait for interrupt signal or server error
    select {
    case err := <-serverErrors:
        return err
    case sig := <-stop:
        s.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
        return s.shutdown()
    }
}

type ActiveOperation struct {
    ID        string
    StartTime time.Time
    Type      string
}

type OperationTracker struct {
    mu         sync.RWMutex
    operations map[string]*ActiveOperation
}

func (s *GracefulServer) shutdown() error {
    s.logger.Info("Starting graceful shutdown", 
        zap.Duration("timeout", s.shutdownTimeout))
    
    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
    defer cancel()
    
    // Wait for active clearing operations (max 5 minutes)
    if err := s.waitForActiveOperations(ctx); err != nil {
        s.logger.Warn("Some operations did not complete", zap.Error(err))
        // Persist state for recovery
        s.persistActiveOperations()
    }
    
    // Shutdown HTTP server
    if err := s.server.Shutdown(ctx); err != nil {
        s.logger.Error("Failed to shutdown server gracefully", zap.Error(err))
        return err
    }
    
    // Run cleanup functions
    s.logger.Info("Running cleanup tasks", zap.Int("count", len(s.cleanup)))
    for i, cleanupFn := range s.cleanup {
        s.logger.Debug("Running cleanup task", zap.Int("index", i))
        if err := cleanupFn(); err != nil {
            s.logger.Error("Cleanup task failed", 
                zap.Int("index", i), 
                zap.Error(err))
        }
    }
    
    s.logger.Info("Graceful shutdown completed")
    return nil
}

func (s *GracefulServer) waitForActiveOperations(ctx context.Context) error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    maxWait := 5 * time.Minute
    deadline := time.Now().Add(maxWait)
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            count := s.tracker.Count()
            if count == 0 {
                return nil
            }
            
            if time.Now().After(deadline) {
                return fmt.Errorf("%d operations still active after %v", count, maxWait)
            }
            
            s.logger.Info("Waiting for active operations", 
                zap.Int("count", count),
                zap.Duration("remaining", time.Until(deadline)))
        }
    }
}
```

### Integration in Main
```go
// relayer-middleware/api/cmd/server/main.go
func main() {
    // ... initialization ...
    
    // Create HTTP server
    httpServer := &http.Server{
        Addr:         fmt.Sprintf(":%d", config.Port),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Create graceful server
    gracefulServer := server.NewGracefulServer(httpServer, 30*time.Second)
    
    // Register cleanup functions
    gracefulServer.RegisterCleanup(func() error {
        logging.Info("Closing database connections")
        sqlDB, _ := db.DB()
        return sqlDB.Close()
    })
    
    gracefulServer.RegisterCleanup(func() error {
        logging.Info("Closing Redis connections")
        return redisClient.Close()
    })
    
    gracefulServer.RegisterCleanup(func() error {
        logging.Info("Stopping background workers")
        cancelWorkers() // Cancel context for background workers
        return nil
    })
    
    gracefulServer.RegisterCleanup(func() error {
        logging.Info("Flushing logs")
        return logging.Logger.Sync()
    })
    
    // Start server with graceful shutdown
    if err := gracefulServer.ListenAndServe(); err != nil {
        logging.Fatal("Server failed", zap.Error(err))
    }
}
```

## 5. Request Context and Tracing

### Request ID Middleware
```go
// relayer-middleware/api/pkg/middleware/request_id.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        c.Next()
    }
}
```

### Context Propagation
```go
// relayer-middleware/api/pkg/clearing/service.go
// Update service methods to accept context
func (s *Service) GenerateToken(ctx context.Context, request ClearingRequest) (*TokenResponse, error) {
    // Extract request ID from context
    requestID := ctx.Value("request_id").(string)
    
    logger := logging.With(
        zap.String("request_id", requestID),
        zap.String("operation", "generate_token"),
    )
    
    // Pass context through the call chain
    if err := s.validateRequest(ctx, request); err != nil {
        logger.Error("Request validation failed", zap.Error(err))
        return nil, err
    }
    
    // ... rest of implementation
}
```

## Testing Considerations

1. **Health Check Tests**
   - Test with all dependencies healthy
   - Test with degraded dependencies
   - Verify correct HTTP status codes

2. **Logging Tests**
   - Verify log levels work correctly
   - Test structured fields are included
   - Check sensitive data is not logged

3. **Connection Pool Tests**
   - Test under high concurrency
   - Verify pool exhaustion handling
   - Test connection recycling

4. **Graceful Shutdown Tests**
   - Test shutdown during active requests
   - Verify cleanup functions run
   - Test timeout scenarios

5. **Integration Tests**
   - Full server startup/shutdown cycle
   - Concurrent request handling
   - Resource cleanup verification