package clearing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// hermesClient implements the HermesClient interface for interacting with Hermes REST API
type hermesClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHermesClient creates a new Hermes client
func NewHermesClient(baseURL string) HermesClient {
	return &hermesClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ClearPacketsRequest represents a request to clear packets
type HermesClearRequest struct {
	ChainID string   `json:"chain_id"`
	Port    string   `json:"port"`
	Channel string   `json:"channel"`
	Sequences []uint64 `json:"sequences,omitempty"`
}

// ClearPackets sends a request to Hermes to clear specific packets
func (c *hermesClient) ClearPackets(ctx context.Context, req *ClearPacketsRequest) (*ClearPacketsResponse, error) {
	// Construct the Hermes API request
	hermesReq := HermesClearRequest{
		ChainID:   req.Chain,
		Port:      req.Port,
		Channel:   req.Channel,
		Sequences: req.Sequences,
	}

	// Marshal the request
	body, err := json.Marshal(hermesReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/clear_packets", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hermes returned status %d", resp.StatusCode)
	}

	// For now, return a simple success response
	// In a real implementation, we would parse Hermes response and extract transaction details
	return &ClearPacketsResponse{
		Success:  true,
		TxHashes: []string{"mock-tx-hash"}, // Would be populated from actual Hermes response
		Error:    "",
	}, nil
}

// GetVersion retrieves the Hermes version information
func (c *hermesClient) GetVersion(ctx context.Context) (*VersionResponse, error) {
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/version", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hermes returned status %d", resp.StatusCode)
	}

	// Parse response
	var version VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &version, nil
}