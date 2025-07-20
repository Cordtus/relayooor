package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// GetChainpulseMetrics proxies to the actual Chainpulse metrics endpoint
func GetChainpulseMetrics(c *gin.Context) {
	// Get Chainpulse URL from environment
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://localhost:3001"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request to Chainpulse metrics endpoint
	resp, err := client.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	if err != nil {
		c.String(http.StatusServiceUnavailable, fmt.Sprintf("Failed to connect to Chainpulse: %v", err))
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to read Chainpulse response: %v", err))
		return
	}

	// Forward the metrics with the same content type
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(resp.StatusCode, string(body))
}

// GetMonitoringData returns structured monitoring data for the dashboard
func GetMonitoringData(c *gin.Context) {
	// Get Chainpulse URL from environment
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://localhost:3001"
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch metrics from Chainpulse
	metricsResp, err := client.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("Failed to connect to Chainpulse: %v", err),
		})
		return
	}
	defer metricsResp.Body.Close()

	// Parse metrics to extract data
	metricsBody, err := io.ReadAll(metricsResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read metrics: %v", err),
		})
		return
	}

	// Parse metrics and build monitoring data
	metrics := string(metricsBody)
	chains := parseChainData(metrics)
	channels := parseChannelData(metrics)
	relayers := parseRelayerData(metrics)
	alerts := parseAlerts(metrics)

	data := gin.H{
		"status":       "healthy",
		"chains":       chains,
		"channels":     channels,
		"top_relayers": relayers,
		"alerts":       alerts,
		"timestamp":    time.Now(),
	}

	c.JSON(http.StatusOK, data)
}