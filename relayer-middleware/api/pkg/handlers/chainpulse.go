package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	
	"relayooor/api/pkg/chainpulse"
)

// ChainpulseHandler handles chainpulse integration endpoints
type ChainpulseHandler struct {
	client *chainpulse.Client
	logger *zap.Logger
}

// NewChainpulseHandler creates a new chainpulse handler
func NewChainpulseHandler(chainpulseURL string, logger *zap.Logger) *ChainpulseHandler {
	return &ChainpulseHandler{
		client: chainpulse.NewClient(chainpulseURL, logger),
		logger: logger,
	}
}

// RegisterRoutes registers chainpulse routes
func (h *ChainpulseHandler) RegisterRoutes(api *gin.RouterGroup) {
	cp := api.Group("/chainpulse")
	{
		cp.GET("/packets/by-user", h.GetUserPackets)
		cp.GET("/packets/stuck", h.GetStuckPackets)
		cp.GET("/packets/:chain/:channel/:sequence", h.GetPacketDetails)
		cp.GET("/channels/congestion", h.GetChannelCongestion)
		cp.GET("/metrics", h.GetMetrics)
		cp.GET("/health", h.HealthCheck)
	}
}

// GetUserPackets returns packets for a specific user
func (h *ChainpulseHandler) GetUserPackets(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter required"})
		return
	}

	packets, err := h.client.GetPacketsByUser(c.Request.Context(), address)
	if err != nil {
		h.logger.Error("Failed to get user packets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user packets"})
		return
	}

	// Transform to match expected API format
	transfers := make([]gin.H, len(packets))
	for i, p := range packets {
		transfers[i] = gin.H{
			"id":               p.TxHash + "-" + strconv.FormatUint(p.Sequence, 10),
			"channelId":        p.Channel,
			"sequence":         p.Sequence,
			"sourceChain":      p.Chain,
			"destinationChain": getCounterpartyChain(p.Chain), // Helper function needed
			"amount":           p.Amount,
			"denom":            p.Denom,
			"sender":           p.Sender,
			"receiver":         p.Receiver,
			"status":           p.Status,
			"timestamp":        p.Timestamp,
			"txHash":           p.TxHash,
		}
		
		// Add stuck duration if packet is stuck
		if p.Status == "stuck" {
			duration := time.Since(p.Timestamp).String()
			transfers[i]["stuckDuration"] = &duration
		}
	}

	c.JSON(http.StatusOK, transfers)
}

// GetStuckPackets returns currently stuck packets
func (h *ChainpulseHandler) GetStuckPackets(c *gin.Context) {
	minStuckMinutes := 30 // Default
	if min := c.Query("min_stuck_minutes"); min != "" {
		if parsed, err := strconv.Atoi(min); err == nil {
			minStuckMinutes = parsed
		}
	}

	packets, err := h.client.GetStuckPackets(c.Request.Context(), minStuckMinutes)
	if err != nil {
		h.logger.Error("Failed to get stuck packets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stuck packets"})
		return
	}

	// Transform to match expected format
	stuckPackets := make([]gin.H, len(packets))
	for i, p := range packets {
		stuckPackets[i] = gin.H{
			"id":               strconv.FormatUint(p.Sequence, 10),
			"channelId":        p.SrcChannel,
			"sequence":         p.Sequence,
			"sourceChain":      p.SrcChain,
			"destinationChain": p.DstChain,
			"stuckDuration":    formatDuration(p.StuckDuration),
			"amount":           p.Amount,
			"denom":            p.Denom,
			"sender":           p.Sender,
			"receiver":         p.Receiver,
			"timestamp":        p.StuckSince,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stuck_packets": stuckPackets,
		"total":         len(stuckPackets),
	})
}

// GetPacketDetails returns details for a specific packet
func (h *ChainpulseHandler) GetPacketDetails(c *gin.Context) {
	chain := c.Param("chain")
	channel := c.Param("channel")
	sequence, err := strconv.ParseUint(c.Param("sequence"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sequence number"})
		return
	}

	details, err := h.client.GetPacketDetails(c.Request.Context(), chain, channel, sequence)
	if err != nil {
		h.logger.Error("Failed to get packet details", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch packet details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chain":           details.Chain,
		"channel":         details.Channel,
		"sequence":        details.Sequence,
		"status":          details.Status,
		"packet_data":     details.PacketData,
		"acknowledgement": details.Acknowledgement,
		"tx_hash":         details.TxHash,
		"height":          details.Height,
		"timestamp":       details.Timestamp,
	})
}

// GetChannelCongestion returns channel congestion statistics
func (h *ChainpulseHandler) GetChannelCongestion(c *gin.Context) {
	congestion, err := h.client.GetChannelCongestion(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get channel congestion", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channel congestion"})
		return
	}

	// Transform to match expected format
	channels := make([]gin.H, len(congestion))
	for i, ch := range congestion {
		channels[i] = gin.H{
			"channelId":             ch.Channel,
			"counterpartyChannelId": ch.CounterpartyChannel,
			"sourceChain":           ch.Chain,
			"destinationChain":      ch.CounterpartyChain,
			"pendingPackets":        ch.PendingPackets,
			"avgClearTime":          ch.AvgClearTime,
			"congestionLevel":       ch.CongestionLevel,
		}
	}

	c.JSON(http.StatusOK, channels)
}

// GetMetrics returns raw Prometheus metrics
func (h *ChainpulseHandler) GetMetrics(c *gin.Context) {
	metrics, err := h.client.GetMetrics(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}

	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(http.StatusOK, metrics)
}

// HealthCheck checks chainpulse health
func (h *ChainpulseHandler) HealthCheck(c *gin.Context) {
	if err := h.client.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Helper functions
func getCounterpartyChain(chain string) string {
	// This would normally come from configuration
	switch chain {
	case "cosmoshub-4":
		return "osmosis-1"
	case "osmosis-1":
		return "cosmoshub-4"
	default:
		return "unknown"
	}
}

func formatDuration(minutes int) string {
	if minutes < 60 {
		return strconv.Itoa(minutes) + "m"
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return strconv.Itoa(hours) + "h"
	}
	return strconv.Itoa(hours) + "h" + strconv.Itoa(mins) + "m"
}