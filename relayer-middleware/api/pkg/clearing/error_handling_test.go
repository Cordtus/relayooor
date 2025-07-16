package clearing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Test error types
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "validation error",
			err:      ErrInvalidRequest,
			expected: "invalid request",
		},
		{
			name:     "not found error",
			err:      ErrTokenNotFound,
			expected: "token not found",
		},
		{
			name:     "duplicate payment error",
			err:      ErrDuplicatePayment,
			expected: "duplicate payment detected",
		},
		{
			name:     "payment validation error",
			err:      ErrPaymentValidation,
			expected: "payment validation failed",
		},
		{
			name:     "execution error",
			err:      ErrExecutionFailed,
			expected: "execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.err.Error(), tt.expected)
		})
	}
}

// Test database error handling
func TestDatabaseErrorHandling(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	// Close database to simulate connection error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	// Try to request token with closed DB
	req := &ClearingRequest{
		WalletAddress: "osmo1test123",
		ChainID:       "osmosis-1",
		Type:          "packet",
		Targets: ClearingTargets{
			Packets: []PacketIdentifier{
				{Chain: "osmosis-1", Channel: "channel-0", Sequence: 123},
			},
		},
	}

	_, err := service.RequestToken(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

// Test Redis error handling
func TestRedisErrorHandling(t *testing.T) {
	service, _, redisClient := setupTestService(t)
	ctx := context.Background()

	// Close Redis to simulate connection error
	redisClient.Close()

	// Create operation in DB first
	op := &ClearingOperation{
		Token:  "test-token",
		Status: "pending",
	}

	// Cache operations should fail gracefully
	err := service.cache.Set(ctx, op)
	assert.Error(t, err)

	// Get should return error
	_, err = service.cache.Get(ctx, "test-token")
	assert.Error(t, err)
}

// Test timeout handling
func TestTimeoutHandling(t *testing.T) {
	service, _, _ := setupTestService(t)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Let context expire
	time.Sleep(10 * time.Millisecond)

	req := &ClearingRequest{
		WalletAddress: "osmo1test123",
		ChainID:       "osmosis-1",
		Type:          "packet",
	}

	_, err := service.RequestToken(ctx, req)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

// Test payment validation errors
func TestPaymentValidationErrors(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		setupOp       func() *ClearingOperation
		mockBehavior  func(*MockPaymentValidator)
		expectedError string
	}{
		{
			name: "invalid amount",
			setupOp: func() *ClearingOperation {
				return &ClearingOperation{
					Token:         "test-token",
					Status:        "pending_payment",
					PaymentAmount: "invalid",
					ExpiresAt:     time.Now().Add(5 * time.Minute),
				}
			},
			mockBehavior: func(m *MockPaymentValidator) {
				m.On("ValidatePayment", ctx, mock.Anything).
					Return(errors.New("invalid payment amount"))
			},
			expectedError: "invalid payment amount",
		},
		{
			name: "insufficient amount",
			setupOp: func() *ClearingOperation {
				return &ClearingOperation{
					Token:         "test-token",
					Status:        "pending_payment",
					PaymentAmount: "1000000",
					ExpiresAt:     time.Now().Add(5 * time.Minute),
				}
			},
			mockBehavior: func(m *MockPaymentValidator) {
				m.On("ValidatePayment", ctx, mock.Anything).
					Return(errors.New("insufficient payment amount"))
			},
			expectedError: "insufficient payment amount",
		},
		{
			name: "wrong memo",
			setupOp: func() *ClearingOperation {
				return &ClearingOperation{
					Token:         "test-token",
					Status:        "pending_payment",
					PaymentAmount: "1000000",
					PaymentMemo:   "expected-memo",
					ExpiresAt:     time.Now().Add(5 * time.Minute),
				}
			},
			mockBehavior: func(m *MockPaymentValidator) {
				m.On("ValidatePayment", ctx, mock.Anything).
					Return(errors.New("payment memo mismatch"))
			},
			expectedError: "payment memo mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			op := tt.setupOp()
			err := db.Create(op).Error
			assert.NoError(t, err)

			// Mock behavior
			mockValidator := service.paymentValidator.(*MockPaymentValidator)
			tt.mockBehavior(mockValidator)

			// Mock duplicate detector
			mockDetector := service.duplicateDetector.(*MockDuplicateDetector)
			mockDetector.On("IsDuplicate", ctx, mock.Anything).Return(false, nil).Once()

			// Test
			err = service.VerifyPayment(ctx, op.Token, "tx123")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Cleanup
			db.Delete(op)
			mockValidator.ExpectedCalls = nil
			mockDetector.ExpectedCalls = nil
		})
	}
}

// Test execution errors with refund
func TestExecutionErrorsWithRefund(t *testing.T) {
	execService, mockRelayer := setupTestExecutionService(t)
	ctx := context.Background()

	op := &ClearingOperation{
		ID:    "test-op",
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

	payment := &PaymentRecord{
		TxHash:        "payment-tx-123",
		Amount:        "1000000",
		WalletAddress: "osmo1sender",
	}

	// Mock clearing failure after max retries
	mockRelayer.On("ClearPacket", ctx, "osmosis-1", "channel-0", uint64(123)).
		Return(errors.New("permanent failure")).Times(3)

	// Mock refund service
	mockRefund := execService.refundService.(*MockRefundService)
	mockRefund.On("ProcessRefund", ctx, payment, mock.MatchedBy(func(reason string) bool {
		return strings.Contains(reason, "clearing failed")
	})).Return(nil)

	// Mock tracker
	tracker := execService.tracker.(*MockOperationTracker)
	tracker.On("StartOperation", mock.Anything).Return()

	// Execute with payment for refund
	result, err := execService.ExecutePacketClearingWithRefund(ctx, op, payment)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FailedPackets, 1)

	// Verify refund was called
	mockRefund.AssertCalled(t, "ProcessRefund", ctx, payment, mock.Anything)
}

// Test concurrent error scenarios
func TestConcurrentErrorHandling(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()

	// Create multiple operations that will conflict
	const numOps = 5
	token := "conflict-token"

	errors := make(chan error, numOps)

	for i := 0; i < numOps; i++ {
		go func() {
			op := &ClearingOperation{
				Token:         token,
				WalletAddress: "osmo1test123",
				Status:        "pending",
				ExpiresAt:     time.Now().Add(5 * time.Minute),
			}
			err := db.Create(op).Error
			errors <- err
		}()
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numOps; i++ {
		err := <-errors
		if err == nil {
			successCount++
		} else {
			errorCount++
			// Should be unique constraint violation
			assert.Contains(t, err.Error(), "UNIQUE") // SQLite error
		}
	}

	// Only one should succeed due to unique token constraint
	assert.Equal(t, 1, successCount)
	assert.Equal(t, numOps-1, errorCount)
}

// Test panic recovery
func TestPanicRecovery(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a service method that panics
	panicService := &ServiceWithPanicRecovery{
		logger: logger,
	}

	// Should recover from panic and return error
	err := panicService.MethodThatMightPanic()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
}

// ServiceWithPanicRecovery demonstrates panic recovery
type ServiceWithPanicRecovery struct {
	logger *zap.Logger
}

func (s *ServiceWithPanicRecovery) MethodThatMightPanic() (err error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic recovered", zap.Any("panic", r))
			err = errors.New("panic recovered: operation failed")
		}
	}()

	// Simulate panic
	panic("unexpected error")
}

// Test error aggregation
func TestErrorAggregation(t *testing.T) {
	var errs ErrorList

	errs.Add(errors.New("error 1"))
	errs.Add(errors.New("error 2"))
	errs.Add(nil) // Should be ignored
	errs.Add(errors.New("error 3"))

	assert.Len(t, errs, 3)
	assert.Contains(t, errs.Error(), "error 1")
	assert.Contains(t, errs.Error(), "error 2")
	assert.Contains(t, errs.Error(), "error 3")
}

// ErrorList aggregates multiple errors
type ErrorList []error

func (e *ErrorList) Add(err error) {
	if err != nil {
		*e = append(*e, err)
	}
}

func (e ErrorList) Error() string {
	if len(e) == 0 {
		return ""
	}
	
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	
	return "multiple errors: " + strings.Join(messages, "; ")
}

// Test graceful degradation
func TestGracefulDegradation(t *testing.T) {
	service, _, _ := setupTestService(t)
	ctx := context.Background()

	// Simulate various component failures
	tests := []struct {
		name             string
		failureScenario  func()
		operation        func() error
		shouldSucceed    bool
		expectedBehavior string
	}{
		{
			name: "cache failure - should continue",
			failureScenario: func() {
				// Cache is unavailable
				service.cache = nil
			},
			operation: func() error {
				req := &ClearingRequest{
					WalletAddress: "osmo1test123",
					ChainID:       "osmosis-1",
					Type:          "packet",
				}
				_, err := service.RequestToken(ctx, req)
				return err
			},
			shouldSucceed:    true,
			expectedBehavior: "continues without cache",
		},
		{
			name: "metrics failure - should continue",
			failureScenario: func() {
				// Metrics collector is nil
				service.metrics = nil
			},
			operation: func() error {
				service.recordMetric("test_metric", 1)
				return nil
			},
			shouldSucceed:    true,
			expectedBehavior: "continues without metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply failure scenario
			tt.failureScenario()

			// Execute operation
			err := tt.operation()

			if tt.shouldSucceed {
				assert.NoError(t, err, tt.expectedBehavior)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Helper method for metrics recording
func (s *ServiceV2) recordMetric(name string, value float64) {
	// Gracefully handle nil metrics
	if s.metrics == nil {
		return
	}
	
	// Record metric
	s.metrics.Record(name, value)
}

