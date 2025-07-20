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
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/relayooor/api/internal/config"
	"github.com/relayooor/api/pkg/chainpulse"
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

// ChannelInfo stores counterparty chain information for a channel
type ChannelInfo struct {
	CounterpartyChainID string
	CounterpartyChannel string
	QueryTime           time.Time
}

// Global channel cache with mutex for thread safety
var (
	channelCache      = make(map[string]ChannelInfo) // key: "chainID/channelID"
	channelCacheMutex sync.RWMutex
	chainRegistry     = config.DefaultChainRegistry()
)

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize Chainpulse client
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://localhost:3001"
	}
	chainpulseClient := chainpulse.NewClient(chainpulseURL)
	
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
			chainpulseURL = "http://localhost:3001" // Correct port for Chainpulse
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
							// Extract packet data
							sourceChain := fmt.Sprintf("%v", packet["chain_id"])
							srcChannel := fmt.Sprintf("%v", packet["src_channel"])
							receiver := fmt.Sprintf("%v", packet["receiver"])
							
							// Get destination chain using cached lookup
							destChain := getDestinationChain(sourceChain, srcChannel, receiver)
							
							sp := StuckPacket{
								ID:               fmt.Sprintf("%v-%v", packet["chain_id"], packet["sequence"]),
								ChannelID:        srcChannel,
								Sequence:         int(getFloat64(packet["sequence"])),
								SourceChain:      sourceChain,
								DestinationChain: destChain,
								StuckDuration:    formatDuration(int(getFloat64(packet["age_seconds"]))),
								Amount:           fmt.Sprintf("%v", packet["amount"]),
								Denom:            fmt.Sprintf("%v", packet["denom"]),
								Sender:           fmt.Sprintf("%v", packet["sender"]),
								Receiver:         receiver,
								TxHash:           "", // Not provided by Chainpulse
							}
							
							// Calculate timestamp from age
							if ageSeconds := getFloat64(packet["age_seconds"]); ageSeconds > 0 {
								sp.Timestamp = time.Now().Add(-time.Duration(ageSeconds) * time.Second)
							}
							
							// Add relay attempts if available
							if attempts, ok := packet["relay_attempts"]; ok {
								sp.RelayAttempts = int(getFloat64(attempts))
							}
							
							// Add IBC version if available
							if version, ok := packet["ibc_version"]; ok {
								sp.IBCVersion = fmt.Sprintf("%v", version)
							}
							
							// Add last attempt by if available
							if lastAttempt, ok := packet["last_attempt_by"]; ok {
								sp.LastAttemptBy = fmt.Sprintf("%v", lastAttempt)
							}
							
							stuckPackets = append(stuckPackets, sp)
						}
					}
					
					log.Printf("Successfully fetched %d stuck packets from Chainpulse", len(stuckPackets))
					json.NewEncoder(w).Encode(stuckPackets)
					return
				}
			}
		} else if err != nil {
			log.Printf("Failed to fetch stuck packets from Chainpulse: %v", err)
		} else {
			log.Printf("Chainpulse returned status %d for stuck packets", resp.StatusCode)
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
			chainpulseURL = "http://localhost:3001" // Correct port for Chainpulse
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
		// Use Chainpulse API to get real data
		stuckPacketsResp, err := chainpulseClient.GetStuckPackets(900, 100)
		if err != nil {
			log.Printf("Error fetching stuck packets: %v", err)
		}
		
		congestionResp, err := chainpulseClient.GetChannelCongestion()
		if err != nil {
			log.Printf("Error fetching channel congestion: %v", err)
		}
		
		// Get chain registry for metadata
		registry := config.DefaultChainRegistry()
		
		// Build chains data
		chains := []map[string]interface{}{}
		for chainID, chainConfig := range registry.Chains {
			chains = append(chains, map[string]interface{}{
				"chain_id": chainID,
				"name":     chainConfig.ChainName,
				"status":   "connected",
				"height":   0, // TODO: Get from metrics
				"packets_24h": 0, // TODO: Get from metrics
			})
		}
		
		// Build channel data from congestion response
		channels := []map[string]interface{}{}
		if congestionResp != nil {
			for _, ch := range congestionResp.Channels {
				// Try to determine source chain from stuck packets
				srcChain := "unknown"
				if stuckPacketsResp != nil {
					for _, packet := range stuckPacketsResp.Packets {
						if packet.SrcChannel == ch.SrcChannel {
							srcChain = packet.ChainID
							break
						}
					}
				}
				
				channels = append(channels, map[string]interface{}{
					"src":             srcChain,
					"dst":             inferDestinationChain(srcChain, ch.SrcChannel),
					"src_channel":     ch.SrcChannel,
					"dst_channel":     ch.DstChannel,
					"status":          "active",
					"packets_pending": ch.StuckCount,
					"success_rate":    0.0, // TODO: Calculate from metrics
				})
			}
		}
		
		// Build top relayers from stuck packets
		relayerMap := make(map[string]map[string]interface{})
		if stuckPacketsResp != nil {
			for _, packet := range stuckPacketsResp.Packets {
				if packet.LastAttemptBy != "" {
					if _, exists := relayerMap[packet.LastAttemptBy]; !exists {
						relayerMap[packet.LastAttemptBy] = map[string]interface{}{
							"address": packet.LastAttemptBy,
							"packets_relayed": 0,
							"success_rate": 0.0,
							"earnings_24h": "$0",
						}
					}
					relayerMap[packet.LastAttemptBy]["packets_relayed"] = relayerMap[packet.LastAttemptBy]["packets_relayed"].(int) + packet.RelayAttempts
				}
			}
		}
		
		topRelayers := []map[string]interface{}{}
		for _, relayer := range relayerMap {
			topRelayers = append(topRelayers, relayer)
		}
		
		// Recent activity from stuck packets
		recentActivity := []map[string]interface{}{}
		if stuckPacketsResp != nil && len(stuckPacketsResp.Packets) > 0 {
			limit := 10
			if len(stuckPacketsResp.Packets) < limit {
				limit = len(stuckPacketsResp.Packets)
			}
			
			for i := 0; i < limit; i++ {
				packet := stuckPacketsResp.Packets[i]
				recentActivity = append(recentActivity, map[string]interface{}{
					"from_chain": packet.ChainID,
					"to_chain":   inferDestinationChain(packet.ChainID, packet.SrcChannel),
					"channel":    packet.SrcChannel,
					"status":     "pending",
					"timestamp":  time.Now().Add(-time.Duration(packet.AgeSeconds) * time.Second),
				})
			}
		}
		
		// Calculate system totals
		totalStuckPackets := 0
		if stuckPacketsResp != nil {
			totalStuckPackets = stuckPacketsResp.Total
		}
		
		data := map[string]interface{}{
			"timestamp": time.Now(),
			"chains": chains,
			"top_relayers": topRelayers,
			"recent_activity": recentActivity,
			"channels": channels,
			"system": map[string]interface{}{
				"totalChains": len(chains),
				"totalPackets": totalStuckPackets,
				"totalErrors": 0, // TODO: Get from metrics
				"uptime": 99.8, // TODO: Calculate from actual uptime
				"lastSync": time.Now(),
			},
		}
		
		json.NewEncoder(w).Encode(data)
	}).Methods("GET")

	// Channel congestion endpoint
	api.HandleFunc("/channels/congestion", func(w http.ResponseWriter, r *http.Request) {
		// Use Chainpulse client to get congestion data
		congestionResp, err := chainpulseClient.GetChannelCongestion()
		if err != nil {
			log.Printf("Error fetching channel congestion: %v", err)
			// Return empty channels on error
			json.NewEncoder(w).Encode(map[string]interface{}{
				"channels": []map[string]interface{}{},
				"api_version": "1.0",
			})
			return
		}
		
		// Transform response to match expected format
		channels := []map[string]interface{}{}
		for _, ch := range congestionResp.Channels {
			channels = append(channels, map[string]interface{}{
				"src_channel": ch.SrcChannel,
				"dst_channel": ch.DstChannel,
				"stuck_count": ch.StuckCount,
				"oldest_stuck_age_seconds": ch.OldestStuckAgeSeconds,
				"total_value": ch.TotalValue,
			})
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"channels": channels,
			"api_version": congestionResp.APIVersion,
		})
	}).Methods("GET")

	// Packet details endpoint
	api.HandleFunc("/packets/{chain}/{channel}/{sequence}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		chain := vars["chain"]
		channel := vars["channel"]
		seqStr := vars["sequence"]
		
		// Parse sequence to int64
		sequence, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid sequence number"})
			return
		}

		// Use Chainpulse client to get packet details
		packet, err := chainpulseClient.GetPacketDetails(chain, channel, sequence)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Packet not found"})
			} else {
				log.Printf("Error fetching packet details: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch packet details"})
			}
			return
		}
		
		// Return packet data
		json.NewEncoder(w).Encode(packet)
	}).Methods("GET")

	// Expired packets endpoint - returns packets that have already timed out
	api.HandleFunc("/packets/expired", func(w http.ResponseWriter, r *http.Request) {
		// Use Chainpulse client to get expired packets
		expiredResp, err := chainpulseClient.GetExpiredPackets()
		if err != nil {
			log.Printf("Error fetching expired packets: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to fetch expired packets",
				"packets": []interface{}{},
			})
			return
		}
		
		// Transform response to match expected format
		packets := []map[string]interface{}{}
		for _, p := range expiredResp.Packets {
			packets = append(packets, map[string]interface{}{
				"id":                      fmt.Sprintf("%s-%s-%d", p.ChainID, p.SrcChannel, p.Sequence),
				"chain_id":                p.ChainID,
				"sequence":                p.Sequence,
				"src_channel":             p.SrcChannel,
				"dst_channel":             p.DstChannel,
				"sender":                  p.Sender,
				"receiver":                p.Receiver,
				"amount":                  p.Amount,
				"denom":                   p.Denom,
				"seconds_since_timeout":   p.SecondsSinceTimeout,
				"timeout_type":            p.TimeoutType,
				"age_seconds":             p.AgeSeconds,
			})
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"packets": packets,
			"api_version": expiredResp.APIVersion,
		})
	}).Methods("GET")

	// Expiring packets endpoint - returns packets that will timeout soon
	api.HandleFunc("/packets/expiring", func(w http.ResponseWriter, r *http.Request) {
		// Get minutes parameter from query
		minutes := 60 // Default to 60 minutes
		if m := r.URL.Query().Get("minutes"); m != "" {
			if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
				minutes = parsed
			}
		}
		
		// Use Chainpulse client to get expiring packets
		expiringResp, err := chainpulseClient.GetExpiringPackets(minutes)
		if err != nil {
			log.Printf("Error fetching expiring packets: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to fetch expiring packets",
				"packets": []interface{}{},
			})
			return
		}
		
		// Transform response to match expected format
		packets := []map[string]interface{}{}
		for _, p := range expiringResp.Packets {
			packets = append(packets, map[string]interface{}{
				"id":                     fmt.Sprintf("%s-%s-%d", p.ChainID, p.SrcChannel, p.Sequence),
				"chain_id":               p.ChainID,
				"sequence":               p.Sequence,
				"src_channel":            p.SrcChannel,
				"dst_channel":            p.DstChannel,
				"sender":                 p.Sender,
				"receiver":               p.Receiver,
				"amount":                 p.Amount,
				"denom":                  p.Denom,
				"seconds_until_timeout":  p.SecondsUntilTimeout,
				"timeout_type":           p.TimeoutType,
				"timeout_value":          p.TimeoutValue,
				"age_seconds":            p.AgeSeconds,
			})
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"packets": packets,
			"api_version": expiringResp.APIVersion,
		})
	}).Methods("GET")

	// Duplicate packets endpoint - returns packets with duplicate data
	api.HandleFunc("/packets/duplicates", func(w http.ResponseWriter, r *http.Request) {
		// Use Chainpulse client to get duplicate packets
		duplicatesResp, err := chainpulseClient.GetDuplicatePackets()
		if err != nil {
			log.Printf("Error fetching duplicate packets: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to fetch duplicate packets",
				"duplicates": []interface{}{},
			})
			return
		}
		
		// Transform response to match expected format
		duplicates := []map[string]interface{}{}
		for _, d := range duplicatesResp.Duplicates {
			packets := []map[string]interface{}{}
			for _, p := range d.Packets {
				packets = append(packets, map[string]interface{}{
					"chain_id":    p.ChainID,
					"sequence":    p.Sequence,
					"src_channel": p.SrcChannel,
					"sender":      p.Sender,
					"created_at":  p.CreatedAt,
				})
			}
			
			duplicates = append(duplicates, map[string]interface{}{
				"data_hash": d.DataHash,
				"count":     d.Count,
				"packets":   packets,
			})
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"duplicates": duplicates,
			"api_version": duplicatesResp.APIVersion,
		})
	}).Methods("GET")

	// User transfer history endpoint using Chainpulse
	api.HandleFunc("/user/{wallet}/history", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		wallet := vars["wallet"]
		
		// Validate wallet address format
		if !isValidWalletAddress(wallet) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid wallet address"})
			return
		}
		
		// Get query parameters
		role := r.URL.Query().Get("role") // sender, receiver, or both
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		
		if limit <= 0 {
			limit = 100
		}
		
		// Use Chainpulse client to get user packets
		packetsResp, err := chainpulseClient.GetPacketsByUser(wallet, role, limit, offset)
		if err != nil {
			log.Printf("Error fetching user packets: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to fetch user transfer history",
				"transfers": []interface{}{},
			})
			return
		}
		
		// Transform to transfer format
		transfers := []map[string]interface{}{}
		for _, p := range packetsResp.Packets {
			transfer := map[string]interface{}{
				"id":                fmt.Sprintf("%s-%s-%d", p.ChainID, p.SrcChannel, p.Sequence),
				"chain_id":          p.ChainID,
				"channel_id":        p.SrcChannel,
				"sequence":          p.Sequence,
				"source_chain":      p.ChainID,
				"destination_chain": inferDestinationChain(p.ChainID, p.SrcChannel),
				"sender":            p.Sender,
				"receiver":          p.Receiver,
				"amount":            p.Amount,
				"denom":             p.Denom,
				"status":            "pending", // All stuck packets are pending
				"age_seconds":       p.AgeSeconds,
				"relay_attempts":    p.RelayAttempts,
			}
			
			// Calculate timestamp from age
			if p.AgeSeconds > 0 {
				transfer["timestamp"] = time.Now().Add(-time.Duration(p.AgeSeconds) * time.Second)
			}
			
			transfers = append(transfers, transfer)
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"transfers": transfers,
			"total": packetsResp.Total,
			"api_version": packetsResp.APIVersion,
		})
	}).Methods("GET")

	// Packet search endpoint - comprehensive search across all packet data
	api.HandleFunc("/packets/search", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		sender := r.URL.Query().Get("sender")
		receiver := r.URL.Query().Get("receiver")
		chainID := r.URL.Query().Get("chain_id")
		denom := r.URL.Query().Get("denom")
		minAgeStr := r.URL.Query().Get("min_age_seconds")
		limitStr := r.URL.Query().Get("limit")
		
		minAge := 0
		if minAgeStr != "" {
			fmt.Sscanf(minAgeStr, "%d", &minAge)
		}
		
		limit := 100
		if limitStr != "" {
			fmt.Sscanf(limitStr, "%d", &limit)
		}
		
		// Get stuck packets first
		stuckResp, err := chainpulseClient.GetStuckPackets(minAge, limit)
		if err != nil {
			log.Printf("Error fetching stuck packets: %v", err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"packets": []interface{}{},
				"error": "Failed to fetch packets",
			})
			return
		}
		
		// Filter results based on search criteria
		filteredPackets := []chainpulse.Packet{}
		for _, packet := range stuckResp.Packets {
			// Apply filters
			if sender != "" && packet.Sender != sender {
				continue
			}
			if receiver != "" && packet.Receiver != receiver {
				continue
			}
			if chainID != "" && packet.ChainID != chainID {
				continue
			}
			if denom != "" && !strings.Contains(packet.Denom, denom) {
				continue
			}
			
			filteredPackets = append(filteredPackets, packet)
		}
		
		// Also search for user-specific packets if wallet address provided
		if (sender != "" || receiver != "") && chainID == "" {
			// If we have a wallet address but no chain filter, also query user packets
			address := sender
			if address == "" {
				address = receiver
			}
			
			userResp, err := chainpulseClient.GetPacketsByUser(address, "", limit, 0)
			if err == nil {
				// Merge results, avoiding duplicates
				existingPackets := make(map[string]bool)
				for _, p := range filteredPackets {
					key := fmt.Sprintf("%s-%d", p.ChainID, p.Sequence)
					existingPackets[key] = true
				}
				
				for _, packet := range userResp.Packets {
					key := fmt.Sprintf("%s-%d", packet.ChainID, packet.Sequence)
					if !existingPackets[key] {
						// Apply same filters
						if sender != "" && packet.Sender != sender {
							continue
						}
						if receiver != "" && packet.Receiver != receiver {
							continue
						}
						if denom != "" && !strings.Contains(packet.Denom, denom) {
							continue
						}
						filteredPackets = append(filteredPackets, packet)
					}
				}
			}
		}
		
		// Return results
		json.NewEncoder(w).Encode(map[string]interface{}{
			"packets": filteredPackets,
			"total": len(filteredPackets),
			"filters_applied": map[string]interface{}{
				"sender": sender,
				"receiver": receiver,
				"chain_id": chainID,
				"denom": denom,
				"min_age_seconds": minAge,
			},
		})
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
		// Use Chainpulse API to get real data
		stuckPacketsResp, err := chainpulseClient.GetStuckPackets(0, 1000) // Get more for metrics
		if err != nil {
			log.Printf("Error fetching stuck packets: %v", err)
		}
		
		congestionResp, err := chainpulseClient.GetChannelCongestion()
		if err != nil {
			log.Printf("Error fetching channel congestion: %v", err)
		}
		
		registry := config.DefaultChainRegistry()
		
		// Build chains data
		chains := []map[string]interface{}{}
		for chainID, chainConfig := range registry.Chains {
			chains = append(chains, map[string]interface{}{
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
		relayerMap := make(map[string]map[string]interface{})
		if stuckPacketsResp != nil {
			for _, packet := range stuckPacketsResp.Packets {
				if packet.LastAttemptBy != "" {
					if _, exists := relayerMap[packet.LastAttemptBy]; !exists {
						relayerMap[packet.LastAttemptBy] = map[string]interface{}{
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
		
		relayers := []map[string]interface{}{}
		for _, stats := range relayerMap {
			relayers = append(relayers, stats)
		}
		
		// Build channel data
		channels := []map[string]interface{}{}
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
							dstChain = inferDestinationChain(packet.ChainID, packet.SrcChannel)
							break
						}
					}
				}
				
				channels = append(channels, map[string]interface{}{
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
		recentPackets := []map[string]interface{}{}
		if stuckPacketsResp != nil && len(stuckPacketsResp.Packets) > 0 {
			limit := 5
			if len(stuckPacketsResp.Packets) < limit {
				limit = len(stuckPacketsResp.Packets)
			}
			for i := 0; i < limit; i++ {
				packet := stuckPacketsResp.Packets[i]
				recentPackets = append(recentPackets, map[string]interface{}{
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
		
		metrics := map[string]interface{}{
			"system": map[string]interface{}{
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
			"frontrunEvents": []map[string]interface{}{}, // TODO: Parse from metrics
			"timestamp": time.Now(),
		}
		
		json.NewEncoder(w).Encode(metrics)
	}).Methods("GET")

	// Chain endpoints - provides REST API URLs for chains
	api.HandleFunc("/chains/endpoints", func(w http.ResponseWriter, r *http.Request) {
		endpoints := []map[string]string{
			{"chain_id": "cosmoshub-4", "rest_url": "https://cosmos-rest.publicnode.com"},
			{"chain_id": "osmosis-1", "rest_url": "https://osmosis-rest.publicnode.com"},
			{"chain_id": "neutron-1", "rest_url": "https://neutron-rest.publicnode.com"},
			{"chain_id": "noble-1", "rest_url": "https://noble-rest.publicnode.com"},
			{"chain_id": "axelar-dojo-1", "rest_url": "https://axelar-rest.publicnode.com"},
			{"chain_id": "stride-1", "rest_url": "https://stride-rest.publicnode.com"},
			{"chain_id": "dydx-mainnet-1", "rest_url": "https://dydx-rest.publicnode.com"},
			{"chain_id": "celestia", "rest_url": "https://celestia-rest.publicnode.com"},
			{"chain_id": "injective-1", "rest_url": "https://injective-rest.publicnode.com"},
			{"chain_id": "kava_2222-10", "rest_url": "https://kava-rest.publicnode.com"},
			{"chain_id": "secret-4", "rest_url": "https://secret-rest.publicnode.com"},
			{"chain_id": "stargaze-1", "rest_url": "https://stargaze-rest.publicnode.com"},
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"endpoints": endpoints,
		})
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
	// Look for matching channel in the registry
	for _, channel := range chainRegistry.Channels {
		if channel.SourceChain == srcChain && channel.SourceChannel == srcChannel {
			return channel.DestChain
		}
	}
	
	// Default fallback
	return "unknown"
}

// inferChainFromAddress attempts to determine the chain ID from a bech32 address prefix
func inferChainFromAddress(address string) string {
	// Extract prefix from address (everything before "1")
	idx := strings.Index(address, "1")
	if idx <= 0 {
		return ""
	}
	prefix := address[:idx]
	
	// Look up chain by prefix in the registry
	if chain, ok := chainRegistry.GetChainByPrefix(prefix); ok {
		return chain.ChainID
	}
	
	return ""
}

// queryChannelInfo queries the chain REST API to get counterparty chain information
func queryChannelInfo(chainID, channelID string) (*ChannelInfo, error) {
	chain, ok := chainRegistry.GetChainByID(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %s not found in registry", chainID)
	}
	
	if chain.RESTEndpoint == "" {
		return nil, fmt.Errorf("no REST endpoint configured for chain %s", chainID)
	}
	
	// Query the IBC channel client state to get counterparty chain ID
	url := fmt.Sprintf("%s/ibc/core/channel/v1/channels/%s/ports/transfer/client_state", chain.RESTEndpoint, channelID)
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query REST API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST API returned status %d", resp.StatusCode)
	}
	
	var result struct {
		IdentifiedClientState struct {
			ClientState struct {
				ChainID string `json:"chain_id"`
			} `json:"client_state"`
		} `json:"identified_client_state"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	counterpartyChainID := result.IdentifiedClientState.ClientState.ChainID
	if counterpartyChainID == "" {
		return nil, fmt.Errorf("no counterparty chain ID in response")
	}
	
	return &ChannelInfo{
		CounterpartyChainID: counterpartyChainID,
		QueryTime:           time.Now(),
	}, nil
}

// getDestinationChain gets the destination chain for a channel, using cache when possible
func getDestinationChain(sourceChain, srcChannel, receiverAddr string) string {
	cacheKey := fmt.Sprintf("%s/%s", sourceChain, srcChannel)
	
	// Check cache first
	channelCacheMutex.RLock()
	if info, ok := channelCache[cacheKey]; ok {
		channelCacheMutex.RUnlock()
		// Cache entries are valid for 1 hour
		if time.Since(info.QueryTime) < time.Hour {
			return info.CounterpartyChainID
		}
	} else {
		channelCacheMutex.RUnlock()
	}
	
	// Try static mappings first (faster)
	if destChain := inferDestinationChain(sourceChain, srcChannel); destChain != "unknown" {
		// Cache the result
		channelCacheMutex.Lock()
		channelCache[cacheKey] = ChannelInfo{
			CounterpartyChainID: destChain,
			QueryTime:           time.Now(),
		}
		channelCacheMutex.Unlock()
		return destChain
	}
	
	// Try to query the REST API
	if info, err := queryChannelInfo(sourceChain, srcChannel); err == nil {
		// Cache the result
		channelCacheMutex.Lock()
		channelCache[cacheKey] = *info
		channelCacheMutex.Unlock()
		log.Printf("Cached channel info: %s/%s -> %s", sourceChain, srcChannel, info.CounterpartyChainID)
		return info.CounterpartyChainID
	} else {
		log.Printf("Failed to query channel info for %s/%s: %v", sourceChain, srcChannel, err)
	}
	
	// Fallback: try to infer from receiver address
	if receiverAddr != "" && receiverAddr != "<nil>" {
		if destChain := inferChainFromAddress(receiverAddr); destChain != "" {
			return destChain
		}
	}
	
	return "unknown"
}

// getChainMap returns a map of chain IDs to chain names
func getChainMap() map[string]string {
	chainMap := make(map[string]string)
	for chainID, chain := range chainRegistry.Chains {
		chainMap[chainID] = chain.ChainName
	}
	return chainMap
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
	// Convert the centralized registry to the API response format
	chainsMap := make(map[string]interface{})
	for chainID, chain := range chainRegistry.Chains {
		chainsMap[chainID] = map[string]interface{}{
			"chain_id":       chain.ChainID,
			"chain_name":     chain.ChainName,
			"address_prefix": chain.AddressPrefix,
			"rpc_endpoint":   chain.RPCEndpoint,
			"rest_endpoint":  chain.RESTEndpoint,
			"ws_endpoint":    chain.WSEndpoint,
			"grpc_endpoint":  chain.GRPCEndpoint,
			"explorer":       chain.Explorer,
			"logo":           chain.Logo,
		}
	}
	
	// Convert channels to interface slice
	channels := make([]map[string]interface{}, len(chainRegistry.Channels))
	for i, ch := range chainRegistry.Channels {
		channels[i] = map[string]interface{}{
			"source_chain":   ch.SourceChain,
			"source_channel": ch.SourceChannel,
			"dest_chain":     ch.DestChain,
			"dest_channel":   ch.DestChannel,
			"source_port":    ch.SourcePort,
			"dest_port":      ch.DestPort,
			"status":         ch.Status,
		}
	}
	
	return map[string]interface{}{
		"chains":      chainsMap,
		"channels":    channels,
		"api_version": "1.0",
	}
}