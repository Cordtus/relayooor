package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// EnrichedPacket contains all available data about a packet from multiple sources
type EnrichedPacket struct {
	// Core packet data
	ID               string    `json:"id"`
	Sequence         int       `json:"sequence"`
	ChannelID        string    `json:"channel_id"`
	PortID           string    `json:"port_id"`
	SourceChain      string    `json:"source_chain"`
	DestinationChain string    `json:"destination_chain"`
	
	// Transaction details
	Amount    string    `json:"amount"`
	Denom     string    `json:"denom"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Timestamp time.Time `json:"timestamp"`
	TxHash    string    `json:"tx_hash,omitempty"`
	
	// Status information
	Status         string    `json:"status"`
	StuckDuration  string    `json:"stuck_duration,omitempty"`
	RelayAttempts  int       `json:"relay_attempts"`
	LastAttemptBy  string    `json:"last_attempt_by,omitempty"`
	LastAttemptAt  time.Time `json:"last_attempt_at,omitempty"`
	
	// Chain information
	SourceChainInfo struct {
		ChainName   string `json:"chain_name"`
		ChainID     string `json:"chain_id"`
		RPCEndpoint string `json:"rpc_endpoint,omitempty"`
		Height      int64  `json:"current_height,omitempty"`
	} `json:"source_chain_info"`
	
	DestinationChainInfo struct {
		ChainName   string `json:"chain_name"`
		ChainID     string `json:"chain_id"`
		RPCEndpoint string `json:"rpc_endpoint,omitempty"`
		Height      int64  `json:"current_height,omitempty"`
	} `json:"destination_chain_info"`
	
	// Channel information
	ChannelInfo struct {
		SourceChannel      string `json:"source_channel"`
		DestinationChannel string `json:"destination_channel"`
		ConnectionID       string `json:"connection_id,omitempty"`
		State              string `json:"state"`
		Version            string `json:"version"`
	} `json:"channel_info"`
	
	// Token information
	TokenInfo struct {
		Symbol       string  `json:"symbol"`
		Decimals     int     `json:"decimals"`
		USDValue     float64 `json:"usd_value,omitempty"`
		IsIBCToken   bool    `json:"is_ibc_token"`
		BaseDenom    string  `json:"base_denom,omitempty"`
		TracePath    string  `json:"trace_path,omitempty"`
	} `json:"token_info"`
	
	// Clearing information
	ClearingInfo struct {
		CanClear       bool     `json:"can_clear"`
		EstimatedGas   int      `json:"estimated_gas,omitempty"`
		EstimatedFee   string   `json:"estimated_fee,omitempty"`
		RequiredWallet string   `json:"required_wallet,omitempty"`
		ClearingSteps  []string `json:"clearing_steps,omitempty"`
	} `json:"clearing_info"`
	
	// Metrics
	Metrics struct {
		ProcessingTime  int    `json:"processing_time_seconds,omitempty"`
		DataSources     []string `json:"data_sources"`
		LastUpdated     time.Time `json:"last_updated"`
	} `json:"metrics"`
}

// PacketEnrichmentService combines data from multiple sources
type PacketEnrichmentService struct {
	chainpulseURL string
	hermesURL     string
	cache         sync.Map
	cacheTTL      time.Duration
}

// NewPacketEnrichmentService creates a new enrichment service
func NewPacketEnrichmentService(chainpulseURL, hermesURL string) *PacketEnrichmentService {
	return &PacketEnrichmentService{
		chainpulseURL: chainpulseURL,
		hermesURL:     hermesURL,
		cacheTTL:      5 * time.Minute,
	}
}

// EnrichPacket combines data from all available sources
func (s *PacketEnrichmentService) EnrichPacket(basicPacket map[string]interface{}) (*EnrichedPacket, error) {
	enriched := &EnrichedPacket{
		Metrics: struct {
			ProcessingTime int       `json:"processing_time_seconds,omitempty"`
			DataSources    []string  `json:"data_sources"`
			LastUpdated    time.Time `json:"last_updated"`
		}{
			DataSources: []string{"chainpulse"},
			LastUpdated: time.Now(),
		},
	}
	
	// Extract basic packet data
	if id, ok := basicPacket["id"].(string); ok {
		enriched.ID = id
	}
	if seq, ok := basicPacket["sequence"].(float64); ok {
		enriched.Sequence = int(seq)
	}
	if channelID, ok := basicPacket["channelId"].(string); ok {
		enriched.ChannelID = channelID
	}
	if sourceChain, ok := basicPacket["sourceChain"].(string); ok {
		enriched.SourceChain = sourceChain
	}
	if destChain, ok := basicPacket["destinationChain"].(string); ok {
		enriched.DestinationChain = destChain
	}
	if amount, ok := basicPacket["amount"].(string); ok {
		enriched.Amount = amount
	}
	if denom, ok := basicPacket["denom"].(string); ok {
		enriched.Denom = denom
	}
	if sender, ok := basicPacket["sender"].(string); ok {
		enriched.Sender = sender
	}
	if receiver, ok := basicPacket["receiver"].(string); ok {
		enriched.Receiver = receiver
	}
	if stuckDur, ok := basicPacket["stuckDuration"].(string); ok {
		enriched.StuckDuration = stuckDur
		enriched.Status = "stuck"
	}
	
	// Set default port
	enriched.PortID = "transfer"
	
	// Enrich chain information
	s.enrichChainInfo(enriched)
	
	// Enrich channel information
	s.enrichChannelInfo(enriched)
	
	// Enrich token information
	s.enrichTokenInfo(enriched)
	
	// Calculate clearing information
	s.enrichClearingInfo(enriched)
	
	return enriched, nil
}

// enrichChainInfo adds chain metadata
func (s *PacketEnrichmentService) enrichChainInfo(packet *EnrichedPacket) {
	// Source chain info
	packet.SourceChainInfo.ChainID = packet.SourceChain
	packet.SourceChainInfo.ChainName = getChainName(packet.SourceChain)
	
	// Destination chain info
	packet.DestinationChainInfo.ChainID = packet.DestinationChain
	packet.DestinationChainInfo.ChainName = getChainName(packet.DestinationChain)
	
	// Try to get chain info from Hermes
	if s.hermesURL != "" {
		// Check if source chain is in Hermes
		if chainInfo := s.getHermesChainInfo(packet.SourceChain); chainInfo != nil {
			if rpcAddr, ok := chainInfo["rpc_addr"].(string); ok {
				packet.SourceChainInfo.RPCEndpoint = rpcAddr
			}
		}
	}
}

// enrichChannelInfo adds channel metadata
func (s *PacketEnrichmentService) enrichChannelInfo(packet *EnrichedPacket) {
	packet.ChannelInfo.SourceChannel = packet.ChannelID
	packet.ChannelInfo.State = "OPEN" // Default assumption
	packet.ChannelInfo.Version = "ics20-1"
	
	// Map known channel pairs
	channelPairs := map[string]map[string]string{
		"osmosis-1": {
			"channel-0":   "channel-141", // to cosmoshub-4
			"channel-141": "channel-326", // to cosmoshub-4
			"channel-208": "channel-5",   // to axelar
		},
		"cosmoshub-4": {
			"channel-141": "channel-0",   // to osmosis-1
			"channel-326": "channel-141", // to osmosis-1
		},
	}
	
	if channels, ok := channelPairs[packet.SourceChain]; ok {
		if destChannel, ok := channels[packet.ChannelID]; ok {
			packet.ChannelInfo.DestinationChannel = destChannel
		}
	}
}

// enrichTokenInfo adds token metadata
func (s *PacketEnrichmentService) enrichTokenInfo(packet *EnrichedPacket) {
	packet.TokenInfo.IsIBCToken = strings.HasPrefix(packet.Denom, "ibc/") || strings.Contains(packet.Denom, "transfer/")
	
	// Extract token symbol
	if strings.Contains(packet.Denom, "uatom") {
		packet.TokenInfo.Symbol = "ATOM"
		packet.TokenInfo.Decimals = 6
	} else if strings.Contains(packet.Denom, "uosmo") {
		packet.TokenInfo.Symbol = "OSMO"
		packet.TokenInfo.Decimals = 6
	} else if strings.Contains(packet.Denom, "uusdc") {
		packet.TokenInfo.Symbol = "USDC"
		packet.TokenInfo.Decimals = 6
	} else if strings.Contains(packet.Denom, "ustrd") {
		packet.TokenInfo.Symbol = "STRD"
		packet.TokenInfo.Decimals = 6
	}
	
	if packet.TokenInfo.IsIBCToken {
		// Extract base denom from IBC denom
		parts := strings.Split(packet.Denom, "/")
		if len(parts) > 2 {
			packet.TokenInfo.BaseDenom = parts[len(parts)-1]
			packet.TokenInfo.TracePath = strings.Join(parts[:len(parts)-1], "/")
		}
	}
}

// enrichClearingInfo adds packet clearing requirements
func (s *PacketEnrichmentService) enrichClearingInfo(packet *EnrichedPacket) {
	packet.ClearingInfo.CanClear = packet.Status == "stuck"
	packet.ClearingInfo.EstimatedGas = 150000
	
	// Estimate fee based on chain
	switch packet.SourceChain {
	case "cosmoshub-4":
		packet.ClearingInfo.EstimatedFee = "3750uatom" // 150000 * 0.025
		packet.ClearingInfo.RequiredWallet = "cosmos"
	case "osmosis-1":
		packet.ClearingInfo.EstimatedFee = "3750uosmo"
		packet.ClearingInfo.RequiredWallet = "osmo"
	case "noble-1":
		packet.ClearingInfo.EstimatedFee = "15000uusdc" // 150000 * 0.1
		packet.ClearingInfo.RequiredWallet = "noble"
	}
	
	packet.ClearingInfo.ClearingSteps = []string{
		"1. Verify packet is truly stuck",
		"2. Check relayer wallet has sufficient balance",
		"3. Submit RecvPacket transaction",
		"4. Wait for confirmation",
		"5. Verify packet cleared",
	}
}

// getHermesChainInfo queries Hermes for chain configuration
func (s *PacketEnrichmentService) getHermesChainInfo(chainID string) map[string]interface{} {
	resp, err := http.Get(fmt.Sprintf("%s/chain/%s", s.hermesURL, chainID))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	
	if data, ok := result["result"].(map[string]interface{}); ok {
		return data
	}
	return nil
}

// Helper function to get chain names
func getChainName(chainID string) string {
	chainNames := map[string]string{
		"cosmoshub-4":    "Cosmos Hub",
		"osmosis-1":      "Osmosis",
		"noble-1":        "Noble",
		"neutron-1":      "Neutron",
		"stride-1":       "Stride",
		"axelar-dojo-1":  "Axelar",
		"jackal-1":       "Jackal",
		"chihuahua-1":    "Chihuahua",
		"core-1":         "Persistence",
	}
	
	if name, ok := chainNames[chainID]; ok {
		return name
	}
	return chainID
}