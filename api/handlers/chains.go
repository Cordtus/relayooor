package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/relayooor/api/internal/config"
)

// ChainEndpoint represents a chain's REST API endpoint
type ChainEndpoint struct {
	ChainID string `json:"chain_id"`
	RestURL string `json:"rest_url"`
}

// GetChainEndpoints returns known chain REST endpoints
func (h *Handler) GetChainEndpoints(c *gin.Context) {
	registry := config.DefaultChainRegistry()
	
	endpoints := make([]ChainEndpoint, 0, len(registry.Chains))
	for chainID, chain := range registry.Chains {
		if chain.RESTEndpoint != "" {
			endpoints = append(endpoints, ChainEndpoint{
				ChainID: chainID,
				RestURL: chain.RESTEndpoint,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoints": endpoints,
	})
}

// GetChainRegistry returns the complete chain registry with all endpoints
func (h *Handler) GetChainRegistry(c *gin.Context) {
	registry := config.DefaultChainRegistry()
	c.JSON(http.StatusOK, registry)
}