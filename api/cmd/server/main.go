package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cordt/relayooor/api/pkg/handlers"
	"github.com/cordt/relayooor/api/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Initialize handlers
	h := handlers.NewHandler()

	// API routes
	api := router.Group("/")
	{
		// Health check
		api.GET("/health", h.HealthCheck)

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.Login)
			auth.POST("/refresh", h.RefreshToken)
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
				ibc.GET("/chains", h.GetChains)
				ibc.GET("/chains/:chain_id", h.GetChain)
				ibc.GET("/chains/:chain_id/status", h.GetChainStatus)

				// Channels
				ibc.GET("/channels", h.GetChannels)
				ibc.GET("/chains/:chain_id/channels", h.GetChainChannels)
				ibc.GET("/channels/:channel_id", h.GetChannel)

				// Connections
				ibc.GET("/connections", h.GetConnections)
				ibc.GET("/chains/:chain_id/connections", h.GetChainConnections)

				// Packets
				ibc.GET("/packets/pending", h.GetPendingPackets)
				ibc.POST("/packets/clear", h.ClearPackets)
				ibc.GET("/packets/stuck", h.GetStuckPackets)

				// Clients
				ibc.GET("/clients", h.GetClients)
				ibc.GET("/chains/:chain_id/clients", h.GetChainClients)
			}

			// Relayer management
			relayer := protected.Group("/relayer")
			{
				relayer.GET("/status", h.GetRelayerStatus)
				relayer.POST("/hermes/start", h.StartHermes)
				relayer.POST("/hermes/stop", h.StopHermes)
				relayer.POST("/rly/start", h.StartGoRelayer)
				relayer.POST("/rly/stop", h.StopGoRelayer)
				relayer.GET("/config", h.GetRelayerConfig)
				relayer.PUT("/config", h.UpdateRelayerConfig)
			}

			// Metrics and monitoring
			metrics := protected.Group("/metrics")
			{
				metrics.GET("/summary", h.GetMetricsSummary)
				metrics.GET("/packets", h.GetPacketMetrics)
				metrics.GET("/channels", h.GetChannelMetrics)
			}
		}
	}

	// WebSocket endpoint for real-time updates
	router.GET("/ws", h.WebSocketHandler)

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server started on port %s", port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}