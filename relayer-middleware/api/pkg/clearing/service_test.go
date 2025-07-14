package clearing

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock dependencies
type MockPaymentValidator struct {
	mock.Mock
}

func (m *MockPaymentValidator) ValidatePayment(ctx context.Context, payment *PaymentRecord) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

type MockDuplicateDetector struct {
	mock.Mock
}

func (m *MockDuplicateDetector) IsDuplicate(ctx context.Context, txHash string) (bool, error) {
	args := m.Called(ctx, txHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockDuplicateDetector) Add(ctx context.Context, txHash string) error {
	args := m.Called(ctx, txHash)
	return args.Error(0)
}

type MockRefundService struct {
	mock.Mock
}

func (m *MockRefundService) ProcessRefund(ctx context.Context, payment *PaymentRecord, reason string) error {
	args := m.Called(ctx, payment, reason)
	return args.Error(0)
}

// Test setup
func setupTestService(t *testing.T) (*ServiceV2, *gorm.DB, *redis.Client) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Run migrations
	err = db.AutoMigrate(&ClearingOperation{}, &PaymentRecord{})
	require.NoError(t, err)
	
	// Setup Redis (using miniredis for testing)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	// Create logger
	logger, _ := zap.NewDevelopment()
	
	// Create service with mocks
	service := &ServiceV2{
		db:                db,
		redisClient:       redisClient,
		duplicateDetector: &MockDuplicateDetector{},
		paymentValidator:  &MockPaymentValidator{},
		cache:             NewPacketCache(redisClient, logger),
		refundService:     &MockRefundService{},
		executionService:  nil, // Will be mocked separately
	}
	
	return service, db, redisClient
}

// Test RequestToken
func TestRequestToken(t *testing.T) {
	service, _, _ := setupTestService(t)
	ctx := context.Background()
	
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
	
	resp, err := service.RequestToken(ctx, req)
	
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.PaymentAddress)
	assert.NotEmpty(t, resp.PaymentMemo)
	assert.Greater(t, resp.PaymentAmount, "0")
	assert.Greater(t, resp.ExpiresIn, int64(0))
}

// Test RequestToken with invalid request
func TestRequestTokenInvalidRequest(t *testing.T) {
	service, _, _ := setupTestService(t)
	ctx := context.Background()
	
	// Missing wallet address
	req := &ClearingRequest{
		ChainID: "osmosis-1",
		Type:    "packet",
	}
	
	resp, err := service.RequestToken(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Test VerifyPayment
func TestVerifyPayment(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()
	
	// Create a clearing operation
	op := &ClearingOperation{
		Token:          "test-token",
		WalletAddress:  "osmo1test123",
		ChainID:        "osmosis-1",
		Type:           "packet",
		Status:         "pending_payment",
		PaymentAmount:  "1000000",
		PaymentAddress: "osmo1payment",
		PaymentMemo:    "test-memo",
		ExpiresAt:      time.Now().Add(5 * time.Minute),
	}
	err := db.Create(op).Error
	require.NoError(t, err)
	
	// Mock duplicate detector
	mockDetector := service.duplicateDetector.(*MockDuplicateDetector)
	mockDetector.On("IsDuplicate", ctx, "tx123").Return(false, nil)
	mockDetector.On("Add", ctx, "tx123").Return(nil)
	
	// Mock payment validator
	mockValidator := service.paymentValidator.(*MockPaymentValidator)
	mockValidator.On("ValidatePayment", ctx, mock.Anything).Return(nil)
	
	// Verify payment
	err = service.VerifyPayment(ctx, "test-token", "tx123")
	
	assert.NoError(t, err)
	
	// Check operation status updated
	var updatedOp ClearingOperation
	err = db.First(&updatedOp, "token = ?", "test-token").Error
	require.NoError(t, err)
	assert.Equal(t, "payment_verified", updatedOp.Status)
	
	mockDetector.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

// Test VerifyPayment with duplicate
func TestVerifyPaymentDuplicate(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()
	
	// Create operation
	op := &ClearingOperation{
		Token:         "test-token",
		Status:        "pending_payment",
		PaymentAmount: "1000000",
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}
	err := db.Create(op).Error
	require.NoError(t, err)
	
	// Mock duplicate detector to return true
	mockDetector := service.duplicateDetector.(*MockDuplicateDetector)
	mockDetector.On("IsDuplicate", ctx, "tx123").Return(true, nil)
	
	// Verify payment
	err = service.VerifyPayment(ctx, "test-token", "tx123")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate payment")
	
	mockDetector.AssertExpectations(t)
}

// Test GetStatus
func TestGetStatus(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()
	
	// Create operation with execution details
	op := &ClearingOperation{
		Token:         "test-token",
		WalletAddress: "osmo1test123",
		Status:        "completed",
		ExecutionDetails: map[string]interface{}{
			"packetsCleared": 2,
			"txHashes":       []string{"tx1", "tx2"},
		},
	}
	err := db.Create(op).Error
	require.NoError(t, err)
	
	// Get status
	status, err := service.GetStatus(ctx, "test-token")
	
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "test-token", status.Token)
	assert.Equal(t, "completed", status.Status)
	assert.NotNil(t, status.Execution)
}

// Test GetStatus with non-existent token
func TestGetStatusNotFound(t *testing.T) {
	service, _, _ := setupTestService(t)
	ctx := context.Background()
	
	status, err := service.GetStatus(ctx, "non-existent-token")
	
	assert.Error(t, err)
	assert.Nil(t, status)
}

// Test calculateServiceFee
func TestCalculateServiceFee(t *testing.T) {
	service, _, _ := setupTestService(t)
	
	tests := []struct {
		name     string
		req      *ClearingRequest
		expected string
	}{
		{
			name: "single packet",
			req: &ClearingRequest{
				Type: "packet",
				Targets: ClearingTargets{
					Packets: []PacketIdentifier{
						{Chain: "osmosis-1", Channel: "channel-0", Sequence: 123},
					},
				},
			},
			expected: "110000", // BaseGasAmount + PerPacketGas
		},
		{
			name: "multiple packets",
			req: &ClearingRequest{
				Type: "packet",
				Targets: ClearingTargets{
					Packets: []PacketIdentifier{
						{Chain: "osmosis-1", Channel: "channel-0", Sequence: 123},
						{Chain: "osmosis-1", Channel: "channel-0", Sequence: 124},
						{Chain: "osmosis-1", Channel: "channel-0", Sequence: 125},
					},
				},
			},
			expected: "130000", // BaseGasAmount + (3 * PerPacketGas)
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee := service.calculateServiceFee(tt.req)
			assert.Equal(t, tt.expected, fee)
		})
	}
}

// Test concurrent token requests
func TestConcurrentTokenRequests(t *testing.T) {
	service, _, _ := setupTestService(t)
	ctx := context.Background()
	
	// Run multiple concurrent requests
	const numRequests = 10
	errors := make(chan error, numRequests)
	tokens := make(chan string, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			req := &ClearingRequest{
				WalletAddress: "osmo1test123",
				ChainID:       "osmosis-1",
				Type:          "packet",
				Targets: ClearingTargets{
					Packets: []PacketIdentifier{
						{Chain: "osmosis-1", Channel: "channel-0", Sequence: uint64(idx)},
					},
				},
			}
			
			resp, err := service.RequestToken(ctx, req)
			if err != nil {
				errors <- err
			} else {
				tokens <- resp.Token
			}
		}(i)
	}
	
	// Collect results
	uniqueTokens := make(map[string]bool)
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		case token := <-tokens:
			assert.NotEmpty(t, token)
			// Ensure all tokens are unique
			assert.False(t, uniqueTokens[token], "Duplicate token generated")
			uniqueTokens[token] = true
		}
	}
	
	assert.Len(t, uniqueTokens, numRequests)
}

// Test token expiration
func TestTokenExpiration(t *testing.T) {
	service, db, _ := setupTestService(t)
	ctx := context.Background()
	
	// Create expired operation
	op := &ClearingOperation{
		Token:     "expired-token",
		Status:    "pending_payment",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err := db.Create(op).Error
	require.NoError(t, err)
	
	// Try to verify payment on expired token
	err = service.VerifyPayment(ctx, "expired-token", "tx123")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}