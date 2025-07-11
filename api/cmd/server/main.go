package main

import (
	"encoding/json"
	"fmt"
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

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost", "http://localhost:80", "http://localhost:3000"},
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