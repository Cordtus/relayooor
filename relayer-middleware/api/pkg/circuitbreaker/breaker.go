package circuitbreaker

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"relayooor/api/pkg/logging"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = fmt.Errorf("circuit breaker is open")

type CircuitBreaker struct {
	name            string
	maxFailures     int
	baseMaxFailures int
	resetTimeout    time.Duration
	halfOpenCalls   int

	mu              sync.Mutex
	state           State
	failures        int
	lastFailureTime time.Time
	successCount    int

	logger *zap.Logger
}

type AdaptiveCircuitBreaker struct {
	*CircuitBreaker
	timeOfDay map[int]int // Hour -> failure threshold
}

func New(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:            name,
		maxFailures:     maxFailures,
		baseMaxFailures: maxFailures,
		resetTimeout:    resetTimeout,
		halfOpenCalls:   3, // Allow 3 calls in half-open state
		state:           StateClosed,
		logger:          logging.With(zap.String("component", "circuit_breaker"), zap.String("name", name)),
	}
}

func NewAdaptive(name string, baseMaxFailures int, resetTimeout time.Duration) *AdaptiveCircuitBreaker {
	cb := &AdaptiveCircuitBreaker{
		CircuitBreaker: &CircuitBreaker{
			name:            name,
			maxFailures:     baseMaxFailures,
			baseMaxFailures: baseMaxFailures,
			resetTimeout:    resetTimeout,
			halfOpenCalls:   3,
			state:           StateClosed,
			logger:          logging.With(zap.String("component", "circuit_breaker"), zap.String("name", name)),
		},
		timeOfDay: make(map[int]int),
	}

	// Start threshold updater
	go cb.updateThresholdsLoop()

	return cb
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from Open to Half-Open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.logger.Info("Circuit breaker transitioning to half-open")
		}
	}

	state := cb.state
	cb.mu.Unlock()

	// Reject calls if circuit is open
	if state == StateOpen {
		return ErrCircuitOpen
	}

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.logger.Error("Circuit breaker opened",
				zap.Int("failures", cb.failures),
				zap.Duration("reset_timeout", cb.resetTimeout),
			)
		}

	case StateHalfOpen:
		// Single failure in half-open state reopens the circuit
		cb.state = StateOpen
		cb.failures = cb.maxFailures
		cb.logger.Warn("Circuit breaker reopened from half-open state")
	}
}

func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenCalls {
			// Gradual ramp-up to prevent thundering herd
			cb.state = StateClosed
			cb.failures = 0
			cb.successCount = 0
			cb.maxFailures = cb.baseMaxFailures // Reset to base threshold
			cb.logger.Info("Circuit breaker closed after successful recovery")
		}
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
	cb.logger.Info("Circuit breaker manually reset")
}

// Adaptive circuit breaker methods
func (acb *AdaptiveCircuitBreaker) updateThresholds() {
	hour := time.Now().Hour()

	acb.mu.Lock()
	defer acb.mu.Unlock()

	// Adjust thresholds based on time of day
	if threshold, ok := acb.timeOfDay[hour]; ok {
		acb.maxFailures = threshold
	} else {
		// Default: higher threshold during peak hours
		if hour >= 9 && hour <= 17 {
			acb.maxFailures = acb.baseMaxFailures * 2
		} else {
			acb.maxFailures = acb.baseMaxFailures
		}
	}
}

func (acb *AdaptiveCircuitBreaker) updateThresholdsLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		acb.updateThresholds()
	}
}

func (acb *AdaptiveCircuitBreaker) SetTimeOfDayThreshold(hour int, threshold int) {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	acb.timeOfDay[hour] = threshold
	acb.logger.Info("Set time-of-day threshold",
		zap.Int("hour", hour),
		zap.Int("threshold", threshold),
	)
}

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}