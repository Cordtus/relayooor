package clearing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ExecutionService handles the actual packet clearing through Hermes
type ExecutionService struct {
	service   *Service
	hermesURL string
	workers   int
}

// NewExecutionService creates a new execution service
func NewExecutionService(service *Service, workers int) *ExecutionService {
	if workers <= 0 {
		workers = 2
	}
	
	return &ExecutionService{
		service:   service,
		hermesURL: service.hermesURL,
		workers:   workers,
	}
}

// Start begins processing the execution queue
func (es *ExecutionService) Start(ctx context.Context) {
	for i := 0; i < es.workers; i++ {
		go es.worker(ctx, i)
	}
}

// worker processes tokens from the execution queue
func (es *ExecutionService) worker(ctx context.Context, workerID int) {
	log.Printf("Execution worker %d started", workerID)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Execution worker %d stopped", workerID)
			return
		default:
			// Get token from queue
			tokenID, err := es.service.redisClient.BRPop(ctx, 5*time.Second, "clearing:execution:queue").Result()
			if err != nil {
				continue // Timeout or error, try again
			}
			
			if len(tokenID) < 2 {
				continue
			}
			
			// Process the token
			if err := es.processToken(ctx, tokenID[1]); err != nil {
				log.Printf("Worker %d failed to process token %s: %v", workerID, tokenID[1], err)
				// Could implement retry logic here
			}
		}
	}
}

// processToken executes clearing for a specific token
func (es *ExecutionService) processToken(ctx context.Context, tokenID string) error {
	// Get token details
	token, err := es.service.getToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	
	// Verify payment
	paymentKey := fmt.Sprintf("clearing:payment:%s", tokenID)
	paymentExists, err := es.service.redisClient.Exists(ctx, paymentKey).Result()
	if err != nil || paymentExists == 0 {
		return fmt.Errorf("payment not verified")
	}
	
	// Start execution tracking
	execution := &ExecutionInfo{
		StartedAt: timePtr(time.Now()),
	}
	
	if err := es.updateExecutionStatus(ctx, tokenID, execution); err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}
	
	// Execute based on type
	var result *ClearingResult
	startTime := time.Now()
	
	switch token.RequestType {
	case "packet":
		result, err = es.clearPackets(ctx, token.TargetIdentifiers.Packets)
	case "channel":
		result, err = es.clearChannel(ctx, token.TargetIdentifiers.Channels[0])
	case "bulk":
		result, err = es.clearBulkChannels(ctx, token.TargetIdentifiers.Channels)
	default:
		err = fmt.Errorf("unknown request type: %s", token.RequestType)
	}
	
	// Update execution status
	execution.CompletedAt = timePtr(time.Now())
	if err != nil {
		execution.Error = err.Error()
	} else if result != nil {
		execution.PacketsCleared = result.PacketsCleared
		execution.PacketsFailed = result.PacketsFailed
		execution.TxHashes = result.TxHashes
	}
	
	if err := es.updateExecutionStatus(ctx, tokenID, execution); err != nil {
		log.Printf("Failed to update final execution status: %v", err)
	}
	
	// Record operation in database (would be PostgreSQL in production)
	operation := &ClearingOperation{
		Token:            tokenID,
		WalletAddress:    token.WalletAddress,
		OperationType:    token.RequestType,
		PacketsTargeted:  len(token.TargetIdentifiers.Packets),
		PacketsCleared:   execution.PacketsCleared,
		PacketsFailed:    execution.PacketsFailed,
		StartedAt:        *execution.StartedAt,
		CompletedAt:      execution.CompletedAt,
		DurationMs:       int(time.Since(startTime).Milliseconds()),
		Success:          err == nil,
		ErrorMessage:     execution.Error,
		ExecutionTxHashes: execution.TxHashes,
	}
	
	// Store operation (in Redis for now, would be PostgreSQL)
	if err := es.storeOperation(ctx, operation); err != nil {
		log.Printf("Failed to store operation: %v", err)
	}
	
	// Update user statistics
	if err := es.updateUserStats(ctx, token.WalletAddress, operation); err != nil {
		log.Printf("Failed to update user stats: %v", err)
	}
	
	// Broadcast update via WebSocket
	es.broadcastUpdate(tokenID, execution)
	
	return err
}

// ClearingResult represents the result of a clearing operation
type ClearingResult struct {
	PacketsCleared int
	PacketsFailed  int
	TxHashes       []string
}

// clearPackets clears specific packets
func (es *ExecutionService) clearPackets(ctx context.Context, packets []PacketIdentifier) (*ClearingResult, error) {
	result := &ClearingResult{
		TxHashes: []string{},
	}
	
	// Group packets by channel for efficiency
	packetsByChannel := make(map[string][]PacketIdentifier)
	for _, packet := range packets {
		key := fmt.Sprintf("%s:%s", packet.Chain, packet.Channel)
		packetsByChannel[key] = append(packetsByChannel[key], packet)
	}
	
	// Clear packets by channel group
	for channelKey, channelPackets := range packetsByChannel {
		// Call Hermes API to clear packets
		hermesReq := map[string]interface{}{
			"chain_id": channelPackets[0].Chain,
			"channel":  channelPackets[0].Channel,
			"sequences": func() []uint64 {
				seqs := make([]uint64, len(channelPackets))
				for i, p := range channelPackets {
					seqs[i] = p.Sequence
				}
				return seqs
			}(),
		}
		
		resp, err := es.callHermesAPI("POST", "/clear-packets", hermesReq)
		if err != nil {
			log.Printf("Failed to clear packets for channel %s: %v", channelKey, err)
			result.PacketsFailed += len(channelPackets)
			continue
		}
		
		// Parse response
		if txHash, ok := resp["tx_hash"].(string); ok {
			result.TxHashes = append(result.TxHashes, txHash)
			result.PacketsCleared += len(channelPackets)
		} else {
			result.PacketsFailed += len(channelPackets)
		}
	}
	
	return result, nil
}

// clearChannel clears all pending packets on a channel
func (es *ExecutionService) clearChannel(ctx context.Context, channel ChannelPair) (*ClearingResult, error) {
	// Query Chainpulse for stuck packets on this channel
	stuckPackets, err := es.queryStuckPackets(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to query stuck packets: %w", err)
	}
	
	if len(stuckPackets) == 0 {
		return &ClearingResult{}, nil
	}
	
	// Convert to packet identifiers
	packets := make([]PacketIdentifier, len(stuckPackets))
	for i, sp := range stuckPackets {
		packets[i] = PacketIdentifier{
			Chain:    channel.SrcChain,
			Channel:  channel.SrcChannel,
			Sequence: sp.Sequence,
		}
	}
	
	return es.clearPackets(ctx, packets)
}

// clearBulkChannels clears multiple channels
func (es *ExecutionService) clearBulkChannels(ctx context.Context, channels []ChannelPair) (*ClearingResult, error) {
	totalResult := &ClearingResult{
		TxHashes: []string{},
	}
	
	for _, channel := range channels {
		result, err := es.clearChannel(ctx, channel)
		if err != nil {
			log.Printf("Failed to clear channel %s->%s: %v", 
				channel.SrcChannel, channel.DstChannel, err)
			continue
		}
		
		totalResult.PacketsCleared += result.PacketsCleared
		totalResult.PacketsFailed += result.PacketsFailed
		totalResult.TxHashes = append(totalResult.TxHashes, result.TxHashes...)
	}
	
	return totalResult, nil
}

// Helper functions

func (es *ExecutionService) callHermesAPI(method string, endpoint string, data interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}
	
	req, err := http.NewRequest(method, es.hermesURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hermes API returned status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

type StuckPacket struct {
	Sequence uint64 `json:"sequence"`
	Age      int    `json:"age_seconds"`
}

func (es *ExecutionService) queryStuckPackets(channel ChannelPair) ([]StuckPacket, error) {
	// Query Chainpulse API for stuck packets
	chainpulseURL := "http://chainpulse:3001" // Would be configurable
	
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/packets/stuck?src_channel=%s&dst_channel=%s",
		chainpulseURL, channel.SrcChannel, channel.DstChannel))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var response struct {
		Packets []StuckPacket `json:"packets"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	return response.Packets, nil
}

func (es *ExecutionService) updateExecutionStatus(ctx context.Context, tokenID string, execution *ExecutionInfo) error {
	executionKey := fmt.Sprintf("clearing:execution:%s", tokenID)
	executionData, err := json.Marshal(execution)
	if err != nil {
		return err
	}
	
	return es.service.redisClient.Set(ctx, executionKey, executionData, 24*time.Hour).Err()
}

func (es *ExecutionService) storeOperation(ctx context.Context, operation *ClearingOperation) error {
	// Store in Redis for now (would be PostgreSQL in production)
	operationKey := fmt.Sprintf("clearing:operation:%s", operation.Token)
	operationData, err := json.Marshal(operation)
	if err != nil {
		return err
	}
	
	// Store operation
	if err := es.service.redisClient.Set(ctx, operationKey, operationData, 30*24*time.Hour).Err(); err != nil {
		return err
	}
	
	// Add to user's operation list
	userOpsKey := fmt.Sprintf("clearing:user:%s:operations", operation.WalletAddress)
	return es.service.redisClient.LPush(ctx, userOpsKey, operation.Token).Err()
}

func (es *ExecutionService) updateUserStats(ctx context.Context, wallet string, operation *ClearingOperation) error {
	statsKey := fmt.Sprintf("clearing:user:%s:stats", wallet)
	
	// Get current stats
	var stats UserStatistics
	statsData, err := es.service.redisClient.Get(ctx, statsKey).Result()
	if err == nil {
		json.Unmarshal([]byte(statsData), &stats)
	}
	
	// Update stats
	stats.Wallet = wallet
	stats.TotalRequests++
	if operation.Success {
		stats.SuccessfulClears++
		stats.TotalPacketsCleared += operation.PacketsCleared
	} else {
		stats.FailedClears++
	}
	
	// Calculate success rate
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulClears) / float64(stats.TotalRequests)
	}
	
	// Update average clear time
	if operation.Success && operation.DurationMs > 0 {
		if stats.AvgClearTime == 0 {
			stats.AvgClearTime = operation.DurationMs
		} else {
			stats.AvgClearTime = (stats.AvgClearTime + operation.DurationMs) / 2
		}
	}
	
	// Store updated stats
	updatedData, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	
	return es.service.redisClient.Set(ctx, statsKey, updatedData, 0).Err()
}

func (es *ExecutionService) broadcastUpdate(tokenID string, execution *ExecutionInfo) {
	// This would send updates via WebSocket
	// Implementation depends on WebSocket handler setup
	update := map[string]interface{}{
		"type":      "clearing_update",
		"token":     tokenID,
		"execution": execution,
		"timestamp": time.Now().Unix(),
	}
	
	log.Printf("Broadcasting update for token %s: %+v", tokenID, update)
}

func timePtr(t time.Time) *time.Time {
	return &t
}