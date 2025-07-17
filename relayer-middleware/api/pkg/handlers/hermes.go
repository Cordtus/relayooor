package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"relayooor/api/pkg/clearing"
)

// GetHermesVersion returns the Hermes version information
func (h *Handler) GetHermesVersion(c *gin.Context) {
	hermesClient := clearing.NewHermesClient(h.hermesURL)
	
	version, err := hermesClient.GetVersion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Failed to connect to Hermes",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version.Version,
		"status":  "connected",
	})
}

// GetHermesHealth checks if Hermes is healthy and responsive
func (h *Handler) GetHermesHealth(c *gin.Context) {
	// Try to get version as a health check
	hermesClient := clearing.NewHermesClient(h.hermesURL)
	
	_, err := hermesClient.GetVersion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"healthy": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"healthy": true,
		"service": "hermes",
		"url":     h.hermesURL,
	})
}

// ClearPacketsWithHermes handles direct packet clearing through Hermes
func (h *Handler) ClearPacketsWithHermes(c *gin.Context) {
	var req clearing.ClearPacketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Chain == "" || req.Channel == "" || req.Port == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chain, channel, and port are required",
		})
		return
	}

	// Create Hermes client
	hermesClient := clearing.NewHermesClient(h.hermesURL)

	// Clear packets
	response, err := hermesClient.ClearPackets(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear packets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}