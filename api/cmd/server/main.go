package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	ID               string `json:"id"`
	ChannelID        string `json:"channelId"`
	Sequence         int    `json:"sequence"`
	SourceChain      string `json:"sourceChain"`
	DestinationChain string `json:"destinationChain"`
	StuckDuration    string `json:"stuckDuration"`
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

	// Stuck packets endpoint
	api.HandleFunc("/packets/stuck", func(w http.ResponseWriter, r *http.Request) {
		packets := []StuckPacket{}
		json.NewEncoder(w).Encode(packets)
	}).Methods("GET")

	// Clear packets endpoint
	api.HandleFunc("/packets/clear", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
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