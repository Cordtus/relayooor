package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChainEndpoint represents a chain's REST API endpoint
type ChainEndpoint struct {
	ChainID string `json:"chain_id"`
	RestURL string `json:"rest_url"`
}

// GetChainEndpoints returns known chain REST endpoints
func (h *Handlers) GetChainEndpoints(c *gin.Context) {
	// In production, these would come from a database or configuration
	endpoints := []ChainEndpoint{
		{ChainID: "cosmoshub-4", RestURL: "https://cosmos-rest.publicnode.com"},
		{ChainID: "osmosis-1", RestURL: "https://osmosis-rest.publicnode.com"},
		{ChainID: "neutron-1", RestURL: "https://neutron-rest.publicnode.com"},
		{ChainID: "noble-1", RestURL: "https://noble-rest.publicnode.com"},
		{ChainID: "axelar-dojo-1", RestURL: "https://axelar-rest.publicnode.com"},
		{ChainID: "stride-1", RestURL: "https://stride-rest.publicnode.com"},
		{ChainID: "dydx-mainnet-1", RestURL: "https://dydx-rest.publicnode.com"},
		{ChainID: "celestia", RestURL: "https://celestia-rest.publicnode.com"},
		{ChainID: "injective-1", RestURL: "https://injective-rest.publicnode.com"},
		{ChainID: "kava_2222-10", RestURL: "https://kava-rest.publicnode.com"},
		{ChainID: "secret-4", RestURL: "https://secret-rest.publicnode.com"},
		{ChainID: "stargaze-1", RestURL: "https://stargaze-rest.publicnode.com"},
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoints": endpoints,
	})
}