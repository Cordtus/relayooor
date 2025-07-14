package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	
	"relayooor/api/pkg/clearing"
	"relayooor/api/pkg/database"
	"relayooor/api/pkg/handlers"
	"relayooor/api/pkg/logging"
	"relayooor/api/pkg/middleware"
	"relayooor/api/pkg/server"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize logger
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, sqlDB, err := initializeDatabase(logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer sqlDB.Close()

	// Initialize Redis
	redisClient := initializeRedis(logger)
	defer redisClient.Close()

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORS())

	// Initialize health handler
	healthHandler := handlers.NewHealthHandler(db, redisClient, logger)

	// Initialize clearing handlers with improved error handling
	clearingHandlers := clearing.NewHandlersV2(db, redisClient, logger)

	// Initialize payment handler for UX improvements
	paymentHandler := handlers.NewPaymentHandler(db, redisClient, logger)

	// Initialize help handler for tooltips
	helpHandler := handlers.NewHelpHandler()

	// Initialize packet stream handler
	packetCache := clearing.NewPacketCache(redisClient, logger)
	packetStreamHandler := handlers.NewPacketStreamHandler(db, packetCache, logger)

	// Initialize original handlers for backward compatibility
	originalHandlers := handlers.NewHandler()

	// Initialize Chainpulse handler
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		// Default to local chainpulse instance on correct port
		chainpulseURL = "http://localhost:3000"
	}
	chainpulseHandler := handlers.NewChainpulseHandler(chainpulseURL, logger)

	// API routes
	api := router.Group("/api/v1")
	{
		// Health check endpoint
		api.GET("/health", healthHandler.GetHealth)

		// Clearing service routes
		clearingHandlers.RegisterRoutes(api)
		
		// Payment and UX routes
		paymentHandler.RegisterRoutes(api)
		
		// Help routes
		helpHandler.RegisterRoutes(api)
		
		// Packet stream routes
		api.GET("/packets/stuck/stream", packetStreamHandler.GetStuckPacketsStream)
		api.GET("/packets/channel/stream", packetStreamHandler.GetChannelPacketsStream)
		
		// Chainpulse integration routes
		chainpulseHandler.RegisterRoutes(api)

		// Original authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", originalHandlers.Login)
			auth.POST("/refresh", originalHandlers.RefreshToken)
		}

		// Protected routes
		protected := api.Group("/")
		if os.Getenv("AUTH_ENABLED") == "true" {
			protected.Use(middleware.AuthRequired())
		}
		{
			// IBC routes
			ibc := protected.Group("/ibc")
			{
				// Chains
				ibc.GET("/chains", originalHandlers.GetChains)
				ibc.GET("/chains/:chain_id", originalHandlers.GetChain)
				ibc.GET("/chains/:chain_id/status", originalHandlers.GetChainStatus)

				// Channels
				ibc.GET("/channels", originalHandlers.GetChannels)
				ibc.GET("/chains/:chain_id/channels", originalHandlers.GetChainChannels)
				ibc.GET("/channels/:channel_id", originalHandlers.GetChannel)

				// Connections
				ibc.GET("/connections", originalHandlers.GetConnections)
				ibc.GET("/chains/:chain_id/connections", originalHandlers.GetChainConnections)

				// Packets
				ibc.GET("/packets/pending", originalHandlers.GetPendingPackets)
				ibc.POST("/packets/clear", originalHandlers.ClearPackets)
				ibc.GET("/packets/stuck", originalHandlers.GetStuckPackets)

				// Clients
				ibc.GET("/clients", originalHandlers.GetClients)
				ibc.GET("/chains/:chain_id/clients", originalHandlers.GetChainClients)
			}

			// Relayer management
			relayer := protected.Group("/relayer")
			{
				relayer.GET("/status", originalHandlers.GetRelayerStatus)
				relayer.POST("/hermes/start", originalHandlers.StartHermes)
				relayer.POST("/hermes/stop", originalHandlers.StopHermes)
				relayer.POST("/rly/start", originalHandlers.StartGoRelayer)
				relayer.POST("/rly/stop", originalHandlers.StopGoRelayer)
				relayer.GET("/config", originalHandlers.GetRelayerConfig)
				relayer.PUT("/config", originalHandlers.UpdateRelayerConfig)
			}

			// Metrics and monitoring
			metrics := protected.Group("/metrics")
			{
				metrics.GET("/summary", originalHandlers.GetMetricsSummary)
				metrics.GET("/packets", originalHandlers.GetPacketMetrics)
				metrics.GET("/channels", originalHandlers.GetChannelMetrics)
				metrics.GET("/chainpulse", originalHandlers.GetChainpulseMetrics)
				metrics.GET("/packet-flow", originalHandlers.GetPacketFlowMetrics)
				metrics.GET("/stuck-packets", originalHandlers.GetStuckPackets)
				metrics.GET("/relayer-performance", originalHandlers.GetRelayerPerformance)
			}
		}
	}

	// Legacy WebSocket endpoint
	router.GET("/ws", originalHandlers.WebSocketHandler)

	// Start server with graceful shutdown
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create shutdown handler
	shutdownHandler := server.NewShutdownHandler(srv, logger)

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	if err := shutdownHandler.Shutdown(context.Background()); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}
}

// initializeDatabase sets up the database connection with connection pooling
func initializeDatabase(logger *zap.Logger) (*gorm.DB, *sql.DB, error) {
	// Get database configuration from environment
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "relayooor"
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logging.NewGormLogger(logger),
	})
	if err != nil {
		return nil, nil, err
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// Configure connection pool
	config := database.DefaultConfig()
	if err := config.ConfigureDB(sqlDB); err != nil {
		return nil, nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, nil, err
	}

	// Start index monitoring
	indexMonitor := database.NewIndexMonitor(db, logger)
	go indexMonitor.StartMonitoring(context.Background(), 1*time.Hour)

	logger.Info("Database initialized successfully",
		zap.String("host", dbHost),
		zap.String("database", dbName),
	)

	return db, sqlDB, nil
}

// initializeRedis sets up the Redis connection
func initializeRedis(logger *zap.Logger) *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // Default DB

	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     20,
		MinIdleConns: 5,
		MaxRetries:   3,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		// Don't fail startup, Redis is optional for some features
	} else {
		logger.Info("Redis initialized successfully", zap.String("addr", redisAddr))
	}

	return client
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB) error {
	// Auto-migrate clearing operation tables
	return db.AutoMigrate(
		&clearing.ClearingOperation{},
		&clearing.PaymentRecord{},
		// Add other models as needed
	)
}