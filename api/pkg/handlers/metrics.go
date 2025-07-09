package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMetricsSummary returns a summary of key metrics
func (h *Handler) GetMetricsSummary(c *gin.Context) {
	ctx := context.Background()
	
	// Get cached metrics from Redis
	cachedMetrics, _ := h.redisClient.Get(ctx, "metrics:summary").Result()
	
	summary := gin.H{
		"timestamp": time.Now().Unix(),
		"relayers": gin.H{
			"hermes": h.getHermesMetrics(),
			"rly":    h.getRlyMetrics(),
		},
		"cached": cachedMetrics,
	}

	// Cache the summary
	summaryJSON, _ := json.Marshal(summary)
	h.redisClient.Set(ctx, "metrics:summary", summaryJSON, 30*time.Second)

	c.JSON(http.StatusOK, summary)
}

// GetPacketMetrics returns packet-specific metrics
func (h *Handler) GetPacketMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "1h")
	
	metrics := gin.H{
		"time_range": timeRange,
		"packets": gin.H{
			"total_sent":     0,
			"total_received": 0,
			"pending":        0,
			"stuck":          0,
			"cleared":        0,
		},
		"by_chain": make(map[string]interface{}),
	}

	// TODO: Implement actual metric collection from relayers
	
	c.JSON(http.StatusOK, metrics)
}

// GetChannelMetrics returns channel-specific metrics
func (h *Handler) GetChannelMetrics(c *gin.Context) {
	metrics := gin.H{
		"total_channels": 0,
		"active":        0,
		"inactive":      0,
		"by_state":      make(map[string]int),
		"by_chain_pair": make(map[string]interface{}),
	}

	// TODO: Implement actual channel metric collection
	
	c.JSON(http.StatusOK, metrics)
}

// Helper functions for getting metrics from relayers
func (h *Handler) getHermesMetrics() gin.H {
	metrics := gin.H{
		"status": "unknown",
		"stats":  nil,
	}

	// Get Hermes state
	if state, err := h.callHermesAPI("/state"); err == nil {
		metrics["status"] = "active"
		metrics["state"] = state
	}

	return metrics
}

func (h *Handler) getRlyMetrics() gin.H {
	// The Go relayer exposes Prometheus metrics on port 5184
	metrics := gin.H{
		"status": "unknown",
		"stats":  nil,
	}

	// Try to fetch metrics from the metrics endpoint
	resp, err := http.Get("http://localhost:5184/relayer/metrics")
	if err == nil {
		defer resp.Body.Close()
		metrics["status"] = "active"
		// TODO: Parse Prometheus metrics
	}

	return metrics
}

// MetricsCollector runs in the background to collect and cache metrics
func (h *Handler) StartMetricsCollector() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.collectMetrics()
			}
		}
	}()
}

func (h *Handler) collectMetrics() {
	ctx := context.Background()
	
	// Collect metrics from both relayers
	metrics := gin.H{
		"timestamp": time.Now().Unix(),
		"hermes":    h.getHermesMetrics(),
		"rly":       h.getRlyMetrics(),
	}

	// Store in Redis
	data, _ := json.Marshal(metrics)
	h.redisClient.Set(ctx, "metrics:latest", data, 5*time.Minute)

	// Broadcast to WebSocket clients
	h.broadcast <- gin.H{
		"type":    "metrics_update",
		"metrics": metrics,
	}
}