package chainpulse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is the Chainpulse API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Chainpulse API client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Packet represents an IBC packet with full details
type Packet struct {
	ChainID           string  `json:"chain_id"`
	Sequence          int64   `json:"sequence"`
	SrcChannel        string  `json:"src_channel"`
	DstChannel        string  `json:"dst_channel"`
	Sender            string  `json:"sender"`
	Receiver          string  `json:"receiver"`
	Amount            string  `json:"amount"`
	Denom             string  `json:"denom"`
	AgeSeconds        int64   `json:"age_seconds"`
	RelayAttempts     int     `json:"relay_attempts"`
	LastAttemptBy     string  `json:"last_attempt_by"`
	TimeoutTimestamp  *int64  `json:"timeout_timestamp,omitempty"`
	IBCVersion        string  `json:"ibc_version"`
}

// StuckPacketsResponse represents the response from the stuck packets endpoint
type StuckPacketsResponse struct {
	Packets    []Packet `json:"packets"`
	Total      int      `json:"total"`
	APIVersion string   `json:"api_version"`
}

// ExpiringPacket represents a packet that is about to timeout
type ExpiringPacket struct {
	Packet
	SecondsUntilTimeout int64  `json:"seconds_until_timeout"`
	TimeoutType         string `json:"timeout_type"`
	TimeoutValue        string `json:"timeout_value"`
}

// ExpiringPacketsResponse represents the response from the expiring packets endpoint
type ExpiringPacketsResponse struct {
	Packets    []ExpiringPacket `json:"packets"`
	APIVersion string           `json:"api_version"`
}

// ExpiredPacket represents a packet that has already timed out
type ExpiredPacket struct {
	Packet
	SecondsSinceTimeout int64  `json:"seconds_since_timeout"`
	TimeoutType         string `json:"timeout_type"`
}

// ExpiredPacketsResponse represents the response from the expired packets endpoint
type ExpiredPacketsResponse struct {
	Packets    []ExpiredPacket `json:"packets"`
	APIVersion string          `json:"api_version"`
}

// ChannelCongestion represents congestion data for a channel
type ChannelCongestion struct {
	SrcChannel             string            `json:"src_channel"`
	DstChannel             string            `json:"dst_channel"`
	StuckCount             int               `json:"stuck_count"`
	OldestStuckAgeSeconds  int64             `json:"oldest_stuck_age_seconds"`
	TotalValue             map[string]string `json:"total_value"`
}

// ChannelCongestionResponse represents the response from the channel congestion endpoint
type ChannelCongestionResponse struct {
	Channels   []ChannelCongestion `json:"channels"`
	APIVersion string              `json:"api_version"`
}

// DuplicatePacket represents a duplicate packet entry
type DuplicatePacket struct {
	DataHash string `json:"data_hash"`
	Count    int    `json:"count"`
	Packets  []struct {
		ChainID    string `json:"chain_id"`
		Sequence   int64  `json:"sequence"`
		SrcChannel string `json:"src_channel"`
		Sender     string `json:"sender"`
		CreatedAt  string `json:"created_at"`
	} `json:"packets"`
}

// DuplicatesResponse represents the response from the duplicates endpoint
type DuplicatesResponse struct {
	Duplicates []DuplicatePacket `json:"duplicates"`
	APIVersion string            `json:"api_version"`
}

// GetStuckPackets retrieves stuck packets with optional filters
func (c *Client) GetStuckPackets(minAgeSeconds int, limit int) (*StuckPacketsResponse, error) {
	params := url.Values{}
	if minAgeSeconds > 0 {
		params.Set("min_age_seconds", fmt.Sprintf("%d", minAgeSeconds))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	url := fmt.Sprintf("%s/api/v1/packets/stuck?%s", c.baseURL, params.Encode())
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stuck packets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result StuckPacketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetPacketsByUser retrieves packets for a specific user address
func (c *Client) GetPacketsByUser(address string, role string, limit int, offset int) (*StuckPacketsResponse, error) {
	params := url.Values{}
	params.Set("address", address)
	if role != "" {
		params.Set("role", role)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	url := fmt.Sprintf("%s/api/v1/packets/by-user?%s", c.baseURL, params.Encode())
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user packets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result StuckPacketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetExpiringPackets retrieves packets that will timeout soon
func (c *Client) GetExpiringPackets(minutes int) (*ExpiringPacketsResponse, error) {
	params := url.Values{}
	if minutes > 0 {
		params.Set("minutes", fmt.Sprintf("%d", minutes))
	}

	url := fmt.Sprintf("%s/api/v1/packets/expiring?%s", c.baseURL, params.Encode())
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expiring packets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ExpiringPacketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetExpiredPackets retrieves packets that have already timed out
func (c *Client) GetExpiredPackets() (*ExpiredPacketsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/packets/expired", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expired packets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ExpiredPacketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetChannelCongestion retrieves congestion data for all channels
func (c *Client) GetChannelCongestion() (*ChannelCongestionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/channels/congestion", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel congestion: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ChannelCongestionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetPacketDetails retrieves details for a specific packet
func (c *Client) GetPacketDetails(chainID, channel string, sequence int64) (*Packet, error) {
	url := fmt.Sprintf("%s/api/v1/packets/%s/%s/%d", c.baseURL, chainID, channel, sequence)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packet details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("packet not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var packet Packet
	if err := json.NewDecoder(resp.Body).Decode(&packet); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &packet, nil
}

// GetDuplicatePackets retrieves packets with duplicate data
func (c *Client) GetDuplicatePackets() (*DuplicatesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/packets/duplicates", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch duplicate packets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result DuplicatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HealthCheck verifies the Chainpulse API is accessible
func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("%s/metrics", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}