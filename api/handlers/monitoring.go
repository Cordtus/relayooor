package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/relayooor/api/internal/config"
)

// GetMonitoringData returns structured monitoring data
func (h *Handler) GetMonitoringData(c *gin.Context) {
	// Get stuck packets data from Chainpulse API
	stuckPacketsResp, err := h.chainpulseClient.GetStuckPackets(900, 100) // 15 min threshold
	if err != nil {
		log.Printf("Error fetching stuck packets: %v", err)
	}
	
	// Get channel congestion data
	congestionResp, err := h.chainpulseClient.GetChannelCongestion()
	if err != nil {
		log.Printf("Error fetching channel congestion: %v", err)
	}
	
	// Get chain registry for metadata
	registry := config.DefaultChainRegistry()
	
	// Build chains data - for now using basic connected status
	// TODO: Parse actual metrics from Chainpulse prometheus endpoint
	chains := []gin.H{}
	for chainID, chainConfig := range registry.Chains {
		chains = append(chains, gin.H{
			"chainId": chainID,
			"chainName": chainConfig.ChainName,
			"totalTxs": 0, // TODO: Parse from metrics
			"totalPackets": 0, // TODO: Parse from metrics
			"reconnects": 0, // TODO: Parse from metrics
			"timeouts": 0, // TODO: Parse from metrics
			"errors": 0, // TODO: Parse from metrics
			"status": "connected",
			"lastUpdate": time.Now(),
		})
	}

	// Top relayers data - extract from stuck packets
	relayerStats := make(map[string]gin.H)
	if stuckPacketsResp != nil {
		for _, packet := range stuckPacketsResp.Packets {
			if packet.LastAttemptBy != "" {
				if _, exists := relayerStats[packet.LastAttemptBy]; !exists {
					relayerStats[packet.LastAttemptBy] = gin.H{
						"address": packet.LastAttemptBy,
						"totalPackets": 0,
						"successRate": 0.0,
						"effectedPackets": 0,
						"uneffectedPackets": 0,
						"software": "Unknown",
						"version": "Unknown",
					}
				}
				stats := relayerStats[packet.LastAttemptBy]
				stats["uneffectedPackets"] = stats["uneffectedPackets"].(int) + 1
				stats["totalPackets"] = stats["totalPackets"].(int) + 1
			}
		}
	}
	
	topRelayers := []gin.H{}
	for _, stats := range relayerStats {
		topRelayers = append(topRelayers, stats)
	}

	// Recent activity - derive from stuck packets
	recentActivity := []gin.H{}
	if stuckPacketsResp != nil && len(stuckPacketsResp.Packets) > 0 {
		// Show up to 10 most recent stuck packets as pending activities
		limit := 10
		if len(stuckPacketsResp.Packets) < limit {
			limit = len(stuckPacketsResp.Packets)
		}
		
		for i := 0; i < limit; i++ {
			packet := stuckPacketsResp.Packets[i]
			recentActivity = append(recentActivity, gin.H{
				"from_chain": packet.ChainID,
				"to_chain": getDestinationChainFromChannel(packet.ChainID, packet.SrcChannel),
				"channel": packet.SrcChannel,
				"status": "pending",
				"timestamp": time.Now().Add(-time.Duration(packet.AgeSeconds) * time.Second),
			})
		}
	}

	// Channel data from congestion response
	channels := []gin.H{}
	if congestionResp != nil {
		for _, ch := range congestionResp.Channels {
			channels = append(channels, gin.H{
				"src_channel": ch.SrcChannel,
				"dst_channel": ch.DstChannel,
				"stuck_count": ch.StuckCount,
				"oldest_stuck_age_seconds": ch.OldestStuckAgeSeconds,
				"total_value": ch.TotalValue,
			})
		}
	}

	// Calculate totals
	totalStuckPackets := 0
	if stuckPacketsResp != nil {
		totalStuckPackets = stuckPacketsResp.Total
	}
	
	data := gin.H{
		"timestamp": time.Now(),
		"chains": chains,
		"top_relayers": topRelayers,
		"recent_activity": recentActivity,
		"channels": channels,
		"system": gin.H{
			"totalChains": len(chains),
			"totalPackets": totalStuckPackets, // Only showing stuck packets for now
			"totalErrors": 0, // TODO: Parse from metrics
			"uptime": 99.8, // TODO: Calculate from actual uptime
			"lastSync": time.Now(),
		},
	}

	c.JSON(http.StatusOK, data)
}

// getDestinationChainFromChannel tries to determine destination chain from channel ID
func getDestinationChainFromChannel(sourceChain, srcChannel string) string {
	// This is a simplified mapping - in production should use IBC query
	// TODO: Query actual IBC channel info
	channelMap := map[string]map[string]string{
		"osmosis-1": {
			"channel-0":   "cosmoshub-4",
			"channel-750": "noble-1",
			"channel-874": "neutron-1",
		},
		"cosmoshub-4": {
			"channel-141": "osmosis-1",
			"channel-536": "noble-1",
		},
		"noble-1": {
			"channel-1": "osmosis-1",
			"channel-4": "cosmoshub-4",
		},
	}
	
	if chains, ok := channelMap[sourceChain]; ok {
		if dest, ok := chains[srcChannel]; ok {
			return dest
		}
	}
	
	return fmt.Sprintf("%s-dest", srcChannel) // Fallback
}

// GetMonitoringMetrics returns monitoring metrics in structured format
func (h *Handler) GetMonitoringMetrics(c *gin.Context) {
	// Get data from Chainpulse
	stuckPacketsResp, err := h.chainpulseClient.GetStuckPackets(900, 100)
	if err != nil {
		log.Printf("Error fetching stuck packets: %v", err)
	}
	
	congestionResp, err := h.chainpulseClient.GetChannelCongestion()
	if err != nil {
		log.Printf("Error fetching channel congestion: %v", err)
	}
	
	registry := config.DefaultChainRegistry()
	
	// Build chains data
	chains := []gin.H{}
	for chainID, chainConfig := range registry.Chains {
		chains = append(chains, gin.H{
			"chainId": chainID,
			"chainName": chainConfig.ChainName,
			"totalTxs": 0, // TODO: Parse from metrics
			"totalPackets": 0, // TODO: Parse from metrics
			"reconnects": 0, // TODO: Parse from metrics
			"timeouts": 0, // TODO: Parse from metrics
			"errors": 0, // TODO: Parse from metrics
			"status": "connected",
			"lastUpdate": time.Now(),
		})
	}
	
	// Build relayer stats
	relayerMap := make(map[string]gin.H)
	if stuckPacketsResp != nil {
		for _, packet := range stuckPacketsResp.Packets {
			if packet.LastAttemptBy != "" {
				if _, exists := relayerMap[packet.LastAttemptBy]; !exists {
					relayerMap[packet.LastAttemptBy] = gin.H{
						"address": packet.LastAttemptBy,
						"totalPackets": 0,
						"effectedPackets": 0,
						"uneffectedPackets": 0,
						"frontrunCount": 0,
						"successRate": 0.0,
						"memo": "",
						"software": "Unknown",
						"version": "Unknown",
					}
				}
				stats := relayerMap[packet.LastAttemptBy]
				stats["uneffectedPackets"] = stats["uneffectedPackets"].(int) + 1
				stats["totalPackets"] = stats["totalPackets"].(int) + 1
			}
		}
	}
	
	relayers := []gin.H{}
	for _, stats := range relayerMap {
		relayers = append(relayers, stats)
	}
	
	// Build channel data
	channels := []gin.H{}
	if congestionResp != nil {
		for _, ch := range congestionResp.Channels {
			// Determine source/dest chains
			srcChain := "unknown"
			dstChain := "unknown"
			
			// Try to infer from stuck packets
			if stuckPacketsResp != nil {
				for _, packet := range stuckPacketsResp.Packets {
					if packet.SrcChannel == ch.SrcChannel {
						srcChain = packet.ChainID
						dstChain = getDestinationChainFromChannel(packet.ChainID, packet.SrcChannel)
						break
					}
				}
			}
			
			channels = append(channels, gin.H{
				"srcChain": srcChain,
				"dstChain": dstChain,
				"srcChannel": ch.SrcChannel,
				"dstChannel": ch.DstChannel,
				"srcPort": "transfer", // Assume transfer port
				"dstPort": "transfer",
				"totalPackets": ch.StuckCount,
				"effectedPackets": 0, // TODO: Get from metrics
				"uneffectedPackets": ch.StuckCount,
				"successRate": 0.0, // TODO: Calculate
				"status": "congested",
			})
		}
	}
	
	// Recent packets
	recentPackets := []gin.H{}
	if stuckPacketsResp != nil && len(stuckPacketsResp.Packets) > 0 {
		limit := 5
		if len(stuckPacketsResp.Packets) < limit {
			limit = len(stuckPacketsResp.Packets)
		}
		for i := 0; i < limit; i++ {
			packet := stuckPacketsResp.Packets[i]
			recentPackets = append(recentPackets, gin.H{
				"sequence": packet.Sequence,
				"srcChannel": packet.SrcChannel,
				"dstChannel": packet.DstChannel,
				"amount": packet.Amount,
				"denom": packet.Denom,
				"sender": packet.Sender,
				"receiver": packet.Receiver,
				"status": "stuck",
				"ageSeconds": packet.AgeSeconds,
			})
		}
	}
	
	totalStuckPackets := 0
	if stuckPacketsResp != nil {
		totalStuckPackets = stuckPacketsResp.Total
	}
	
	metrics := gin.H{
		"system": gin.H{
			"totalChains": len(chains),
			"totalTransactions": 0, // TODO: Get from metrics
			"totalPackets": totalStuckPackets,
			"totalErrors": 0, // TODO: Get from metrics
			"uptime": 99.8, // TODO: Calculate
			"lastSync": time.Now(),
		},
		"chains": chains,
		"relayers": relayers,
		"channels": channels,
		"recentPackets": recentPackets,
		"stuckPackets": recentPackets, // Same data for now
		"frontrunEvents": []gin.H{}, // TODO: Parse from metrics
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}

// GetPacketFlowMetrics returns packet flow data
func (h *Handler) GetPacketFlowMetrics(c *gin.Context) {
	// Get congestion data to build flows
	congestionResp, err := h.chainpulseClient.GetChannelCongestion()
	if err != nil {
		log.Printf("Error fetching channel congestion: %v", err)
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	
	registry := config.DefaultChainRegistry()
	flows := []gin.H{}
	
	// Build flows from congestion data
	flowMap := make(map[string]gin.H)
	if congestionResp != nil {
		for _, ch := range congestionResp.Channels {
			// Try to determine source chain from packets
			srcChain := "Unknown"
			dstChain := "Unknown"
			
			// Get stuck packets to find chain info
			packetsResp, _ := h.chainpulseClient.GetStuckPackets(0, 10)
			if packetsResp != nil {
				for _, packet := range packetsResp.Packets {
					if packet.SrcChannel == ch.SrcChannel {
						srcChain = packet.ChainID
						dstChain = getDestinationChainFromChannel(packet.ChainID, packet.SrcChannel)
						
						// Get chain names
						if srcConfig, ok := registry.Chains[srcChain]; ok {
							srcChain = srcConfig.ChainName
						}
						if dstConfig, ok := registry.Chains[dstChain]; ok {
							dstChain = dstConfig.ChainName
						}
						break
					}
				}
			}
			
			flowKey := fmt.Sprintf("%s->%s", srcChain, dstChain)
			volume := int64(0)
			for _, v := range ch.TotalValue {
				// Parse value string to int
				if val, err := fmt.Sscanf(v, "%d", &volume); err == nil && val > 0 {
					break
				}
			}
			
			if flow, exists := flowMap[flowKey]; exists {
				flow["packetCount"] = flow["packetCount"].(int) + ch.StuckCount
				flow["volume"] = flow["volume"].(int64) + volume
			} else {
				flowMap[flowKey] = gin.H{
					"sourceChain": srcChain,
					"targetChain": dstChain,
					"packetCount": ch.StuckCount,
					"volume": volume,
					"avgPacketSize": 0,
				}
			}
		}
	}
	
	// Convert map to slice
	for _, flow := range flowMap {
		if flow["packetCount"].(int) > 0 && flow["volume"].(int64) > 0 {
			flow["avgPacketSize"] = flow["volume"].(int64) / int64(flow["packetCount"].(int))
		}
		flows = append(flows, flow)
	}

	c.JSON(http.StatusOK, flows)
}

// GetStuckPackets returns analytics for stuck packets
func (h *Handler) GetStuckPackets(c *gin.Context) {
	// Get channel congestion data
	congestionResp, err := h.chainpulseClient.GetChannelCongestion()
	if err != nil {
		log.Printf("Error fetching channel congestion: %v", err)
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	
	stuckPackets := []gin.H{}
	
	if congestionResp != nil {
		for _, ch := range congestionResp.Channels {
			// Get some packets to determine chain info
			srcChain := "unknown"
			dstChain := "unknown"
			
			packetsResp, _ := h.chainpulseClient.GetStuckPackets(0, 10)
			if packetsResp != nil {
				for _, packet := range packetsResp.Packets {
					if packet.SrcChannel == ch.SrcChannel {
						srcChain = packet.ChainID
						dstChain = getDestinationChainFromChannel(packet.ChainID, packet.SrcChannel)
						break
					}
				}
			}
			
			// Calculate total value
			totalValue := int64(0)
			for _, v := range ch.TotalValue {
				if val, err := fmt.Sscanf(v, "%d", &totalValue); err == nil && val > 0 {
					break
				}
			}
			
			stuckPackets = append(stuckPackets, gin.H{
				"channelId": ch.SrcChannel,
				"sourceChain": srcChain,
				"destinationChain": dstChain,
				"stuckCount": ch.StuckCount,
				"totalValue": totalValue,
				"avgStuckTime": ch.OldestStuckAgeSeconds / 2, // Estimate average
				"oldestPacketAge": ch.OldestStuckAgeSeconds,
			})
		}
	}

	c.JSON(http.StatusOK, stuckPackets)
}

// GetRelayerPerformance returns relayer performance analytics
func (h *Handler) GetRelayerPerformance(c *gin.Context) {
	// Get stuck packets to analyze relayer performance
	stuckPacketsResp, err := h.chainpulseClient.GetStuckPackets(0, 1000) // Get more for stats
	if err != nil {
		log.Printf("Error fetching stuck packets: %v", err)
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	
	performance := []gin.H{}
	relayerStats := make(map[string]gin.H)
	
	if stuckPacketsResp != nil {
		for _, packet := range stuckPacketsResp.Packets {
			if packet.LastAttemptBy != "" {
				if _, exists := relayerStats[packet.LastAttemptBy]; !exists {
					relayerStats[packet.LastAttemptBy] = gin.H{
						"address": packet.LastAttemptBy,
						"packetCount": 0,
						"successRate": 0.0, // All stuck packets = 0% success
						"avgRelayTime": 0,
						"frontrunRate": 0.0,
						"gasEfficiency": 0.0,
						"uptime": 0.0,
						"isNew": false,
						"totalRelayTime": int64(0),
					}
				}
				
				stats := relayerStats[packet.LastAttemptBy]
				stats["packetCount"] = stats["packetCount"].(int) + packet.RelayAttempts
				stats["totalRelayTime"] = stats["totalRelayTime"].(int64) + packet.AgeSeconds
			}
		}
	}
	
	// Calculate averages and convert to slice
	for _, stats := range relayerStats {
		packetCount := stats["packetCount"].(int)
		if packetCount > 0 {
			stats["avgRelayTime"] = stats["totalRelayTime"].(int64) / int64(packetCount)
		}
		delete(stats, "totalRelayTime") // Remove temporary field
		
		// Mark as new if low packet count
		if packetCount < 100 {
			stats["isNew"] = true
		}
		
		performance = append(performance, stats)
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