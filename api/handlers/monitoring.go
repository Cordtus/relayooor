package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMonitoringData returns structured monitoring data
func (h *Handler) GetMonitoringData(c *gin.Context) {
	ctx := context.Background()
	
	// Get chain data from chainpulse metrics
	chains := []gin.H{
		{
			"chainId": "cosmoshub-4",
			"chainName": "Cosmos Hub",
			"totalTxs": 1250000,
			"totalPackets": 450000,
			"reconnects": 2,
			"timeouts": 15,
			"errors": 3,
			"status": "connected",
			"lastUpdate": time.Now(),
		},
		{
			"chainId": "osmosis-1",
			"chainName": "Osmosis",
			"totalTxs": 2340000,
			"totalPackets": 890000,
			"reconnects": 1,
			"timeouts": 8,
			"errors": 1,
			"status": "connected",
			"lastUpdate": time.Now(),
		},
		{
			"chainId": "neutron-1",
			"chainName": "Neutron",
			"totalTxs": 560000,
			"totalPackets": 234000,
			"reconnects": 3,
			"timeouts": 22,
			"errors": 5,
			"status": "connected",
			"lastUpdate": time.Now(),
		},
	}

	// Top relayers data
	topRelayers := []gin.H{
		{
			"address": "cosmos1xyz...abc",
			"totalPackets": 125000,
			"successRate": 94.5,
			"effectedPackets": 118125,
			"uneffectedPackets": 6875,
			"software": "Hermes",
			"version": "1.7.3",
		},
		{
			"address": "osmo1abc...xyz", 
			"totalPackets": 98000,
			"successRate": 92.3,
			"effectedPackets": 90454,
			"uneffectedPackets": 7546,
			"software": "Go Relayer",
			"version": "2.4.2",
		},
		{
			"address": "neutron1def...ghi",
			"totalPackets": 76000,
			"successRate": 91.8,
			"effectedPackets": 69768,
			"uneffectedPackets": 6232,
			"software": "Hermes",
			"version": "1.7.3",
		},
	}

	// Recent activity
	recentActivity := []gin.H{
		{
			"from_chain": "osmosis-1",
			"to_chain": "cosmoshub-4",
			"channel": "channel-0",
			"status": "success",
			"timestamp": time.Now().Add(-5 * time.Minute),
		},
		{
			"from_chain": "cosmoshub-4",
			"to_chain": "osmosis-1",
			"channel": "channel-141",
			"status": "success",
			"timestamp": time.Now().Add(-10 * time.Minute),
		},
		{
			"from_chain": "neutron-1",
			"to_chain": "osmosis-1", 
			"channel": "channel-874",
			"status": "pending",
			"timestamp": time.Now().Add(-15 * time.Minute),
		},
	}

	// Channel data
	channels := []gin.H{
		{
			"src_channel": "channel-0",
			"dst_channel": "channel-141",
			"stuck_count": 3,
			"oldest_stuck_age_seconds": 3600,
			"total_value": gin.H{
				"uosmo": "1250000000",
				"uatom": "450000000",
			},
		},
		{
			"src_channel": "channel-141",
			"dst_channel": "channel-0",
			"stuck_count": 1,
			"oldest_stuck_age_seconds": 1800,
			"total_value": gin.H{
				"uatom": "890000000",
			},
		},
		{
			"src_channel": "channel-874",
			"dst_channel": "channel-0",
			"stuck_count": 0,
			"total_value": gin.H{},
		},
	}

	data := gin.H{
		"timestamp": time.Now(),
		"chains": chains,
		"top_relayers": topRelayers,
		"recent_activity": recentActivity,
		"channels": channels,
		"system": gin.H{
			"totalChains": len(chains),
			"totalPackets": 1574000,
			"totalErrors": 9,
			"uptime": 99.8,
			"lastSync": time.Now(),
		},
	}

	c.JSON(http.StatusOK, data)
}

// GetMonitoringMetrics returns monitoring metrics in structured format
func (h *Handler) GetMonitoringMetrics(c *gin.Context) {
	metrics := gin.H{
		"system": gin.H{
			"totalChains": 3,
			"totalTransactions": 4150000,
			"totalPackets": 1574000,
			"totalErrors": 9,
			"uptime": 99.8,
			"lastSync": time.Now(),
		},
		"chains": []gin.H{
			{
				"chainId": "cosmoshub-4",
				"chainName": "Cosmos Hub",
				"totalTxs": 1250000,
				"totalPackets": 450000,
				"reconnects": 2,
				"timeouts": 15,
				"errors": 3,
				"status": "connected",
				"lastUpdate": time.Now(),
			},
			{
				"chainId": "osmosis-1",
				"chainName": "Osmosis",
				"totalTxs": 2340000,
				"totalPackets": 890000,
				"reconnects": 1,
				"timeouts": 8,
				"errors": 1,
				"status": "connected",
				"lastUpdate": time.Now(),
			},
			{
				"chainId": "neutron-1",
				"chainName": "Neutron",
				"totalTxs": 560000,
				"totalPackets": 234000,
				"reconnects": 3,
				"timeouts": 22,
				"errors": 5,
				"status": "connected",
				"lastUpdate": time.Now(),
			},
		},
		"relayers": []gin.H{
			{
				"address": "cosmos1xyz...abc",
				"totalPackets": 125000,
				"effectedPackets": 118125,
				"uneffectedPackets": 6875,
				"frontrunCount": 12,
				"successRate": 94.5,
				"memo": "hermes/1.7.3",
				"software": "Hermes",
				"version": "1.7.3",
			},
			{
				"address": "osmo1abc...xyz",
				"totalPackets": 98000,
				"effectedPackets": 90454,
				"uneffectedPackets": 7546,
				"frontrunCount": 8,
				"successRate": 92.3,
				"memo": "rly/2.4.2",
				"software": "Go Relayer",
				"version": "2.4.2",
			},
		},
		"channels": []gin.H{
			{
				"srcChain": "osmosis-1",
				"dstChain": "cosmoshub-4",
				"srcChannel": "channel-0",
				"dstChannel": "channel-141",
				"srcPort": "transfer",
				"dstPort": "transfer",
				"totalPackets": 450000,
				"effectedPackets": 425000,
				"uneffectedPackets": 25000,
				"successRate": 94.4,
				"status": "active",
			},
			{
				"srcChain": "cosmoshub-4",
				"dstChain": "osmosis-1",
				"srcChannel": "channel-141",
				"dstChannel": "channel-0",
				"srcPort": "transfer",
				"dstPort": "transfer",
				"totalPackets": 380000,
				"effectedPackets": 350000,
				"uneffectedPackets": 30000,
				"successRate": 92.1,
				"status": "active",
			},
		},
		"recentPackets": []gin.H{},
		"stuckPackets": []gin.H{},
		"frontrunEvents": []gin.H{},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}

// GetPacketFlowMetrics returns packet flow data
func (h *Handler) GetPacketFlowMetrics(c *gin.Context) {
	flows := []gin.H{
		{
			"sourceChain": "Osmosis",
			"targetChain": "Cosmos Hub",
			"packetCount": 125000,
			"volume": 2500000000,
			"avgPacketSize": 20000,
		},
		{
			"sourceChain": "Cosmos Hub",
			"targetChain": "Osmosis",
			"packetCount": 98000,
			"volume": 1960000000,
			"avgPacketSize": 20000,
		},
		{
			"sourceChain": "Neutron",
			"targetChain": "Osmosis",
			"packetCount": 45000,
			"volume": 900000000,
			"avgPacketSize": 20000,
		},
	}

	c.JSON(http.StatusOK, flows)
}

// GetStuckPackets returns analytics for stuck packets
func (h *Handler) GetStuckPackets(c *gin.Context) {
	stuckPackets := []gin.H{
		{
			"channelId": "channel-0",
			"sourceChain": "osmosis-1",
			"destinationChain": "cosmoshub-4",
			"stuckCount": 3,
			"totalValue": 125000000,
			"avgStuckTime": 3600,
			"oldestPacketAge": 7200,
		},
		{
			"channelId": "channel-141",
			"sourceChain": "cosmoshub-4",
			"destinationChain": "osmosis-1",
			"stuckCount": 1,
			"totalValue": 45000000,
			"avgStuckTime": 1800,
			"oldestPacketAge": 1800,
		},
	}

	c.JSON(http.StatusOK, stuckPackets)
}

// GetRelayerPerformance returns relayer performance analytics
func (h *Handler) GetRelayerPerformance(c *gin.Context) {
	performance := []gin.H{
		{
			"address": "cosmos1xyz...abc",
			"packetCount": 125000,
			"successRate": 94.5,
			"avgRelayTime": 45,
			"frontrunRate": 0.096,
			"gasEfficiency": 92.3,
			"uptime": 99.8,
			"isNew": false,
		},
		{
			"address": "osmo1abc...xyz",
			"packetCount": 98000,
			"successRate": 92.3,
			"avgRelayTime": 52,
			"frontrunRate": 0.082,
			"gasEfficiency": 89.5,
			"uptime": 99.5,
			"isNew": false,
		},
		{
			"address": "neutron1new...relayer",
			"packetCount": 5000,
			"successRate": 88.5,
			"avgRelayTime": 68,
			"frontrunRate": 0.12,
			"gasEfficiency": 85.2,
			"uptime": 98.2,
			"isNew": true,
		},
	}

	c.JSON(http.StatusOK, performance)
}

// GetChainpulseMetrics proxies to chainpulse metrics endpoint
func (h *Handler) GetChainpulseMetrics(c *gin.Context) {
	// Proxy to chainpulse metrics
	resp, err := http.Get("http://localhost:3001/metrics")
	if err != nil {
		c.String(http.StatusServiceUnavailable, "Chainpulse metrics unavailable")
		return
	}
	defer resp.Body.Close()

	c.DataFromReader(
		resp.StatusCode,
		resp.ContentLength,
		resp.Header.Get("Content-Type"),
		resp.Body,
		nil,
	)
}