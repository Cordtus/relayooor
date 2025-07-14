package clearing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	TokenTTL             = 5 * time.Minute
	SessionTTL           = 24 * time.Hour
	PaymentGracePeriod   = 2 * time.Minute
	DefaultServiceFee    = "1000000"  // 1 TOKEN in smallest unit
	DefaultPerPacketFee  = "100000"   // 0.1 TOKEN per packet
	DefaultGasPrice      = "25000"    // 0.025 TOKEN
	BaseGasAmount        = 150000     // Base gas for IBC tx
	PerPacketGas         = 50000      // Additional gas per packet
)

type Service struct {
	redisClient      *redis.Client
	secretKey        string
	serviceAddress   string
	chainRPCs        map[string]string
	hermesURL        string
}

// NewService creates a new clearing service instance
func NewService(redisClient *redis.Client) *Service {
	secretKey := os.Getenv("CLEARING_SECRET_KEY")
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
	}

	serviceAddress := os.Getenv("SERVICE_WALLET_ADDRESS")
	if serviceAddress == "" {
		serviceAddress = "cosmos1service..."
	}

	hermesURL := os.Getenv("HERMES_REST_URL")
	if hermesURL == "" {
		hermesURL = "http://localhost:5185"
	}

	// Parse chain RPCs from environment
	chainRPCs := make(map[string]string)
	rpcConfig := os.Getenv("CHAIN_RPC_ENDPOINTS")
	if rpcConfig != "" {
		// Format: "osmosis:https://rpc.osmosis.zone,cosmoshub:https://rpc.cosmos.network"
		// Parse and populate chainRPCs map
	}

	return &Service{
		redisClient:    redisClient,
		secretKey:      secretKey,
		serviceAddress: serviceAddress,
		chainRPCs:      chainRPCs,
		hermesURL:      hermesURL,
	}
}

// GenerateToken creates a new clearing token
func (s *Service) GenerateToken(ctx context.Context, request ClearingRequest) (*TokenResponse, error) {
	// Validate request
	if err := s.validateClearingRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Calculate fees
	fees, err := s.calculateFees(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fees: %w", err)
	}

	// Generate token
	token := ClearingToken{
		Token:             uuid.New().String(),
		Version:           1,
		RequestType:       request.Type,
		TargetIdentifiers: request.Targets,
		WalletAddress:     request.WalletAddress,
		ChainID:           request.ChainID,
		IssuedAt:          time.Now().Unix(),
		ExpiresAt:         time.Now().Add(TokenTTL).Unix(),
		ServiceFee:        fees.ServiceFee,
		EstimatedGasFee:   fees.GasFee,
		TotalRequired:     fees.Total,
		AcceptedDenom:     fees.Denom,
		Nonce:             generateNonce(),
	}

	// Sign token
	token.Signature = s.signToken(token)

	// Store in Redis
	tokenKey := fmt.Sprintf("clearing:token:%s", token.Token)
	tokenData, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := s.redisClient.Set(ctx, tokenKey, tokenData, TokenTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	// Generate payment memo
	memo := s.generatePaymentMemo(token.Token, request.Type, request.Targets)

	return &TokenResponse{
		Token:          token,
		PaymentAddress: s.serviceAddress,
		Memo:           memo,
	}, nil
}

// VerifyPayment verifies a payment transaction
func (s *Service) VerifyPayment(ctx context.Context, tokenID string, txHash string) (*PaymentVerificationResponse, error) {
	// Get token from Redis
	token, err := s.getToken(ctx, tokenID)
	if err != nil {
		return &PaymentVerificationResponse{
			Verified: false,
			Status:   "invalid",
			Message:  "Token not found or expired",
		}, nil
	}

	// Check if already paid
	paymentKey := fmt.Sprintf("clearing:payment:%s", tokenID)
	exists, err := s.redisClient.Exists(ctx, paymentKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check payment status: %w", err)
	}
	if exists > 0 {
		return &PaymentVerificationResponse{
			Verified: true,
			Status:   "verified",
			Message:  "Payment already verified",
		}, nil
	}

	// Verify transaction on chain
	verified, amount, err := s.verifyTransactionOnChain(ctx, token.ChainID, txHash, token)
	if err != nil {
		log.Printf("Failed to verify transaction %s: %v", txHash, err)
		return &PaymentVerificationResponse{
			Verified: false,
			Status:   "pending",
			Message:  "Transaction verification pending",
		}, nil
	}

	if !verified {
		return &PaymentVerificationResponse{
			Verified: false,
			Status:   "invalid",
			Message:  "Invalid payment transaction",
		}, nil
	}

	// Check amount
	requiredAmount, _ := new(big.Int).SetString(token.TotalRequired, 10)
	paidAmount, _ := new(big.Int).SetString(amount, 10)
	
	if paidAmount.Cmp(requiredAmount) < 0 {
		return &PaymentVerificationResponse{
			Verified: false,
			Status:   "insufficient",
			Message:  fmt.Sprintf("Insufficient payment: required %s, received %s", token.TotalRequired, amount),
		}, nil
	}

	// Mark as paid
	paymentData := map[string]interface{}{
		"txHash":    txHash,
		"amount":    amount,
		"timestamp": time.Now().Unix(),
	}
	paymentJSON, _ := json.Marshal(paymentData)
	
	// Extend token expiry by grace period for execution
	extendedTTL := TokenTTL + PaymentGracePeriod
	if err := s.redisClient.Set(ctx, paymentKey, paymentJSON, extendedTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to store payment data: %w", err)
	}

	// Queue for execution
	if err := s.queueForExecution(ctx, tokenID); err != nil {
		log.Printf("Failed to queue token %s for execution: %v", tokenID, err)
	}

	return &PaymentVerificationResponse{
		Verified: true,
		Status:   "verified",
		Message:  "Payment verified successfully",
	}, nil
}

// GetStatus returns the current status of a clearing operation
func (s *Service) GetStatus(ctx context.Context, tokenID string) (*ClearingStatus, error) {
	// Get token
	token, err := s.getToken(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("token not found")
	}

	// Check payment status
	paymentKey := fmt.Sprintf("clearing:payment:%s", tokenID)
	paymentData, err := s.redisClient.Get(ctx, paymentKey).Result()
	paymentStatus := PaymentStatus{Received: false}
	
	if err == nil {
		var payment map[string]interface{}
		if err := json.Unmarshal([]byte(paymentData), &payment); err == nil {
			paymentStatus.Received = true
			paymentStatus.TxHash = payment["txHash"].(string)
			paymentStatus.Amount = payment["amount"].(string)
		}
	}

	// Check execution status
	executionKey := fmt.Sprintf("clearing:execution:%s", tokenID)
	executionData, err := s.redisClient.Get(ctx, executionKey).Result()
	
	var execution *ExecutionInfo
	var status string = "pending"
	
	if err == nil {
		execution = &ExecutionInfo{}
		if err := json.Unmarshal([]byte(executionData), execution); err == nil {
			if execution.CompletedAt != nil {
				if execution.Error != "" {
					status = "failed"
				} else {
					status = "completed"
				}
			} else {
				status = "executing"
			}
		}
	} else if paymentStatus.Received {
		status = "paid"
	}

	return &ClearingStatus{
		Token:     tokenID,
		Status:    status,
		Payment:   paymentStatus,
		Execution: execution,
	}, nil
}

// Helper functions

func (s *Service) validateClearingRequest(request ClearingRequest) error {
	switch request.Type {
	case "packet":
		if len(request.Targets.Packets) == 0 {
			return errors.New("no packets specified")
		}
	case "channel":
		if len(request.Targets.Channels) == 0 {
			return errors.New("no channels specified")
		}
	case "bulk":
		if len(request.Targets.Channels) == 0 {
			return errors.New("no channels specified for bulk operation")
		}
	default:
		return errors.New("invalid request type")
	}
	return nil
}

type FeeCalculation struct {
	ServiceFee string
	GasFee     string
	Total      string
	Denom      string
}

func (s *Service) calculateFees(ctx context.Context, request ClearingRequest) (*FeeCalculation, error) {
	serviceFeeStr := os.Getenv("CLEARING_SERVICE_FEE")
	if serviceFeeStr == "" {
		serviceFeeStr = DefaultServiceFee
	}
	
	perPacketFeeStr := os.Getenv("CLEARING_PER_PACKET_FEE")
	if perPacketFeeStr == "" {
		perPacketFeeStr = DefaultPerPacketFee
	}

	serviceFee, _ := new(big.Int).SetString(serviceFeeStr, 10)
	perPacketFee, _ := new(big.Int).SetString(perPacketFeeStr, 10)

	// Calculate packet count
	packetCount := 0
	switch request.Type {
	case "packet":
		packetCount = len(request.Targets.Packets)
	case "channel":
		// Estimate packets in channel (would query actual count in production)
		packetCount = 10 // placeholder
	case "bulk":
		// Estimate total packets across channels
		packetCount = len(request.Targets.Channels) * 10 // placeholder
	}

	// Calculate total service fee
	totalServiceFee := new(big.Int).Mul(perPacketFee, big.NewInt(int64(packetCount)))
	totalServiceFee.Add(totalServiceFee, serviceFee)

	// Calculate gas fee
	totalGas := BaseGasAmount + (packetCount * PerPacketGas)
	gasPrice, _ := new(big.Int).SetString(DefaultGasPrice, 10)
	gasFee := new(big.Int).Mul(big.NewInt(int64(totalGas)), gasPrice)

	// Total required
	total := new(big.Int).Add(totalServiceFee, gasFee)

	return &FeeCalculation{
		ServiceFee: totalServiceFee.String(),
		GasFee:     gasFee.String(),
		Total:      total.String(),
		Denom:      "utoken", // Would be dynamic based on chain
	}, nil
}

func (s *Service) signToken(token ClearingToken) string {
	// Create signing payload (exclude signature field)
	payload := fmt.Sprintf("%s:%d:%s:%s:%d:%d:%s",
		token.Token,
		token.Version,
		token.WalletAddress,
		token.ChainID,
		token.IssuedAt,
		token.ExpiresAt,
		token.Nonce,
	)

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Service) generatePaymentMemo(tokenID string, action string, targets ClearingTargets) string {
	// Create abbreviated token (first 8 and last 8 chars)
	abbrevToken := tokenID
	if len(tokenID) > 16 {
		abbrevToken = tokenID[:8] + "..." + tokenID[len(tokenID)-8:]
	}

	memoData := PaymentMemo{
		Version: 1,
		Token:   abbrevToken,
		Action:  "clear_" + action,
		Data:    make(map[string]interface{}),
	}

	// Add target data
	switch action {
	case "packet":
		sequences := make([]uint64, len(targets.Packets))
		for i, p := range targets.Packets {
			sequences[i] = p.Sequence
		}
		memoData.Data["p"] = sequences
	case "channel":
		if len(targets.Channels) > 0 {
			memoData.Data["c"] = fmt.Sprintf("%s->%s", 
				targets.Channels[0].SrcChannel, 
				targets.Channels[0].DstChannel)
		}
	}

	// Marshal to compact JSON
	memoJSON, _ := json.Marshal(memoData)
	return string(memoJSON)
}

func (s *Service) getToken(ctx context.Context, tokenID string) (*ClearingToken, error) {
	tokenKey := fmt.Sprintf("clearing:token:%s", tokenID)
	tokenData, err := s.redisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		return nil, err
	}

	var token ClearingToken
	if err := json.Unmarshal([]byte(tokenData), &token); err != nil {
		return nil, err
	}

	// Verify token hasn't expired
	if time.Now().Unix() > token.ExpiresAt {
		return nil, errors.New("token expired")
	}

	return &token, nil
}

func (s *Service) verifyTransactionOnChain(ctx context.Context, chainID string, txHash string, token *ClearingToken) (bool, string, error) {
	// This would implement actual chain verification
	// For now, return mock success
	return true, token.TotalRequired, nil
}

func (s *Service) queueForExecution(ctx context.Context, tokenID string) error {
	// Add to execution queue
	return s.redisClient.LPush(ctx, "clearing:execution:queue", tokenID).Err()
}

func generateNonce() string {
	return uuid.New().String()[:8]
}

// GetServiceAddress returns the service wallet address
func (s *Service) GetServiceAddress() string {
	return s.serviceAddress
}