package chainpulse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// Client represents a Chainpulse API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new Chainpulse API client
func NewClient(baseURL string, logger *zap.Logger) *Client {
	// Default to correct Chainpulse port if not specified
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger.With(zap.String("component", "chainpulse_client")),
	}
}

// PacketByUser represents a packet associated with a user
type PacketByUser struct {
	Chain     string    `json:"chain"`
	Channel   string    `json:"channel"`
	Sequence  uint64    `json:"sequence"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Amount    string    `json:"amount"`
	Denom     string    `json:"denom"`
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // pending, completed, stuck
}

// StuckPacket represents a stuck IBC packet
type StuckPacket struct {
	SrcChain      string    `json:"src_chain"`
	DstChain      string    `json:"dst_chain"`
	SrcChannel    string    `json:"src_channel"`
	DstChannel    string    `json:"dst_channel"`
	Sequence      uint64    `json:"sequence"`
	Sender        string    `json:"sender"`
	Receiver      string    `json:"receiver"`
	Amount        string    `json:"amount"`
	Denom         string    `json:"denom"`
	StuckSince    time.Time `json:"stuck_since"`
	StuckDuration int       `json:"stuck_duration_minutes"`
}

// ChannelCongestion represents channel congestion status
type ChannelCongestion struct {
	Chain           string  `json:"chain"`
	Channel         string  `json:"channel"`
	CounterpartyChain string `json:"counterparty_chain"`
	CounterpartyChannel string `json:"counterparty_channel"`
	PendingPackets  int     `json:"pending_packets"`
	AvgClearTime    float64 `json:"avg_clear_time_seconds"`
	CongestionLevel string  `json:"congestion_level"` // low, medium, high
}

// PacketDetails represents detailed packet information
type PacketDetails struct {
	Chain         string                 `json:"chain"`
	Channel       string                 `json:"channel"`
	Sequence      uint64                 `json:"sequence"`
	PacketData    map[string]interface{} `json:"packet_data"`
	Status        string                 `json:"status"`
	Acknowledgement string               `json:"acknowledgement,omitempty"`
	TxHash        string                 `json:"tx_hash"`
	Height        uint64                 `json:"height"`
	Timestamp     time.Time              `json:"timestamp"`
}

// GetPacketsByUser retrieves packets for a specific user address
func (c *Client) GetPacketsByUser(ctx context.Context, userAddress string) ([]PacketByUser, error) {
	endpoint := fmt.Sprintf("/api/v1/packets/by-user?address=%s", url.QueryEscape(userAddress))
	
	var packets []PacketByUser
	if err := c.get(ctx, endpoint, &packets); err != nil {
		return nil, fmt.Errorf("failed to get packets by user: %w", err)
	}
	
	return packets, nil
}

// GetStuckPackets retrieves all currently stuck packets
func (c *Client) GetStuckPackets(ctx context.Context, minStuckMinutes int) ([]StuckPacket, error) {
	endpoint := fmt.Sprintf("/api/v1/packets/stuck?min_stuck_minutes=%d", minStuckMinutes)
	
	var packets []StuckPacket
	if err := c.get(ctx, endpoint, &packets); err != nil {
		return nil, fmt.Errorf("failed to get stuck packets: %w", err)
	}
	
	return packets, nil
}

// GetPacketDetails retrieves details for a specific packet
func (c *Client) GetPacketDetails(ctx context.Context, chain, channel string, sequence uint64) (*PacketDetails, error) {
	endpoint := fmt.Sprintf("/api/v1/packets/%s/%s/%d", chain, channel, sequence)
	
	var details PacketDetails
	if err := c.get(ctx, endpoint, &details); err != nil {
		return nil, fmt.Errorf("failed to get packet details: %w", err)
	}
	
	return &details, nil
}

// GetChannelCongestion retrieves congestion status for all channels
func (c *Client) GetChannelCongestion(ctx context.Context) ([]ChannelCongestion, error) {
	endpoint := "/api/v1/channels/congestion"
	
	var congestion []ChannelCongestion
	if err := c.get(ctx, endpoint, &congestion); err != nil {
		return nil, fmt.Errorf("failed to get channel congestion: %w", err)
	}
	
	return congestion, nil
}

// GetMetrics retrieves raw Prometheus metrics
func (c *Client) GetMetrics(ctx context.Context) (string, error) {
	endpoint := "/metrics"
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return "", err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch metrics", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}

// HealthCheck checks if Chainpulse is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	endpoint := "/health"
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// get performs a GET request to the specified endpoint
func (c *Client) get(ctx context.Context, endpoint string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to perform request",
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("Unexpected status code",
			zap.String("endpoint", endpoint),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		c.logger.Error("Failed to decode response",
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
		return err
	}
	
	return nil
}