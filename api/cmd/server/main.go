package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func main() {
	r := mux.NewRouter()

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

		// Mock data for demo - in production this would query blockchain
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
		// Proxy to the actual Chainpulse endpoint
		chainpulseURL := "http://localhost:3000/metrics"
		
		resp, err := http.Get(chainpulseURL)
		if err != nil {
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

	// Structured monitoring data endpoint
	api.HandleFunc("/monitoring/data", func(w http.ResponseWriter, r *http.Request) {
		data := getStructuredMonitoringData()
		json.NewEncoder(w).Encode(data)
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
	return map[string]interface{}{
		"status": "healthy",
		"chains": []map[string]interface{}{
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
		},
		"channels": []map[string]interface{}{
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
		},
		"timestamp": time.Now(),
	}
}