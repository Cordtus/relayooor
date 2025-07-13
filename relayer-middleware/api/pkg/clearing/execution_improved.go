package clearing

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"relayooor/api/pkg/circuitbreaker"
	"relayooor/api/pkg/retry"
)

var (
	ErrChannelClosed     = errors.New("channel closed")
	ErrHermesUnavailable = errors.New("hermes unavailable")
	ErrInsufficientGas   = errors.New("insufficient gas")
)

// ExecutionServiceV2 handles the actual packet clearing through Hermes with improved error handling
type ExecutionServiceV2 struct {
	db             *gorm.DB
	redisClient    *redis.Client
	hermesClient   HermesClient
	logger         *zap.Logger
	workerPool     chan struct{}
	activeTasks    sync.WaitGroup
	circuitBreaker *circuitbreaker.CircuitBreaker
	retrier        *retry.Retrier
	refundService  *RefundService
	tracker        *OperationTracker
}

// HermesClient interface for Hermes interactions
type HermesClient interface {
	ClearPackets(ctx context.Context, req *ClearPacketsRequest) (*ClearPacketsResponse, error)
	GetVersion(ctx context.Context) (*VersionResponse, error)
}

type ClearPacketsRequest struct {
	Chain     string   `json:"chain"`
	Channel   string   `json:"channel"`
	Port      string   `json:"port"`
	Sequences []uint64 `json:"sequences"`
}

type ClearPacketsResponse struct {
	Success  bool     `json:"success"`
	TxHashes []string `json:"tx_hashes"`
	Error    string   `json:"error,omitempty"`
}

type VersionResponse struct {
	Version string `json:"version"`
}

type OperationTracker interface {
	Add(op *ActiveOperation)
	Remove(id string)
}

type ActiveOperation struct {
	ID        string
	StartTime time.Time
	Type      string
}

func NewExecutionServiceV2(
	db *gorm.DB,
	redis *redis.Client,
	hermesClient HermesClient,
	refundService *RefundService,
	tracker *OperationTracker,
	logger *zap.Logger,
) *ExecutionServiceV2 {
	// Wrap Hermes client with circuit breaker
	cbClient := NewCircuitBreakerClient(hermesClient)

	return &ExecutionServiceV2{
		db:             db,
		redisClient:    redis,
		hermesClient:   cbClient,
		logger:         logger.With(zap.String("component", "execution")),
		workerPool:     make(chan struct{}, 5), // Max 5 concurrent executions
		circuitBreaker: circuitbreaker.New("hermes", 5, 30*time.Second),
		retrier:        retry.NewRetrier(retry.DefaultConfig(), logger),
		refundService:  refundService,
		tracker:        tracker,
	}
}

func (es *ExecutionServiceV2) Start(ctx context.Context) {
	// Start execution queue processor
	go es.processQueue(ctx)
}

func (es *ExecutionServiceV2) processQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			es.logger.Info("Stopping execution processor")
			es.activeTasks.Wait()
			return
		default:
			// Get next operation from queue
			tokenID, err := es.redisClient.BLPop(ctx, 5*time.Second, "clearing:execution:queue").Result()
			if err != nil {
				if err != redis.Nil {
					es.logger.Error("Failed to get from queue", zap.Error(err))
				}
				continue
			}

			if len(tokenID) < 2 {
				continue
			}

			// Process in goroutine with worker pool
			es.workerPool <- struct{}{} // Acquire worker slot
			es.activeTasks.Add(1)

			go func(token string) {
				defer func() {
					<-es.workerPool // Release worker slot
					es.activeTasks.Done()
				}()

				// Track operation
				op := &ActiveOperation{
					ID:        token,
					StartTime: time.Now(),
					Type:      "clearing",
				}
				es.tracker.Add(op)
				defer es.tracker.Remove(op.ID)

				if err := es.executeClearing(ctx, token); err != nil {
					es.logger.Error("Failed to execute clearing",
						zap.String("token", token),
						zap.Error(err),
					)
				}
			}(tokenID[1])
		}
	}
}

func (es *ExecutionServiceV2) executeClearing(ctx context.Context, tokenID string) error {
	// Get operation details
	operation, err := es.getOperation(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}

	// Update status to processing
	if err := es.updateOperationStatus(operation.ID, "processing", ""); err != nil {
		es.logger.Error("Failed to update status", zap.Error(err))
	}

	// Clear packets with retry logic
	result, err := es.clearPackets(ctx, operation.Packets)
	if err != nil {
		es.handleClearingFailure(ctx, operation.ID, err)
		return err
	}

	// Update operation as successful
	if err := es.completeOperation(operation.ID, result); err != nil {
		es.logger.Error("Failed to update operation", zap.Error(err))
	}

	// Broadcast success via WebSocket
	es.broadcastStatus(tokenID, ClearingStatus{
		Status:    "completed",
		Message:   "Packets cleared successfully",
		Progress:  100,
		UpdatedAt: time.Now(),
		TxHashes:  result.TxHashes,
	})

	return nil
}

func (es *ExecutionServiceV2) clearPackets(ctx context.Context, packets []PacketIdentifier) (*ClearingResult, error) {
	retrier := retry.NewRetrier(retry.Config{
		MaxAttempts:     3,
		InitialInterval: 2 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		RandomFactor:    0.2,
	}, es.logger)

	var result *ClearingResult

	err := retrier.Do(ctx, "clear_packets", func() error {
		// Group packets by channel
		channelGroups := es.groupPacketsByChannel(packets)

		result = &ClearingResult{
			Success:   true,
			TxHashes:  make([]string, 0),
			Timestamp: time.Now(),
		}

		for channel, channelPackets := range channelGroups {
			// Clear packets for this channel
			resp, err := es.hermesClient.ClearPackets(ctx, &ClearPacketsRequest{
				Chain:     channel.ChainID,
				Channel:   channel.ChannelID,
				Port:      channel.PortID,
				Sequences: channelPackets,
			})

			if err != nil {
				return fmt.Errorf("failed to clear packets on channel %s: %w", channel.ChannelID, err)
			}

			if resp.Success {
				result.TxHashes = append(result.TxHashes, resp.TxHashes...)
			} else {
				result.Success = false
				result.Error = resp.Error
				return fmt.Errorf("clearing failed: %s", resp.Error)
			}
		}

		return nil
	})

	return result, err
}

func (es *ExecutionServiceV2) handleClearingFailure(ctx context.Context, operationID string, err error) {
	logger := es.logger.With(
		zap.String("operation_id", operationID),
		zap.Error(err),
	)

	// Determine if failure is refundable
	refundReason := es.determineRefundReason(err)
	if refundReason == "" {
		logger.Info("Failure not eligible for refund")
		return
	}

	// Mark operation for refund
	if err := es.db.Model(&ClearingOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]interface{}{
			"refund_status": "pending",
			"refund_reason": refundReason,
		}).Error; err != nil {
		logger.Error("Failed to mark operation for refund", zap.Error(err))
	}

	// Trigger refund processing
	go func() {
		refundCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := es.refundService.ProcessRefund(refundCtx, operationID, refundReason); err != nil {
			logger.Error("Failed to process refund", zap.Error(err))
		}
	}()
}

func (es *ExecutionServiceV2) determineRefundReason(err error) string {
	switch {
	case errors.Is(err, ErrChannelClosed):
		return "Channel closed during clearing"
	case errors.Is(err, ErrHermesUnavailable):
		return "Clearing service temporarily unavailable"
	case errors.Is(err, ErrInsufficientGas):
		return "Insufficient gas for clearing transaction"
	case errors.Is(err, circuitbreaker.ErrCircuitOpen):
		return "Service temporarily unavailable due to high failure rate"
	default:
		return ""
	}
}

func (es *ExecutionServiceV2) groupPacketsByChannel(packets []PacketIdentifier) map[ChannelKey][]uint64 {
	groups := make(map[ChannelKey][]uint64)

	for _, packet := range packets {
		key := ChannelKey{
			ChainID:   packet.ChainID,
			ChannelID: packet.ChannelID,
			PortID:    packet.PortID,
		}

		groups[key] = append(groups[key], packet.Sequence)
	}

	return groups
}

func (es *ExecutionServiceV2) getOperation(ctx context.Context, tokenID string) (*QueuedOperation, error) {
	// Implementation would fetch from database
	return &QueuedOperation{
		ID:      tokenID,
		TokenID: tokenID,
		Packets: []PacketIdentifier{},
	}, nil
}

func (es *ExecutionServiceV2) updateOperationStatus(operationID, status, message string) error {
	return es.db.Model(&ClearingOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]interface{}{
			"status":        status,
			"error_message": message,
			"updated_at":    time.Now(),
		}).Error
}

func (es *ExecutionServiceV2) completeOperation(operationID string, result *ClearingResult) error {
	now := time.Now()
	return es.db.Model(&ClearingOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]interface{}{
			"status":          "completed",
			"success":         result.Success,
			"clearing_tx_hash": result.TxHashes[0], // Store first tx hash
			"completed_at":    &now,
		}).Error
}

func (es *ExecutionServiceV2) broadcastStatus(tokenID string, status ClearingStatus) {
	// Implementation would use WebSocket manager
	es.logger.Info("Broadcasting status update",
		zap.String("token", tokenID),
		zap.String("status", status.Status),
	)
}

// Circuit breaker wrapper for Hermes client
type CircuitBreakerClient struct {
	client  HermesClient
	breaker *circuitbreaker.CircuitBreaker
}

func NewCircuitBreakerClient(client HermesClient) *CircuitBreakerClient {
	return &CircuitBreakerClient{
		client: client,
		breaker: circuitbreaker.New(
			"hermes",
			5,                // Open after 5 failures
			30*time.Second,   // Try again after 30 seconds
		),
	}
}

func (c *CircuitBreakerClient) ClearPackets(ctx context.Context, req *ClearPacketsRequest) (*ClearPacketsResponse, error) {
	var resp *ClearPacketsResponse
	var err error

	circuitErr := c.breaker.Execute(func() error {
		resp, err = c.client.ClearPackets(ctx, req)
		return err
	})

	if circuitErr != nil {
		if circuitErr == circuitbreaker.ErrCircuitOpen {
			return nil, ErrHermesUnavailable
		}
		return nil, circuitErr
	}

	return resp, nil
}

func (c *CircuitBreakerClient) GetVersion(ctx context.Context) (*VersionResponse, error) {
	var resp *VersionResponse
	var err error

	circuitErr := c.breaker.Execute(func() error {
		resp, err = c.client.GetVersion(ctx)
		return err
	})

	if circuitErr != nil {
		if circuitErr == circuitbreaker.ErrCircuitOpen {
			return nil, ErrHermesUnavailable
		}
		return nil, circuitErr
	}

	return resp, nil
}