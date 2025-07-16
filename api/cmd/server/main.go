package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type MetricsResponse struct {
	StuckPackets    int     `json:"stuckPackets"`
	ActiveChannels  int     `json:"activeChannels"`
	PacketFlowRate  float64 `json:"packetFlowRate"`
	SuccessRate     float64 `json:"successRate"`
}

type Channel struct {
	ChannelID             string `json:"channelId"`
	CounterpartyChannelID string `json:"counterpartyChannelId"`
	SourceChain           string `json:"sourceChain"`
	DestinationChain      string `json:"destinationChain"`
	State                 string `json:"state"`
	PendingPackets        int    `json:"pendingPackets"`
	TotalPackets          int    `json:"totalPackets"`
}

type StuckPacket struct {
	ID               string    `json:"id"`
	ChannelID        string    `json:"channelId"`
	Sequence         int       `json:"sequence"`
	SourceChain      string    `json:"sourceChain"`
	DestinationChain string    `json:"destinationChain"`
	StuckDuration    string    `json:"stuckDuration"`
	Amount           string    `json:"amount"`
	Denom            string    `json:"denom"`
	Sender           string    `json:"sender"`
	Receiver         string    `json:"receiver"`
	Timestamp        time.Time `json:"timestamp"`
	TxHash           string    `json:"txHash"`
	RelayAttempts    int       `json:"relayAttempts,omitempty"`
	IBCVersion       string    `json:"ibcVersion,omitempty"`
	LastAttemptBy    string    `json:"lastAttemptBy,omitempty"`
}

type UserTransfer struct {
	ID               string    `json:"id"`
	ChannelID        string    `json:"channelId"`
	Sequence         int       `json:"sequence"`
	SourceChain      string    `json:"sourceChain"`
	DestinationChain string    `json:"destinationChain"`
	Amount           string    `json:"amount"`
	Denom            string    `json:"denom"`
	Sender           string    `json:"sender"`
	Receiver         string    `json:"receiver"`
	Status           string    `json:"status"` // "pending", "stuck", "completed"
	Timestamp        time.Time `json:"timestamp"`
	TxHash           string    `json:"txHash"`
	StuckDuration    *string   `json:"stuckDuration,omitempty"`
}

type ClearPacketRequest struct {
	PacketIDs []string `json:"packetIds"`
	Wallet    string   `json:"wallet"`
	Signature string   `json:"signature"`
}

type ClearPacketResponse struct {
	Status    string   `json:"status"`
	TxHash    string   `json:"txHash,omitempty"`
	Cleared   []string `json:"cleared,omitempty"`
	Failed    []string `json:"failed,omitempty"`
	Message   string   `json:"message,omitempty"`
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(loggingMiddleware)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}).Methods("GET")

	// API endpoints
	api := r.PathPrefix("/api").Subrouter()

	// Metrics endpoint
	api.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := MetricsResponse{
			StuckPackets:   0,
			ActiveChannels: 2,
			PacketFlowRate: 12.5,
			SuccessRate:    98.5,
		}
		json.NewEncoder(w).Encode(metrics)
	}).Methods("GET")

	// Channels endpoint
	api.HandleFunc("/channels", func(w http.ResponseWriter, r *http.Request) {
		channels := []Channel{
			{
				ChannelID:             "channel-0",
				CounterpartyChannelID: "channel-141",
				SourceChain:           "osmosis-1",
				DestinationChain:      "cosmoshub-4",
				State:                 "OPEN",
				PendingPackets:        0,
				TotalPackets:          15234,
			},
			{
				ChannelID:             "channel-141",
				CounterpartyChannelID: "channel-0",
				SourceChain:           "cosmoshub-4",
				DestinationChain:      "osmosis-1",
				State:                 "OPEN",
				PendingPackets:        0,
				TotalPackets:          18765,
			},
		}
		json.NewEncoder(w).Encode(channels)
	}).Methods("GET")

	// Stuck packets endpoint (global view)
	api.HandleFunc("/packets/stuck", func(w http.ResponseWriter, r *http.Request) {
		// Try to get data from Chainpulse
		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3000"
		}

		// Forward query parameters
		queryParams := r.URL.Query().Encode()
		if queryParams == "" {
			queryParams = "min_age_seconds=900" // Default 15 minutes
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/packets/stuck?%s", chainpulseURL, queryParams))
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			
			// Parse Chainpulse response
			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
				if packets, ok := data["packets"].([]interface{}); ok {
					stuckPackets := make([]StuckPacket, 0)
					for _, p := range packets {
						if packet, ok := p.(map[string]interface{}); ok {
							sp := StuckPacket{
								ID:               fmt.Sprintf("%v-%v", packet["chain_id"], packet["sequence"]),
								ChannelID:        fmt.Sprintf("%v", packet["src_channel"]),
								Sequence:         int(getFloat64(packet["sequence"])),
								SourceChain:      fmt.Sprintf("%v", packet["chain_id"]),
								DestinationChain: fmt.Sprintf("%v", packet["dst_channel"]),
								StuckDuration:    formatDuration(int(getFloat64(packet["age_seconds"]))),
								Amount:           fmt.Sprintf("%v", packet["amount"]),
								Denom:            fmt.Sprintf("%v", packet["denom"]),
								Sender:           fmt.Sprintf("%v", packet["sender"]),
								Receiver:         fmt.Sprintf("%v", packet["receiver"]),
								TxHash:           fmt.Sprintf("%v", packet["tx_hash"]),
							}
							
							// Calculate timestamp from age
							if ageSeconds := getFloat64(packet["age_seconds"]); ageSeconds > 0 {
								sp.Timestamp = time.Now().Add(-time.Duration(ageSeconds) * time.Second)
							}
							
							// Add relay attempts if available
							if attempts, ok := packet["relay_attempts"]; ok {
								sp.RelayAttempts = int(getFloat64(attempts))
							}
							
							stuckPackets = append(stuckPackets, sp)
						}
					}
					
					json.NewEncoder(w).Encode(stuckPackets)
					return
				}
			}
		}
		
		// Fallback to empty array if Chainpulse is not available
		packets := []StuckPacket{}
		json.NewEncoder(w).Encode(packets)
	}).Methods("GET")

	// User transfers endpoint - shows all transfers for a wallet
	api.HandleFunc("/user/{wallet}/transfers", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		wallet := vars["wallet"]
		
		// Validate wallet address format
		if !isValidWalletAddress(wallet) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid wallet address"})
			return
		}

		// Try to get data from Chainpulse
		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3000"
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/packets/by-user?address=%s", chainpulseURL, wallet))
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			
			// Parse Chainpulse response
			var packets []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&packets); err == nil {
				// Transform to UserTransfer format
				transfers := make([]UserTransfer, 0)
				for _, p := range packets {
					transfer := UserTransfer{
						ID:               fmt.Sprintf("%v-%v", p["tx_hash"], p["sequence"]),
						ChannelID:        fmt.Sprintf("%v", p["channel"]),
						Sequence:         int(getFloat64(p["sequence"])),
						SourceChain:      fmt.Sprintf("%v", p["chain"]),
						DestinationChain: getCounterpartyChain(fmt.Sprintf("%v", p["chain"])),
						Amount:           fmt.Sprintf("%v", p["amount"]),
						Denom:            fmt.Sprintf("%v", p["denom"]),
						Sender:           fmt.Sprintf("%v", p["sender"]),
						Receiver:         fmt.Sprintf("%v", p["receiver"]),
						Status:           fmt.Sprintf("%v", p["status"]),
						TxHash:           fmt.Sprintf("%v", p["tx_hash"]),
					}
					
					// Parse timestamp
					if ts, ok := p["timestamp"].(string); ok {
						if t, err := time.Parse(time.RFC3339, ts); err == nil {
							transfer.Timestamp = t
						}
					}
					
					// Add stuck duration if stuck
					if transfer.Status == "stuck" {
						duration := time.Since(transfer.Timestamp).Round(time.Minute).String()
						transfer.StuckDuration = &duration
					}
					
					transfers = append(transfers, transfer)
				}
				
				json.NewEncoder(w).Encode(transfers)
				return
			}
		}

		// Fallback to mock data if Chainpulse is not available
		transfers := []UserTransfer{
			{
				ID:               "transfer-1",
				ChannelID:        "channel-0",
				Sequence:         12345,
				SourceChain:      "osmosis-1",
				DestinationChain: "cosmoshub-4",
				Amount:           "1000000",
				Denom:            "uosmo",
				Sender:           wallet,
				Receiver:         convertAddress(wallet, "cosmos"),
				Status:           "stuck",
				Timestamp:        time.Now().Add(-2 * time.Hour),
				TxHash:           "ABC123DEF456",
				StuckDuration:    stringPtr("2h"),
			},
		}

		json.NewEncoder(w).Encode(transfers)
	}).Methods("GET")

	// User stuck packets endpoint - shows only stuck transfers
	api.HandleFunc("/user/{wallet}/stuck", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		wallet := vars["wallet"]
		
		if !isValidWalletAddress(wallet) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid wallet address"})
			return
		}

		// Return only stuck transfers
		stuckTransfers := []UserTransfer{}
		json.NewEncoder(w).Encode(stuckTransfers)
	}).Methods("GET")

	// Clear packets endpoint
	api.HandleFunc("/packets/clear", func(w http.ResponseWriter, r *http.Request) {
		var req ClearPacketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Validate wallet signature
		if !verifyWalletSignature(req.Wallet, req.Signature) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid signature"})
			return
		}

		// Mock clearing response
		resp := ClearPacketResponse{
			Status:  "success",
			TxHash:  "CLEAR_TX_" + generateTxHash(),
			Cleared: req.PacketIDs,
			Failed:  []string{},
			Message: "Packets cleared successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}).Methods("POST")

	// Chainpulse metrics endpoint (Prometheus format)
	api.HandleFunc("/metrics/chainpulse", func(w http.ResponseWriter, r *http.Request) {
		// Get Chainpulse URL from environment or use default
		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3000"
		}
		
		resp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
		if err != nil {
			log.Printf("Failed to fetch from Chainpulse at %s: %v", chainpulseURL, err)
			// Fall back to mock data if Chainpulse is not available
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			metrics := generateMockPrometheusMetrics()
			fmt.Fprint(w, metrics)
			return
		}
		defer resp.Body.Close()
		
		// Copy headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		
		// Copy status code
		w.WriteHeader(resp.StatusCode)
		
		// Copy body
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("Error copying response body: %v", err)
		}
	}).Methods("GET")

	// Configuration endpoint
	api.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		registry := getChainRegistry()
		json.NewEncoder(w).Encode(registry)
	}).Methods("GET")

	// Structured monitoring data endpoint
	api.HandleFunc("/monitoring/data", func(w http.ResponseWriter, r *http.Request) {
		data := getStructuredMonitoringData()
		json.NewEncoder(w).Encode(data)
	}).Methods("GET")

	// Channel congestion endpoint
	api.HandleFunc("/channels/congestion", func(w http.ResponseWriter, r *http.Request) {
		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3000"
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/channels/congestion", chainpulseURL))
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			
			// Parse and forward the response
			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
				json.NewEncoder(w).Encode(data)
				return
			}
		}
		
		// Fallback to mock data
		congestion := map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"src_channel": "channel-141",
					"dst_channel": "channel-0",
					"stuck_count": 2,
					"oldest_stuck_age_seconds": 3600,
					"total_value": map[string]string{
						"uatom": "50000000",
						"uosmo": "100000000",
					},
				},
			},
			"api_version": "1.0",
		}
		json.NewEncoder(w).Encode(congestion)
	}).Methods("GET")

	// Packet details endpoint
	api.HandleFunc("/packets/{chain}/{channel}/{sequence}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		chain := vars["chain"]
		channel := vars["channel"]
		sequence := vars["sequence"]

		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3000"
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/packets/%s/%s/%s", chainpulseURL, chain, channel, sequence))
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			
			// Forward the response
			var data map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
				json.NewEncoder(w).Encode(data)
				return
			}
		}
		
		// Return 404 if not found
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Packet not found"})
	}).Methods("GET")

	// Chain registry endpoint - returns known chain information
	api.HandleFunc("/chains/registry", func(w http.ResponseWriter, r *http.Request) {
		// Enhanced chain registry with more chains and RPC endpoints
		registry := map[string]interface{}{
			"chains": []map[string]interface{}{
				{
					"chain_id": "cosmoshub-4",
					"chain_name": "Cosmos Hub",
					"pretty_name": "Cosmos Hub",
					"network_type": "mainnet",
					"prefix": "cosmos",
					"rest_api": "https://cosmos-rest.publicnode.com",
					"rpc": "https://rpc.cosmos.network:443",
					"comet_version": "0.34",
				},
				{
					"chain_id": "osmosis-1",
					"chain_name": "Osmosis",
					"pretty_name": "Osmosis",
					"network_type": "mainnet",
					"prefix": "osmo",
					"rest_api": "https://lcd.osmosis.zone",
					"rpc": "https://rpc.osmosis.zone:443",
					"comet_version": "0.38",
				},
				{
					"chain_id": "neutron-1",
					"chain_name": "Neutron",
					"pretty_name": "Neutron",
					"network_type": "mainnet",
					"prefix": "neutron",
					"rest_api": "https://rest-kralum.neutron-1.neutron.org",
					"rpc": "https://rpc-kralum.neutron-1.neutron.org:443",
					"comet_version": "0.37",
				},
				{
					"chain_id": "stride-1",
					"chain_name": "Stride",
					"pretty_name": "Stride",
					"network_type": "mainnet",
					"prefix": "stride",
					"rest_api": "https://stride-api.polkachu.com",
					"rpc": "https://stride-rpc.polkachu.com:443",
					"comet_version": "0.37",
				},
				{
					"chain_id": "noble-1",
					"chain_name": "Noble",
					"pretty_name": "Noble",
					"network_type": "mainnet",
					"prefix": "noble",
					"rest_api": "https://noble-api.polkachu.com",
					"rpc": "https://noble-rpc.polkachu.com:443",
					"comet_version": "0.38",
				},
				{
					"chain_id": "juno-1",
					"chain_name": "Juno",
					"pretty_name": "Juno",
					"network_type": "mainnet",
					"prefix": "juno",
					"rest_api": "https://juno-api.polkachu.com",
					"rpc": "https://juno-rpc.polkachu.com:443",
					"comet_version": "0.34",
				},
				{
					"chain_id": "axelar-dojo-1",
					"chain_name": "Axelar",
					"pretty_name": "Axelar",
					"network_type": "mainnet",
					"prefix": "axelar",
					"rest_api": "https://axelar-api.polkachu.com",
					"rpc": "https://axelar-rpc.polkachu.com:443",
					"comet_version": "0.34",
				},
				{
					"chain_id": "dydx-mainnet-1",
					"chain_name": "dYdX",
					"pretty_name": "dYdX",
					"network_type": "mainnet",
					"prefix": "dydx",
					"rest_api": "https://dydx-rest.publicnode.com",
					"rpc": "https://dydx-rpc.publicnode.com:443",
					"comet_version": "0.38",
				},
			},
		}
		
		json.NewEncoder(w).Encode(registry)
	}).Methods("GET")

	// Comprehensive monitoring metrics endpoint
	api.HandleFunc("/monitoring/metrics", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		// Try to get real data from Chainpulse
		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3001"
		}
		
		// Fetch and parse real metrics
		metricsResp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
		if err != nil || metricsResp.StatusCode != http.StatusOK {
			log.Printf("Failed to fetch from Chainpulse: %v", err)
			// Return error response
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "Chainpulse unavailable"})
			return
		}
		
		defer metricsResp.Body.Close()
		body, _ := io.ReadAll(metricsResp.Body)
		metricsBody := string(body)
		log.Printf("Successfully fetched metrics from Chainpulse")
		
		// Parse metrics using the comprehensive parser
		parsedData := parsePrometheusMetrics(metricsBody)
		
		// Extract parsed data
		chains := parsedData["chains"].([]map[string]interface{})
		channels := parsedData["channels"].([]map[string]interface{})
		
		// Calculate totals from real data
		totalChains := len(chains)
		totalPackets := 0
		totalTxs := 0
		totalErrors := 0
		
		for _, chain := range chains {
			if packets, ok := chain["packets_24h"].(int); ok {
				totalPackets += packets
			}
			if txs, ok := chain["txs_total"].(int); ok {
				totalTxs += txs
			}
			if errors, ok := chain["errors"].(int); ok {
				totalErrors += errors
			}
		}
		
		log.Printf("Real totals - Chains: %d, Packets: %d, Txs: %d, Errors: %d", totalChains, totalPackets, totalTxs, totalErrors)
		
		// Extract relayer data from packet metrics
		relayerMap := make(map[string]map[string]interface{})
		
		// Parse packet metrics to get relayer info
		lines := strings.Split(metricsBody, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ibc_effected_packets{") || strings.HasPrefix(line, "ibc_uneffected_packets{") {
				parts := strings.Split(line, " ")
				if len(parts) >= 2 {
					// Extract signer label
					labelRe := regexp.MustCompile(`signer="([^"]+)"`)
					matches := labelRe.FindStringSubmatch(parts[0])
					if len(matches) > 1 {
						signer := matches[1]
						value, _ := strconv.ParseFloat(parts[1], 64)
						
						if relayerMap[signer] == nil {
							relayerMap[signer] = map[string]interface{}{
								"address": signer,
								"effectedPackets": float64(0),
								"uneffectedPackets": float64(0),
								"totalPackets": float64(0),
							}
						}
						
						if strings.HasPrefix(line, "ibc_effected_packets{") {
							relayerMap[signer]["effectedPackets"] = relayerMap[signer]["effectedPackets"].(float64) + value
						} else {
							relayerMap[signer]["uneffectedPackets"] = relayerMap[signer]["uneffectedPackets"].(float64) + value
						}
						relayerMap[signer]["totalPackets"] = relayerMap[signer]["effectedPackets"].(float64) + relayerMap[signer]["uneffectedPackets"].(float64)
					}
				}
			}
		}
		
		// Convert relayer map to array
		relayers := []map[string]interface{}{}
		for _, relayer := range relayerMap {
			effected := relayer["effectedPackets"].(float64)
			total := relayer["totalPackets"].(float64)
			successRate := float64(0)
			if total > 0 {
				successRate = (effected / total) * 100
			}
			
			relayers = append(relayers, map[string]interface{}{
				"address": relayer["address"],
				"totalPackets": int(total),
				"effectedPackets": int(effected),
				"uneffectedPackets": int(relayer["uneffectedPackets"].(float64)),
				"successRate": successRate,
				"frontrunCount": 0, // TODO: Calculate from frontrun metrics
				"memo": "",
				"software": "unknown",
				"version": "unknown",
			})
		}
		
		// Sort relayers by total packets
		sort.Slice(relayers, func(i, j int) bool {
			return relayers[i]["totalPackets"].(int) > relayers[j]["totalPackets"].(int)
		})
		
		// Create comprehensive metrics response matching MetricsSnapshot interface
		now := time.Now()
		
		// Transform chains data to match expected format
		transformedChains := []map[string]interface{}{}
		for _, chain := range chains {
			transformedChains = append(transformedChains, map[string]interface{}{
				"chainId": chain["chain_id"],
				"chainName": chain["name"],
				"totalTxs": chain["txs_total"],
				"totalPackets": chain["packets_24h"],
				"reconnects": 0, // Not available in metrics
				"timeouts": 0,   // Not available in metrics
				"errors": chain["errors"],
				"status": chain["status"],
				"lastUpdate": now,
			})
		}
		
		response := map[string]interface{}{
			"system": map[string]interface{}{
				"totalChains": totalChains,
				"totalTransactions": totalTxs,
				"totalPackets": totalPackets,
				"totalErrors": totalErrors,
				"uptime": 86400, // 24 hours in seconds
				"lastSync": now,
			},
			"chains": transformedChains,
			"relayers": relayers,
			"channels": transformChannelsForMetrics(channels),
			"recentPackets": extractRecentPackets(lines, 10), // Extract last 10 packets
			"stuckPackets": []map[string]interface{}{},
			"frontrunEvents": []map[string]interface{}{}, // TODO: Extract from frontrun metrics
			"timestamp": now,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Monitoring data endpoint
	api.HandleFunc("/monitoring/data", func(w http.ResponseWriter, r *http.Request) {
		data := getStructuredMonitoringData()
		json.NewEncoder(w).Encode(data)
	}).Methods("GET")

	// Monitoring metrics endpoint
	api.HandleFunc("/monitoring/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := getMonitoringMetrics()
		json.NewEncoder(w).Encode(metrics)
	}).Methods("GET")

	// Platform statistics endpoint
	api.HandleFunc("/statistics/platform", func(w http.ResponseWriter, r *http.Request) {
		// Get real metrics from Chainpulse
		chainpulseURL := os.Getenv("CHAINPULSE_URL")
		if chainpulseURL == "" {
			chainpulseURL = "http://localhost:3001"
		}
		
		// Default values
		totalPackets := 0
		totalChains := 0
		avgSuccessRate := 95.0
		
		// Try to get real data
		metricsResp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
		if err == nil && metricsResp.StatusCode == http.StatusOK {
			defer metricsResp.Body.Close()
			body, _ := io.ReadAll(metricsResp.Body)
			parsedData := parsePrometheusMetrics(string(body))
			
			// Extract real totals
			if chains, ok := parsedData["chains"].([]map[string]interface{}); ok {
				totalChains = len(chains)
				for _, chain := range chains {
					if packets, ok := chain["packets_24h"].(int); ok {
						totalPackets += packets
					}
				}
			}
			
			// Calculate average success rate from channels
			if channels, ok := parsedData["channels"].([]map[string]interface{}); ok && len(channels) > 0 {
				totalRate := 0.0
				for _, channel := range channels {
					if rate, ok := channel["success_rate"].(float64); ok {
						totalRate += rate
					}
				}
				avgSuccessRate = totalRate / float64(len(channels))
			}
		}
		
		// Create response with mix of real and estimated data
		stats := map[string]interface{}{
			"global": map[string]interface{}{
				"totalPacketsCleared": totalPackets * 30, // Estimate 30 days of operation
				"totalUsers": totalChains * 50, // Estimate users per chain
				"totalFeesCollected": fmt.Sprintf("%d", totalPackets * 10), // Estimate $10 per packet
				"avgClearTime": 45,
				"successRate": avgSuccessRate,
			},
			"daily": map[string]interface{}{
				"packetsCleared": totalPackets,
				"activeUsers": totalChains * 2, // Estimate active users
				"feesCollected": fmt.Sprintf("%d", totalPackets * 10),
			},
			"topChannels": []map[string]interface{}{
				{
					"channel": "channel-0 → channel-141",
					"packetsCleared": totalPackets / 3, // Estimate distribution
					"avgClearTime": 42,
				},
				{
					"channel": "channel-141 → channel-0",
					"packetsCleared": totalPackets / 3,
					"avgClearTime": 45,
				},
			},
			"peakHours": []map[string]interface{}{
				{"hour": 14, "activity": totalPackets / 24 * 2}, // Peak at 2pm
				{"hour": 15, "activity": totalPackets / 24 * 2},
				{"hour": 16, "activity": totalPackets / 24 * 2},
			},
		}
		json.NewEncoder(w).Encode(stats)
	}).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost", "http://localhost:80", "http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API server starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler); err != nil {
		log.Fatal(err)
	}
}

// Helper functions
func isValidWalletAddress(address string) bool {
	// Basic validation - check if it's a valid bech32 address
	if strings.HasPrefix(address, "osmo1") || strings.HasPrefix(address, "cosmos1") {
		return len(address) > 10 && len(address) < 100
	}
	return false
}

func convertAddress(address, targetPrefix string) string {
	// Mock address conversion - in production use proper bech32 conversion
	if strings.HasPrefix(address, "osmo1") && targetPrefix == "cosmos" {
		return "cosmos1" + address[5:]
	}
	if strings.HasPrefix(address, "cosmos1") && targetPrefix == "osmo" {
		return "osmo1" + address[7:]
	}
	return address
}

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

func getFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

func verifyWalletSignature(wallet, signature string) bool {
	// Mock signature verification - in production use proper crypto verification
	return signature != "" && wallet != ""
}

func generateTxHash() string {
	// Mock tx hash generation
	return fmt.Sprintf("%d", time.Now().Unix())
}

func stringPtr(s string) *string {
	return &s
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	} else if seconds < 86400 {
		return fmt.Sprintf("%dh", seconds/3600)
	} else {
		return fmt.Sprintf("%dd", seconds/86400)
	}
}

func parsePrometheusMetrics(metricsText string) map[string]interface{} {
	// Parse Prometheus metrics to extract chain and channel data
	result := map[string]interface{}{
		"chains":   []map[string]interface{}{},
		"channels": []map[string]interface{}{},
		"relayers": []map[string]interface{}{},
	}
	
	// Track unique chains and their metrics
	chainMetrics := make(map[string]map[string]float64)
	channelMetrics := make(map[string]map[string]interface{})
	
	lines := strings.Split(metricsText, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		// Parse chainpulse_packets metric
		if strings.HasPrefix(line, "chainpulse_packets{") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				// Extract chain_id from label
				re := regexp.MustCompile(`chain_id="([^"]+)"`)
				matches := re.FindStringSubmatch(parts[0])
				if len(matches) > 1 {
					chainID := matches[1]
					value, _ := strconv.ParseFloat(parts[1], 64)
					if chainMetrics[chainID] == nil {
						chainMetrics[chainID] = make(map[string]float64)
					}
					chainMetrics[chainID]["packets"] = value
				}
			}
		}
		
		// Parse chainpulse_txs metric
		if strings.HasPrefix(line, "chainpulse_txs{") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				re := regexp.MustCompile(`chain_id="([^"]+)"`)
				matches := re.FindStringSubmatch(parts[0])
				if len(matches) > 1 {
					chainID := matches[1]
					value, _ := strconv.ParseFloat(parts[1], 64)
					if chainMetrics[chainID] == nil {
						chainMetrics[chainID] = make(map[string]float64)
					}
					chainMetrics[chainID]["txs"] = value
				}
			}
		}
		
		// Parse chainpulse_errors metric
		if strings.HasPrefix(line, "chainpulse_errors{") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				re := regexp.MustCompile(`chain_id="([^"]+)"`)
				matches := re.FindStringSubmatch(parts[0])
				if len(matches) > 1 {
					chainID := matches[1]
					value, _ := strconv.ParseFloat(parts[1], 64)
					if chainMetrics[chainID] == nil {
						chainMetrics[chainID] = make(map[string]float64)
					}
					chainMetrics[chainID]["errors"] = value
				}
			}
		}
		
		// Parse ibc_effected_packets metric for channel data
		if strings.HasPrefix(line, "ibc_effected_packets{") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				// Extract all labels
				labelRe := regexp.MustCompile(`(\w+)="([^"]+)"`)
				labelMatches := labelRe.FindAllStringSubmatch(parts[0], -1)
				
				labels := make(map[string]string)
				for _, match := range labelMatches {
					if len(match) > 2 {
						labels[match[1]] = match[2]
					}
				}
				
				if srcChannel, ok := labels["src_channel"]; ok {
					if dstChannel, ok := labels["dst_channel"]; ok {
						channelKey := fmt.Sprintf("%s-%s", srcChannel, dstChannel)
						value, _ := strconv.ParseFloat(parts[1], 64)
						
						if channelMetrics[channelKey] == nil {
							channelMetrics[channelKey] = make(map[string]interface{})
							channelMetrics[channelKey]["src_channel"] = srcChannel
							channelMetrics[channelKey]["dst_channel"] = dstChannel
							channelMetrics[channelKey]["chain_id"] = labels["chain_id"]
							channelMetrics[channelKey]["src_port"] = labels["src_port"]
							channelMetrics[channelKey]["dst_port"] = labels["dst_port"]
							channelMetrics[channelKey]["effected_packets"] = float64(0)
							channelMetrics[channelKey]["uneffected_packets"] = float64(0)
						}
						channelMetrics[channelKey]["effected_packets"] = channelMetrics[channelKey]["effected_packets"].(float64) + value
					}
				}
			}
		}
		
		// Parse ibc_uneffected_packets metric
		if strings.HasPrefix(line, "ibc_uneffected_packets{") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				labelRe := regexp.MustCompile(`(\w+)="([^"]+)"`)
				labelMatches := labelRe.FindAllStringSubmatch(parts[0], -1)
				
				labels := make(map[string]string)
				for _, match := range labelMatches {
					if len(match) > 2 {
						labels[match[1]] = match[2]
					}
				}
				
				if srcChannel, ok := labels["src_channel"]; ok {
					if dstChannel, ok := labels["dst_channel"]; ok {
						channelKey := fmt.Sprintf("%s-%s", srcChannel, dstChannel)
						value, _ := strconv.ParseFloat(parts[1], 64)
						
						if channelMetrics[channelKey] == nil {
							channelMetrics[channelKey] = make(map[string]interface{})
							channelMetrics[channelKey]["src_channel"] = srcChannel
							channelMetrics[channelKey]["dst_channel"] = dstChannel
							channelMetrics[channelKey]["chain_id"] = labels["chain_id"]
							channelMetrics[channelKey]["src_port"] = labels["src_port"]
							channelMetrics[channelKey]["dst_port"] = labels["dst_port"]
							channelMetrics[channelKey]["effected_packets"] = float64(0)
							channelMetrics[channelKey]["uneffected_packets"] = float64(0)
						}
						channelMetrics[channelKey]["uneffected_packets"] = channelMetrics[channelKey]["uneffected_packets"].(float64) + value
					}
				}
			}
		}
	}
	
	// Convert chainMetrics to result format
	chains := []map[string]interface{}{}
	registry := getChainRegistry()
	chainsData := registry["chains"].(map[string]interface{})
	
	for chainID, metrics := range chainMetrics {
		name := chainID
		if chainData, exists := chainsData[chainID]; exists {
			if chainMap, ok := chainData.(map[string]interface{}); ok {
				if chainName, ok := chainMap["chain_name"].(string); ok {
					name = chainName
				}
			}
		}
		
		chain := map[string]interface{}{
			"chain_id":    chainID,
			"name":        name,
			"status":      "connected",
			"packets_24h": int(metrics["packets"]),
			"txs_total":   int(metrics["txs"]),
			"errors":      int(metrics["errors"]),
		}
		
		// Mark as degraded if there are errors
		if metrics["errors"] > 0 {
			chain["status"] = "degraded"
		}
		
		chains = append(chains, chain)
	}
	
	// Convert channelMetrics to result format
	channels := []map[string]interface{}{}
	for _, channel := range channelMetrics {
		effected := channel["effected_packets"].(float64)
		uneffected := channel["uneffected_packets"].(float64)
		total := effected + uneffected
		
		successRate := float64(0)
		if total > 0 {
			successRate = (effected / total) * 100
		}
		
		channelData := map[string]interface{}{
			"src":             channel["chain_id"],
			"src_channel":     channel["src_channel"],
			"dst_channel":     channel["dst_channel"],
			"src_port":        channel["src_port"],
			"dst_port":        channel["dst_port"],
			"status":          "active",
			"packets_pending": 0, // This would need separate metric
			"success_rate":    successRate,
			"total_packets":   int(total),
			"effected":        int(effected),
			"uneffected":      int(uneffected),
		}
		
		// Determine status based on success rate
		if successRate < 50 {
			channelData["status"] = "degraded"
		} else if total == 0 {
			channelData["status"] = "idle"
		}
		
		channels = append(channels, channelData)
	}
	
	result["chains"] = chains
	result["channels"] = channels
	return result
}

// transformChannelsForMetrics transforms channel data from monitoring/data format to monitoring/metrics format
func transformChannelsForMetrics(channels []map[string]interface{}) []map[string]interface{} {
	transformed := []map[string]interface{}{}
	
	for _, channel := range channels {
		// Get source chain from channel data
		srcChain := ""
		if src, ok := channel["src"].(string); ok {
			srcChain = src
		} else if chainId, ok := channel["chain_id"].(string); ok {
			srcChain = chainId
		}
		
		// Infer destination chain from channel numbers
		dstChain := inferDestinationChain(srcChain, channel["src_channel"].(string))
		
		transformed = append(transformed, map[string]interface{}{
			"srcChain": srcChain,
			"dstChain": dstChain,
			"srcChannel": channel["src_channel"],
			"dstChannel": channel["dst_channel"],
			"srcPort": channel["src_port"],
			"dstPort": channel["dst_port"],
			"totalPackets": channel["total_packets"],
			"effectedPackets": channel["effected"],
			"uneffectedPackets": channel["uneffected"],
			"successRate": channel["success_rate"],
			"avgProcessingTime": 45.0, // Default since not in metrics
			"status": channel["status"],
		})
	}
	
	return transformed
}

// inferDestinationChain attempts to determine the destination chain based on known channel mappings
func inferDestinationChain(srcChain, srcChannel string) string {
	// Known channel mappings
	channelMappings := map[string]map[string]string{
		"cosmoshub-4": {
			"channel-141": "osmosis-1",
			"channel-536": "noble-1",
			"channel-569": "neutron-1",
		},
		"osmosis-1": {
			"channel-0": "cosmoshub-4",
			"channel-750": "noble-1",
			"channel-874": "neutron-1",
		},
		"neutron-1": {
			"channel-1": "cosmoshub-4",
			"channel-10": "osmosis-1",
			"channel-30": "noble-1",
		},
		"noble-1": {
			"channel-1": "osmosis-1",
			"channel-4": "cosmoshub-4",
			"channel-18": "neutron-1",
		},
	}
	
	if chainChannels, ok := channelMappings[srcChain]; ok {
		if dstChain, ok := chainChannels[srcChannel]; ok {
			return dstChain
		}
	}
	
	// Default fallback
	return "unknown"
}

// getChainMap returns a map of chain IDs to chain names
func getChainMap() map[string]string {
	return map[string]string{
		"cosmoshub-4": "Cosmos Hub",
		"osmosis-1": "Osmosis",
		"neutron-1": "Neutron",
		"noble-1": "Noble",
		"akash-1": "Akash",
		"stargaze-1": "Stargaze",
		"juno-1": "Juno",
		"stride-1": "Stride",
		"axelar-dojo-1": "Axelar",
	}
}

// extractRecentPackets extracts recent packet information from metrics
func extractRecentPackets(metricsLines []string, limit int) []map[string]interface{} {
	packets := []map[string]interface{}{}
	
	// Extract packet info from effected/uneffected metrics
	for _, line := range metricsLines {
		if strings.HasPrefix(line, "ibc_effected_packets{") || strings.HasPrefix(line, "ibc_uneffected_packets{") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 && parts[1] != "0" {
				// Extract labels
				labelRe := regexp.MustCompile(`(\w+)="([^"]+)"`)
				labelMatches := labelRe.FindAllStringSubmatch(parts[0], -1)
				
				labels := make(map[string]string)
				for _, match := range labelMatches {
					if len(match) > 2 {
						labels[match[1]] = match[2]
					}
				}
				
				// Create packet entry
				packet := map[string]interface{}{
					"chain_id":    labels["chain_id"],
					"src_channel": labels["src_channel"],
					"dst_channel": labels["dst_channel"],
					"src_port":    labels["src_port"],
					"dst_port":    labels["dst_port"],
					"signer":      labels["signer"],
					"memo":        labels["memo"],
					"effected":    strings.HasPrefix(line, "ibc_effected_packets{"),
					"timestamp":   time.Now().Add(-time.Duration(len(packets)) * time.Minute), // Estimate timestamp
				}
				
				// Estimate sequence number (would need real data in production)
				packet["sequence"] = 100000 + len(packets)
				
				packets = append(packets, packet)
				
				if len(packets) >= limit {
					break
				}
			}
		}
	}
	
	return packets
}

func generateMockPrometheusMetrics() string {
	metrics := ""

	// System metrics
	metrics += "# HELP chainpulse_chains Number of chains being monitored\n"
	metrics += "# TYPE chainpulse_chains gauge\n"
	metrics += "chainpulse_chains 2\n\n"

	metrics += "# HELP chainpulse_txs Total number of transactions processed\n"
	metrics += "# TYPE chainpulse_txs counter\n"
	metrics += "chainpulse_txs{chain_id=\"cosmoshub-4\"} 12543\n"
	metrics += "chainpulse_txs{chain_id=\"osmosis-1\"} 18976\n\n"

	metrics += "# HELP chainpulse_packets Total number of packets processed\n"
	metrics += "# TYPE chainpulse_packets counter\n"
	metrics += "chainpulse_packets{chain_id=\"cosmoshub-4\"} 4532\n"
	metrics += "chainpulse_packets{chain_id=\"osmosis-1\"} 6789\n\n"

	// IBC packet metrics
	metrics += "# HELP ibc_effected_packets IBC packets effected (successfully relayed)\n"
	metrics += "# TYPE ibc_effected_packets counter\n"
	metrics += `ibc_effected_packets{chain_id="cosmoshub-4",src_channel="channel-141",src_port="transfer",dst_channel="channel-0",dst_port="transfer",signer="cosmos1abc123",memo=""} 856` + "\n"
	metrics += `ibc_effected_packets{chain_id="osmosis-1",src_channel="channel-0",src_port="transfer",dst_channel="channel-141",dst_port="transfer",signer="osmo1xyz789",memo=""} 1243` + "\n\n"

	metrics += "# HELP ibc_uneffected_packets IBC packets relayed but not effected (frontrun)\n"
	metrics += "# TYPE ibc_uneffected_packets counter\n"
	metrics += `ibc_uneffected_packets{chain_id="cosmoshub-4",src_channel="channel-141",src_port="transfer",dst_channel="channel-0",dst_port="transfer",signer="cosmos1abc123",memo=""} 124` + "\n"
	metrics += `ibc_uneffected_packets{chain_id="osmosis-1",src_channel="channel-0",src_port="transfer",dst_channel="channel-141",dst_port="transfer",signer="osmo1xyz789",memo=""} 189` + "\n\n"

	// Stuck packets
	metrics += "# HELP ibc_stuck_packets Number of stuck packets on an IBC channel\n"
	metrics += "# TYPE ibc_stuck_packets gauge\n"
	metrics += `ibc_stuck_packets{src_chain="cosmoshub-4",dst_chain="osmosis-1",src_channel="channel-141"} 0` + "\n"
	metrics += `ibc_stuck_packets{src_chain="osmosis-1",dst_chain="cosmoshub-4",src_channel="channel-0"} 0` + "\n"

	return metrics
}

func getStructuredMonitoringData() map[string]interface{} {
	// Try to fetch real data from Chainpulse
	chainpulseURL := os.Getenv("CHAINPULSE_METRICS_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://localhost:3001"
	}

	// Fetch metrics data
	metricsResp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	chains := []map[string]interface{}{}
	channels := []map[string]interface{}{}
	topRelayers := []map[string]interface{}{}
	recentActivity := []map[string]interface{}{}
	
	if err == nil && metricsResp.StatusCode == http.StatusOK {
		defer metricsResp.Body.Close()
		body, _ := io.ReadAll(metricsResp.Body)
		metrics := parsePrometheusMetrics(string(body))
		
		// Extract chain data
		if chainsData, ok := metrics["chains"].([]map[string]interface{}); ok {
			chains = chainsData
		}
		
		// Extract channel data
		if channelsData, ok := metrics["channels"].([]map[string]interface{}); ok {
			channels = channelsData
		}
		
		// Extract top relayers
		if relayersData, ok := metrics["relayers"].([]map[string]interface{}); ok {
			if len(relayersData) > 5 {
				topRelayers = relayersData[:5] // Top 5
			} else {
				topRelayers = relayersData
			}
		}
	}
	
	// If no real data, use mock data
	if len(chains) == 0 {
		chains = []map[string]interface{}{
			{
				"chain_id": "cosmoshub-4",
				"name":     "Cosmos Hub",
				"status":   "connected",
				"height":   18234567,
				"packets_24h": 4532,
			},
			{
				"chain_id": "osmosis-1",
				"name":     "Osmosis",
				"status":   "connected",
				"height":   12345678,
				"packets_24h": 6789,
			},
		}
	}
	
	if len(channels) == 0 {
		channels = []map[string]interface{}{
			{
				"src":             "cosmoshub-4",
				"dst":             "osmosis-1",
				"src_channel":     "channel-141",
				"dst_channel":     "channel-0",
				"status":          "active",
				"packets_pending": 0,
				"success_rate":    87.3,
			},
			{
				"src":             "osmosis-1",
				"dst":             "cosmoshub-4",
				"src_channel":     "channel-0",
				"dst_channel":     "channel-141",
				"status":          "active",
				"packets_pending": 0,
				"success_rate":    86.8,
			},
		}
	}
	
	if len(topRelayers) == 0 {
		topRelayers = []map[string]interface{}{
			{
				"address": "cosmos1relayer1",
				"packets_relayed": 1234,
				"success_rate": 89.5,
				"earnings_24h": "$1,234",
			},
			{
				"address": "osmo1relayer1",
				"packets_relayed": 1567,
				"success_rate": 92.3,
				"earnings_24h": "$1,567",
			},
		}
	}
	
	// Create some recent activity
	recentActivity = []map[string]interface{}{
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
	}

	return map[string]interface{}{
		"status": "healthy",
		"chains": chains,
		"channels": channels,
		"top_relayers": topRelayers,
		"recent_activity": recentActivity,
		"timestamp": time.Now(),
	}
}

// getMonitoringMetrics returns comprehensive monitoring metrics
func getMonitoringMetrics() map[string]interface{} {
	// Try to get real data from Chainpulse first
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://localhost:3000"
	}
	
	// Fetch and parse real metrics
	realData := map[string]interface{}{}
	metricsResp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	if err == nil && metricsResp.StatusCode == http.StatusOK {
		defer metricsResp.Body.Close()
		body, _ := io.ReadAll(metricsResp.Body)
		metricsBody := string(body)
		
		// Parse Prometheus metrics
		realData = parsePrometheusMetrics(metricsBody)
	}
	
	// Build response combining real and mock data
	chains := []map[string]interface{}{}
	channels := []map[string]interface{}{}
	relayers := []map[string]interface{}{}
	totalPackets := 0
	totalTxs := 0
	totalErrors := 0
	
	// Process chains from real data if available
	if realChains, ok := realData["chains"].([]map[string]interface{}); ok && len(realChains) > 0 {
		for _, chain := range realChains {
			chainID := chain["chain_id"].(string)
			chainName := chain["name"].(string)
			packets := chain["packets_24h"].(int)
			txs := chain["txs_total"].(int)
			errors := chain["errors"].(int)
			
			totalPackets += packets
			totalTxs += txs
			totalErrors += errors
			
			chains = append(chains, map[string]interface{}{
				"chainId": chainID,
				"chainName": chainName,
				"totalTxs": txs,
				"totalPackets": packets,
				"reconnects": 0, // Not available in metrics
				"timeouts": 0,   // Not available in metrics
				"errors": errors,
				"status": chain["status"].(string),
				"lastUpdate": time.Now(),
			})
		}
	} else {
		// Fallback to default chains
		chains = []map[string]interface{}{
			{
				"chainId": "cosmoshub-4",
				"chainName": "Cosmos Hub",
				"totalTxs": 12543,
				"totalPackets": 4532,
				"reconnects": 2,
				"timeouts": 0,
				"errors": 23,
				"status": "connected",
				"lastUpdate": time.Now(),
			},
			{
				"chainId": "osmosis-1",
				"chainName": "Osmosis",
				"totalTxs": 18976,
				"totalPackets": 6789,
				"reconnects": 1,
				"timeouts": 1,
				"errors": 24,
				"status": "connected",
				"lastUpdate": time.Now(),
			},
		}
		totalPackets = 11321
		totalTxs = 31519
		totalErrors = 47
	}
	
	// Process channels from real data if available
	if realChannels, ok := realData["channels"].([]map[string]interface{}); ok && len(realChannels) > 0 {
		channels = realChannels
	} else {
		// Fallback channels
		channels = []map[string]interface{}{
			{
				"srcChain": "osmosis-1",
				"dstChain": "cosmoshub-4",
				"srcChannel": "channel-0",
				"dstChannel": "channel-141",
				"srcPort": "transfer",
				"dstPort": "transfer",
				"totalPackets": 4500,
				"effectedPackets": 4250,
				"uneffectedPackets": 250,
				"successRate": 94.4,
				"status": "active",
			},
		}
	}
	
	// Process relayers from real data if available
	if realRelayers, ok := realData["relayers"].([]map[string]interface{}); ok && len(realRelayers) > 0 {
		relayers = realRelayers
	} else {
		// Fallback relayers
		relayers = []map[string]interface{}{
			{
				"address": "cosmos1xyz...abc",
				"totalPackets": 980,
				"effectedPackets": 856,
				"uneffectedPackets": 124,
				"frontrunCount": 12,
				"successRate": 87.3,
				"memo": "",
				"software": "hermes",
				"version": "1.8.0",
			},
			{
				"address": "osmo1abc...xyz",
				"totalPackets": 756,
				"effectedPackets": 697,
				"uneffectedPackets": 59,
				"frontrunCount": 8,
				"successRate": 92.2,
				"memo": "rly/2.4.2",
				"software": "go-relayer",
				"version": "2.4.2",
			},
		}
	}
	
	// Return comprehensive metrics matching MetricsSnapshot interface
	return map[string]interface{}{
		"system": map[string]interface{}{
			"totalChains": len(chains),
			"totalTransactions": totalTxs,
			"totalPackets": totalPackets,
			"totalErrors": totalErrors,
			"uptime": 86400, // 24 hours in seconds
			"lastSync": time.Now(),
		},
		"chains": chains,
		"relayers": relayers,
		"channels": channels,
		"recentPackets": []map[string]interface{}{},
		"stuckPackets": []map[string]interface{}{},
		"frontrunEvents": []map[string]interface{}{},
		"timestamp": time.Now(),
	}
}

// getChainRegistry returns the chain configuration registry
func getChainRegistry() map[string]interface{} {
	// In production, this should be loaded from a database or config file
	// For now, return a static configuration that can be extended
	
	// Define chain configurations
	chainConfigs := []map[string]interface{}{
		{
			"chain_id":       "cosmoshub-4",
			"chain_name":     "Cosmos Hub",
			"address_prefix": "cosmos",
			"explorer":       "https://www.mintscan.io/cosmos/txs",
			"logo":           "/images/chains/cosmos.svg",
		},
		{
			"chain_id":       "osmosis-1",
			"chain_name":     "Osmosis",
			"address_prefix": "osmo",
			"explorer":       "https://www.mintscan.io/osmosis/txs",
			"logo":           "/images/chains/osmosis.svg",
		},
		{
			"chain_id":       "neutron-1",
			"chain_name":     "Neutron",
			"address_prefix": "neutron",
			"explorer":       "https://www.mintscan.io/neutron/txs",
			"logo":           "/images/chains/neutron.svg",
		},
		{
			"chain_id":       "noble-1",
			"chain_name":     "Noble",
			"address_prefix": "noble",
			"explorer":       "https://www.mintscan.io/noble/txs",
			"logo":           "/images/chains/noble.svg",
		},
		{
			"chain_id":       "akash-1",
			"chain_name":     "Akash",
			"address_prefix": "akash",
			"explorer":       "https://www.mintscan.io/akash/txs",
			"logo":           "/images/chains/akash.svg",
		},
		{
			"chain_id":       "stargaze-1",
			"chain_name":     "Stargaze",
			"address_prefix": "stars",
			"explorer":       "https://www.mintscan.io/stargaze/txs",
			"logo":           "/images/chains/stargaze.svg",
		},
		{
			"chain_id":       "juno-1",
			"chain_name":     "Juno",
			"address_prefix": "juno",
			"explorer":       "https://www.mintscan.io/juno/txs",
			"logo":           "/images/chains/juno.svg",
		},
		{
			"chain_id":       "stride-1",
			"chain_name":     "Stride",
			"address_prefix": "stride",
			"explorer":       "https://www.mintscan.io/stride/txs",
			"logo":           "/images/chains/stride.svg",
		},
		{
			"chain_id":       "axelar-1",
			"chain_name":     "Axelar",
			"address_prefix": "axelar",
			"explorer":       "https://www.mintscan.io/axelar/txs",
			"logo":           "/images/chains/axelar.svg",
		},
	}
	
	// Convert to map for easy lookup
	chainsMap := make(map[string]interface{})
	for _, chain := range chainConfigs {
		chainsMap[chain["chain_id"].(string)] = chain
	}
	
	// Define channel configurations
	channels := []map[string]interface{}{
		// Cosmos Hub channels
		{"source_chain": "cosmoshub-4", "source_channel": "channel-141", "dest_chain": "osmosis-1", "dest_channel": "channel-0", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "cosmoshub-4", "source_channel": "channel-536", "dest_chain": "noble-1", "dest_channel": "channel-4", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "cosmoshub-4", "source_channel": "channel-569", "dest_chain": "neutron-1", "dest_channel": "channel-1", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		// Osmosis channels
		{"source_chain": "osmosis-1", "source_channel": "channel-0", "dest_chain": "cosmoshub-4", "dest_channel": "channel-141", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "osmosis-1", "source_channel": "channel-750", "dest_chain": "noble-1", "dest_channel": "channel-1", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "osmosis-1", "source_channel": "channel-874", "dest_chain": "neutron-1", "dest_channel": "channel-10", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		// Noble channels
		{"source_chain": "noble-1", "source_channel": "channel-1", "dest_chain": "osmosis-1", "dest_channel": "channel-750", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "noble-1", "source_channel": "channel-4", "dest_chain": "cosmoshub-4", "dest_channel": "channel-536", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "noble-1", "source_channel": "channel-18", "dest_chain": "neutron-1", "dest_channel": "channel-30", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		// Neutron channels
		{"source_chain": "neutron-1", "source_channel": "channel-1", "dest_chain": "cosmoshub-4", "dest_channel": "channel-569", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "neutron-1", "source_channel": "channel-10", "dest_chain": "osmosis-1", "dest_channel": "channel-874", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
		{"source_chain": "neutron-1", "source_channel": "channel-30", "dest_chain": "noble-1", "dest_channel": "channel-18", "source_port": "transfer", "dest_port": "transfer", "status": "active"},
	}
	
	return map[string]interface{}{
		"chains":   chainsMap,
		"channels": channels,
		"api_version": "1.0",
	}
}