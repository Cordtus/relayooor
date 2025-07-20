package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
)

// RegisterEnrichedEndpoints adds enriched packet endpoints to the router
func RegisterEnrichedEndpoints(router *mux.Router, enrichmentService *PacketEnrichmentService) {
	api := router.PathPrefix("/api").Subrouter()
	
	// Enriched packet search endpoint - combines all data sources
	api.HandleFunc("/packets/enriched/search", func(w http.ResponseWriter, r *http.Request) {
		// Get search parameters
		sender := r.URL.Query().Get("sender")
		receiver := r.URL.Query().Get("receiver")
		chainID := r.URL.Query().Get("chain_id")
		channelID := r.URL.Query().Get("channel_id")
		denom := r.URL.Query().Get("denom")
		status := r.URL.Query().Get("status")
		limitStr := r.URL.Query().Get("limit")
		
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}
		
		// First get stuck packets from existing endpoint
		stuckPacketsResp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/packets/stuck?limit=%d", limit))
		if err != nil {
			http.Error(w, "Failed to fetch stuck packets", http.StatusInternalServerError)
			return
		}
		defer stuckPacketsResp.Body.Close()
		
		var stuckPackets []map[string]interface{}
		if err := json.NewDecoder(stuckPacketsResp.Body).Decode(&stuckPackets); err != nil {
			http.Error(w, "Failed to parse stuck packets", http.StatusInternalServerError)
			return
		}
		
		// Enrich each packet
		enrichedPackets := []interface{}{}
		for _, packet := range stuckPackets {
			// Apply filters
			if sender != "" {
				if s, ok := packet["sender"].(string); !ok || s != sender {
					continue
				}
			}
			if receiver != "" {
				if r, ok := packet["receiver"].(string); !ok || r != receiver {
					continue
				}
			}
			if chainID != "" {
				if c, ok := packet["sourceChain"].(string); !ok || c != chainID {
					continue
				}
			}
			if channelID != "" {
				if ch, ok := packet["channelId"].(string); !ok || ch != channelID {
					continue
				}
			}
			if denom != "" {
				if d, ok := packet["denom"].(string); !ok || d != denom {
					continue
				}
			}
			
			// Enrich the packet
			enriched, err := enrichmentService.EnrichPacket(packet)
			if err != nil {
				// Include packet with error info
				packet["enrichment_error"] = err.Error()
				enrichedPackets = append(enrichedPackets, packet)
			} else {
				enrichedPackets = append(enrichedPackets, enriched)
			}
			
			if len(enrichedPackets) >= limit {
				break
			}
		}
		
		// Return enriched results
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"packets": enrichedPackets,
				"total":   len(enrichedPackets),
				"filters": map[string]interface{}{
					"sender":     sender,
					"receiver":   receiver,
					"chain_id":   chainID,
					"channel_id": channelID,
					"denom":      denom,
					"status":     status,
				},
				"enrichment_sources": []string{
					"chainpulse",
					"hermes",
					"chain_registry",
					"token_registry",
				},
			},
		})
	}).Methods("GET")
	
	// Get single enriched packet by ID
	api.HandleFunc("/packets/enriched/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		packetID := vars["id"]
		
		// Try to find the packet
		stuckPacketsResp, err := http.Get("http://localhost:8080/api/packets/stuck?limit=1000")
		if err != nil {
			http.Error(w, "Failed to fetch packets", http.StatusInternalServerError)
			return
		}
		defer stuckPacketsResp.Body.Close()
		
		var stuckPackets []map[string]interface{}
		if err := json.NewDecoder(stuckPacketsResp.Body).Decode(&stuckPackets); err != nil {
			http.Error(w, "Failed to parse packets", http.StatusInternalServerError)
			return
		}
		
		// Find the specific packet
		var targetPacket map[string]interface{}
		for _, packet := range stuckPackets {
			if id, ok := packet["id"].(string); ok && id == packetID {
				targetPacket = packet
				break
			}
		}
		
		if targetPacket == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Packet not found",
			})
			return
		}
		
		// Enrich the packet
		enriched, err := enrichmentService.EnrichPacket(targetPacket)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    enriched,
		})
	}).Methods("GET")
	
	// Packet clearing preparation endpoint
	api.HandleFunc("/packets/enriched/{id}/prepare-clear", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		packetID := vars["id"]
		
		// Get the enriched packet first
		enrichedResp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/packets/enriched/%s", packetID))
		if err != nil {
			http.Error(w, "Failed to fetch packet", http.StatusInternalServerError)
			return
		}
		defer enrichedResp.Body.Close()
		
		var enrichedResult map[string]interface{}
		if err := json.NewDecoder(enrichedResp.Body).Decode(&enrichedResult); err != nil {
			http.Error(w, "Failed to parse packet", http.StatusInternalServerError)
			return
		}
		
		if !enrichedResult["success"].(bool) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(enrichedResult)
			return
		}
		
		enrichedPacket := enrichedResult["data"].(*EnrichedPacket)
		
		// Prepare clearing information
		clearingPrep := map[string]interface{}{
			"packet_id":     packetID,
			"can_clear":     enrichedPacket.ClearingInfo.CanClear,
			"chain_id":      enrichedPacket.DestinationChain,
			"channel_id":    enrichedPacket.ChannelInfo.DestinationChannel,
			"port_id":       enrichedPacket.PortID,
			"sequence":      enrichedPacket.Sequence,
			"estimated_gas": enrichedPacket.ClearingInfo.EstimatedGas,
			"estimated_fee": enrichedPacket.ClearingInfo.EstimatedFee,
			"clearing_steps": enrichedPacket.ClearingInfo.ClearingSteps,
			"hermes_command": fmt.Sprintf(
				"hermes tx packet-recv --dst-chain %s --src-chain %s --src-port %s --src-channel %s",
				enrichedPacket.DestinationChain,
				enrichedPacket.SourceChain,
				enrichedPacket.PortID,
				enrichedPacket.ChannelID,
			),
			"required_wallet": enrichedPacket.ClearingInfo.RequiredWallet,
			"warnings": []string{},
		}
		
		// Add warnings if applicable
		if enrichedPacket.DestinationChain != "cosmoshub-4" {
			clearingPrep["warnings"] = append(
				clearingPrep["warnings"].([]string),
				"Destination chain is not configured in Hermes",
			)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    clearingPrep,
		})
	}).Methods("GET")
	
	// Batch enrichment endpoint
	api.HandleFunc("/packets/enriched/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		var request struct {
			PacketIDs []string `json:"packet_ids"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		// Fetch all stuck packets
		stuckPacketsResp, err := http.Get("http://localhost:8080/api/packets/stuck?limit=1000")
		if err != nil {
			http.Error(w, "Failed to fetch packets", http.StatusInternalServerError)
			return
		}
		defer stuckPacketsResp.Body.Close()
		
		var stuckPackets []map[string]interface{}
		if err := json.NewDecoder(stuckPacketsResp.Body).Decode(&stuckPackets); err != nil {
			http.Error(w, "Failed to parse packets", http.StatusInternalServerError)
			return
		}
		
		// Create a map for quick lookup
		packetMap := make(map[string]map[string]interface{})
		for _, packet := range stuckPackets {
			if id, ok := packet["id"].(string); ok {
				packetMap[id] = packet
			}
		}
		
		// Enrich requested packets
		enrichedPackets := []interface{}{}
		errors := []map[string]string{}
		
		for _, packetID := range request.PacketIDs {
			packet, exists := packetMap[packetID]
			if !exists {
				errors = append(errors, map[string]string{
					"packet_id": packetID,
					"error":     "Packet not found",
				})
				continue
			}
			
			enriched, err := enrichmentService.EnrichPacket(packet)
			if err != nil {
				errors = append(errors, map[string]string{
					"packet_id": packetID,
					"error":     err.Error(),
				})
			} else {
				enrichedPackets = append(enrichedPackets, enriched)
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"enriched": enrichedPackets,
				"errors":   errors,
				"total":    len(enrichedPackets),
			},
		})
	}).Methods("POST")
}