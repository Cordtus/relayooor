package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrNonIdempotentRetry = errors.New("cannot retry non-idempotent operation")
	ErrNetworkTimeout     = errors.New("network timeout")
	ErrRPCUnavailable     = errors.New("RPC unavailable")
)

type Config struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomFactor    float64 // Jitter factor (0.0 to 1.0)
}

type OperationState struct {
	mu       sync.Mutex
	attempts map[string]int
	versions map[string]int
}

type Retrier struct {
	config Config
	logger *zap.Logger
	state  *OperationState
}

func NewRetrier(config Config, logger *zap.Logger) *Retrier {
	return &Retrier{
		config: config,
		logger: logger,
		state: &OperationState{
			attempts: make(map[string]int),
			versions: make(map[string]int),
		},
	}
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts:     3,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		RandomFactor:    0.1,
	}
}

func (r *Retrier) Do(ctx context.Context, operation string, fn func() error) error {
	// Check if operation is idempotent
	operationID := fmt.Sprintf("%s:%s", operation, generateOperationID())
	if !r.isIdempotent(operation) {
		// Track operation state
		r.state.mu.Lock()
		if r.state.attempts[operationID] > 0 {
			r.state.mu.Unlock()
			return ErrNonIdempotentRetry
		}
		r.state.attempts[operationID]++
		r.state.mu.Unlock()
	}

	var lastErr error

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		// Check context
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		// Execute function
		if err := fn(); err != nil {
			lastErr = err

			// Check if error is retryable
			if !isRetryable(err) {
				r.logger.Debug("Error is not retryable",
					zap.String("operation", operation),
					zap.Error(err),
				)
				return err
			}

			// Calculate next interval
			interval := r.calculateInterval(attempt)

			r.logger.Info("Operation failed, retrying",
				zap.String("operation", operation),
				zap.Int("attempt", attempt+1),
				zap.Duration("retry_in", interval),
				zap.Error(err),
			)

			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
				continue
			}
		}

		// Success
		if attempt > 0 {
			r.logger.Info("Operation succeeded after retry",
				zap.String("operation", operation),
				zap.Int("attempts", attempt+1),
			)
		}
		return nil
	}

	return fmt.Errorf("operation failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

func (r *Retrier) calculateInterval(attempt int) time.Duration {
	// Exponential backoff
	interval := float64(r.config.InitialInterval) * math.Pow(r.config.Multiplier, float64(attempt))

	// Add jitter
	if r.config.RandomFactor > 0 {
		delta := interval * r.config.RandomFactor
		minInterval := interval - delta
		maxInterval := interval + delta

		// Generate random value in range
		rand.Seed(time.Now().UnixNano())
		interval = minInterval + (rand.Float64() * (maxInterval - minInterval))
	}

	// Cap at max interval
	if interval > float64(r.config.MaxInterval) {
		interval = float64(r.config.MaxInterval)
	}

	return time.Duration(interval)
}

func (r *Retrier) isIdempotent(operation string) bool {
	idempotentOps := map[string]bool{
		"get_packets":    true,
		"check_status":   true,
		"verify_payment": true,
		"clear_packets":  false, // Not idempotent
		"process_refund": false, // Not idempotent
	}

	return idempotentOps[operation]
}

// Determine if error is retryable
func isRetryable(err error) bool {
	// Define retryable errors
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return true
	case errors.Is(err, ErrNetworkTimeout):
		return true
	case errors.Is(err, ErrRPCUnavailable):
		return true
	default:
		// Check for specific error messages
		errMsg := err.Error()
		retryableMessages := []string{
			"connection refused",
			"connection reset",
			"timeout",
			"temporary failure",
			"unavailable",
		}

		for _, msg := range retryableMessages {
			if strings.Contains(strings.ToLower(errMsg), msg) {
				return true
			}
		}

		return false
	}
}

func generateOperationID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}