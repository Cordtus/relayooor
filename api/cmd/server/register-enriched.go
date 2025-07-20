package main

import (
	"github.com/gorilla/mux"
)

// AddEnrichedEndpoints adds enriched packet endpoints to existing router
func AddEnrichedEndpoints(r *mux.Router, chainpulseURL string) {
	// Initialize packet enrichment service
	enrichmentService := NewPacketEnrichmentService(chainpulseURL, "http://localhost:5185")
	
	// Register the enriched endpoints
	RegisterEnrichedEndpoints(r, enrichmentService)
}