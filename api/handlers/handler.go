package handlers

import (
	"github.com/relayooor/api/pkg/chainpulse"
)

// Handler holds dependencies for all handlers
type Handler struct {
	chainpulseClient *chainpulse.Client
}

// NewHandler creates a new handler with dependencies
func NewHandler(chainpulseURL string) *Handler {
	return &Handler{
		chainpulseClient: chainpulse.NewClient(chainpulseURL),
	}
}