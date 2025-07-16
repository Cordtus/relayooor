package clearing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Mock relayer client
type MockRelayerClient struct {
	mock.Mock
}

func (m *MockRelayerClient) ClearPacket(ctx context.Context, chain, channel string, sequence uint64) error {
	args := m.Called(ctx, chain, channel, sequence)
	return args.Error(0)
}

// Mock operation tracker
type MockOperationTracker struct {
	mock.Mock
}

func (m *MockOperationTracker) StartOperation(id string) {
	m.Called(id)
}

func (m *MockOperationTracker) CompleteOperation(id string) {
	m.Called(id)
}

func (m *MockOperationTracker) WaitForCompletion(timeout time.Duration) bool {
	args := m.Called(timeout)
	return args.Bool(0)
}

// Test setup
func setupTestExecutionService(t *testing.T) (*ExecutionServiceV2, *MockRelayerClient) {
	logger, _ := zap.NewDevelopment()
	
	mockRelayer := &MockRelayerClient{}
	mockTracker := &MockOperationTracker{}
	
	// Create circuit breaker with test settings
	cbSettings := gobreaker.Settings{
		Name:        "test-circuit-breaker",
		MaxRequests: 2,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}
	
	service := &ExecutionServiceV2{
		circuitBreaker: gobreaker.NewCircuitBreaker(cbSettings),
		retrier:        NewRetrier(3, 100*time.Millisecond, 2.0),
		refundService:  &MockRefundService{},
		tracker:        mockTracker,
		relayerClient:  mockRelayer,
		logger:         logger,
	}
	
	return service, mockRelayer
}

// Test successful packet clearing
func TestExecutePacketClearingSuccess(t *testing.T) {
	service, mockRelayer := setupTestExecutionService(t)
	ctx := context.Background()
	
	op := &ClearingOperation{
		ID:    "test-op-1",
		Token: "test-token",
		Targets: map[string]interface{}{
			"packets": []map[string]interface{}{
				{
					"chain":    "osmosis-1",
					"channel":  "channel-0",
					"sequence": float64(123),
				},
			},
		},
	}
	
	// Mock successful clearing
	mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(123)).Return(nil)
	
	// Mock tracker
	tracker := service.tracker.(*MockOperationTracker)
	tracker.On("StartOperation", "clear-osmosis-1-channel-0-123").Return()
	tracker.On("CompleteOperation", "clear-osmosis-1-channel-0-123").Return()
	
	// Execute
	result, err := service.ExecutePacketClearing(ctx, op)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.PacketsCleared)
	assert.Len(t, result.FailedPackets, 0)
	
	mockRelayer.AssertExpectations(t)
	tracker.AssertExpectations(t)
}

// Test packet clearing with retry
func TestExecutePacketClearingWithRetry(t *testing.T) {
	service, mockRelayer := setupTestExecutionService(t)
	ctx := context.Background()
	
	op := &ClearingOperation{
		ID:    "test-op-2",
		Token: "test-token",
		Targets: map[string]interface{}{
			"packets": []map[string]interface{}{
				{
					"chain":    "osmosis-1",
					"channel":  "channel-0",
					"sequence": float64(123),
				},
			},
		},
	}
	
	// Mock first two attempts fail, third succeeds
	mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(123)).
		Return(errors.New("temporary error")).Twice()
	mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(123)).
		Return(nil).Once()
	
	// Mock tracker
	tracker := service.tracker.(*MockOperationTracker)
	tracker.On("StartOperation", mock.Anything).Return()
	tracker.On("CompleteOperation", mock.Anything).Return()
	
	// Execute
	result, err := service.ExecutePacketClearing(ctx, op)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.PacketsCleared)
	
	// Verify retry happened
	mockRelayer.AssertNumberOfCalls(t, "ClearPacket", 3)
}

// Test circuit breaker open
func TestExecutePacketClearingCircuitBreakerOpen(t *testing.T) {
	service, mockRelayer := setupTestExecutionService(t)
	ctx := context.Background()
	
	// Cause circuit breaker to open by failing multiple times
	for i := 0; i < 5; i++ {
		op := &ClearingOperation{
			ID:    "test-op-fail-" + string(rune(i)),
			Token: "test-token",
			Targets: map[string]interface{}{
				"packets": []map[string]interface{}{
					{
						"chain":    "osmosis-1",
						"channel":  "channel-0",
						"sequence": float64(i),
					},
				},
			},
		}
		
		mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(i)).
			Return(errors.New("service unavailable"))
		
		// Mock tracker
		tracker := service.tracker.(*MockOperationTracker)
		tracker.On("StartOperation", mock.Anything).Return()
		
		_, _ = service.ExecutePacketClearing(ctx, op)
	}
	
	// Now circuit breaker should be open
	op := &ClearingOperation{
		ID:    "test-op-blocked",
		Token: "test-token",
		Targets: map[string]interface{}{
			"packets": []map[string]interface{}{
				{
					"chain":    "osmosis-1",
					"channel":  "channel-0",
					"sequence": float64(999),
				},
			},
		},
	}
	
	// This call should fail immediately due to circuit breaker
	result, err := service.ExecutePacketClearing(ctx, op)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
	assert.Nil(t, result)
	
	// Verify the last call wasn't made to the relayer
	mockRelayer.AssertNotCalled(t, "ClearPacket", ctx, "osmosis-1", "channel-0", uint64(999))
}

// Test concurrent execution
func TestConcurrentPacketClearing(t *testing.T) {
	service, mockRelayer := setupTestExecutionService(t)
	ctx := context.Background()
	
	// Setup multiple operations
	numOps := 10
	operations := make([]*ClearingOperation, numOps)
	
	for i := 0; i < numOps; i++ {
		operations[i] = &ClearingOperation{
			ID:    "test-op-concurrent-" + string(rune(i)),
			Token: "test-token-" + string(rune(i)),
			Targets: map[string]interface{}{
				"packets": []map[string]interface{}{
					{
						"chain":    "osmosis-1",
						"channel":  "channel-0",
						"sequence": float64(i),
					},
				},
			},
		}
		
		// Mock successful clearing
		mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(i)).Return(nil)
	}
	
	// Mock tracker for all operations
	tracker := service.tracker.(*MockOperationTracker)
	tracker.On("StartOperation", mock.Anything).Return()
	tracker.On("CompleteOperation", mock.Anything).Return()
	
	// Execute concurrently
	results := make(chan *ExecutionResult, numOps)
	errors := make(chan error, numOps)
	
	for _, op := range operations {
		go func(operation *ClearingOperation) {
			result, err := service.ExecutePacketClearing(ctx, operation)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(op)
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < numOps; i++ {
		select {
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		case result := <-results:
			assert.Equal(t, 1, result.PacketsCleared)
			successCount++
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for results")
		}
	}
	
	assert.Equal(t, numOps, successCount)
}

// Test graceful shutdown
func TestGracefulShutdown(t *testing.T) {
	service, mockRelayer := setupTestExecutionService(t)
	ctx := context.Background()
	
	// Create a long-running operation
	op := &ClearingOperation{
		ID:    "test-op-shutdown",
		Token: "test-token",
		Targets: map[string]interface{}{
			"packets": []map[string]interface{}{
				{
					"chain":    "osmosis-1",
					"channel":  "channel-0",
					"sequence": float64(123),
				},
			},
		},
	}
	
	// Mock slow clearing operation
	mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(123)).
		Run(func(args mock.Arguments) {
			time.Sleep(2 * time.Second)
		}).
		Return(nil)
	
	// Mock tracker
	tracker := service.tracker.(*MockOperationTracker)
	tracker.On("StartOperation", mock.Anything).Return()
	tracker.On("CompleteOperation", mock.Anything).Return()
	tracker.On("WaitForCompletion", 10*time.Second).Return(true)
	
	// Start operation in background
	go func() {
		_, _ = service.ExecutePacketClearing(ctx, op)
	}()
	
	// Give it time to start
	time.Sleep(100 * time.Millisecond)
	
	// Initiate shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := service.Shutdown(shutdownCtx)
	
	assert.NoError(t, err)
	tracker.AssertCalled(t, "WaitForCompletion", 10*time.Second)
}

// Test retry configuration
func TestRetryConfiguration(t *testing.T) {
	retrier := NewRetrier(3, 100*time.Millisecond, 2.0)
	
	attempts := 0
	start := time.Now()
	
	err := retrier.Do(func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	})
	
	duration := time.Since(start)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
	// Should take at least 100ms + 200ms = 300ms (with backoff)
	assert.Greater(t, duration, 250*time.Millisecond)
}