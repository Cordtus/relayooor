package handlers

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetChains returns all configured chains
func (h *Handler) GetChains(c *gin.Context) {
	// Get chains from Hermes
	hermesChains, err := h.callHermesAPI("/chains")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chains from Hermes"})
		return
	}

	// Get chains from Go relayer
	cmd := exec.Command("rly", "chains", "list", "--json")
	output, err := cmd.Output()
	var goRelayerChains interface{}
	if err == nil {
		// Parse JSON output
		// For now, we'll just note that we tried
		goRelayerChains = string(output)
	}

	c.JSON(http.StatusOK, gin.H{
		"hermes":      hermesChains,
		"go_relayer":  goRelayerChains,
	})
}

// GetChain returns details for a specific chain
func (h *Handler) GetChain(c *gin.Context) {
	chainID := c.Param("chain_id")
	
	result, err := h.callHermesAPI("/chain/" + chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chain details"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetChainStatus returns the status of a specific chain
func (h *Handler) GetChainStatus(c *gin.Context) {
	chainID := c.Param("chain_id")

	// Get chain status from relayers
	status := gin.H{
		"chain_id": chainID,
		"hermes":   "unknown",
		"rly":      "unknown",
	}

	// Check Hermes status
	if state, err := h.callHermesAPI("/state"); err == nil {
		// Parse state to check if chain is active
		status["hermes"] = "active"
	}

	c.JSON(http.StatusOK, status)
}

// GetChannels returns all channels
func (h *Handler) GetChannels(c *gin.Context) {
	// This would aggregate channel data from both relayers
	c.JSON(http.StatusOK, gin.H{
		"channels": []interface{}{},
		"message":  "Channel listing implementation pending",
	})
}

// GetChainChannels returns channels for a specific chain
func (h *Handler) GetChainChannels(c *gin.Context) {
	chainID := c.Param("chain_id")
	
	c.JSON(http.StatusOK, gin.H{
		"chain_id": chainID,
		"channels": []interface{}{},
		"message":  "Chain channel listing implementation pending",
	})
}

// GetChannel returns details for a specific channel
func (h *Handler) GetChannel(c *gin.Context) {
	channelID := c.Param("channel_id")
	
	c.JSON(http.StatusOK, gin.H{
		"channel_id": channelID,
		"details":    "Channel details implementation pending",
	})
}

// GetConnections returns all connections
func (h *Handler) GetConnections(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connections": []interface{}{},
		"message":     "Connection listing implementation pending",
	})
}

// GetChainConnections returns connections for a specific chain
func (h *Handler) GetChainConnections(c *gin.Context) {
	chainID := c.Param("chain_id")
	
	c.JSON(http.StatusOK, gin.H{
		"chain_id":    chainID,
		"connections": []interface{}{},
		"message":     "Chain connection listing implementation pending",
	})
}

// GetPendingPackets returns all pending packets
func (h *Handler) GetPendingPackets(c *gin.Context) {
	// Query both relayers for pending packets
	c.JSON(http.StatusOK, gin.H{
		"pending_packets": []interface{}{},
		"message":        "Pending packet query implementation pending",
	})
}

// ClearPackets triggers packet clearing
func (h *Handler) ClearPackets(c *gin.Context) {
	var req struct {
		ChainID   string `json:"chain_id"`
		ChannelID string `json:"channel_id"`
		UseHermes bool   `json:"use_hermes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use Hermes to clear packets
	if req.UseHermes {
		params := make(map[string]interface{})
		if req.ChainID != "" {
			params["chain"] = req.ChainID
		}

		result, err := h.postHermesAPI("/clear_packets", params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear packets"})
			return
		}

		// Broadcast update to WebSocket clients
		h.broadcast <- gin.H{
			"type":    "packet_clear",
			"chain":   req.ChainID,
			"channel": req.ChannelID,
			"result":  result,
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"result":  result,
			"relayer": "hermes",
		})
		return
	}

	// Use Go relayer
	args := []string{"tx", "relay-packets"}
	if req.ChainID != "" && req.ChannelID != "" {
		args = append(args, req.ChainID, req.ChannelID)
	}

	cmd := exec.Command("rly", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to clear packets with rly",
			"output": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"output":  string(output),
		"relayer": "rly",
	})
}

// GetStuckPackets returns packets that appear to be stuck
func (h *Handler) GetStuckPackets(c *gin.Context) {
	// This would analyze packet sequences and timestamps to identify stuck packets
	c.JSON(http.StatusOK, gin.H{
		"stuck_packets": []interface{}{},
		"message":       "Stuck packet detection implementation pending",
	})
}

// GetClients returns all IBC clients
func (h *Handler) GetClients(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"clients": []interface{}{},
		"message": "Client listing implementation pending",
	})
}

// GetChainClients returns clients for a specific chain
func (h *Handler) GetChainClients(c *gin.Context) {
	chainID := c.Param("chain_id")
	
	c.JSON(http.StatusOK, gin.H{
		"chain_id": chainID,
		"clients":  []interface{}{},
		"message":  "Chain client listing implementation pending",
	})
}