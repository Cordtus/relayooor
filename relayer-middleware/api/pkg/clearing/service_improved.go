package clearing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrDuplicatePayment = errors.New("duplicate payment detected")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidToken     = errors.New("invalid token")
)

// ServiceV2 is the improved clearing service with error handling
type ServiceV2 struct {
	db                *gorm.DB
	redisClient       *redis.Client
	secretKey         string
	serviceAddress    string
	chainRPCs         map[string]string
	hermesURL         string
	logger            *zap.Logger
	
	// New components
	duplicateDetector *DuplicateDetector
	paymentValidator  *PaymentValidator
	cache             *PacketCache
	refundService     *RefundService
	executionService  *ExecutionServiceV2
}

// Config holds service configuration
type Config struct {
	SecretKey      string
	ServiceAddress string
	HermesURL      string
	ChainRPCs      map[string]string
}

// NewServiceV2 creates a new improved clearing service
func NewServiceV2(
	db *gorm.DB,
	redisClient *redis.Client,
	config Config,
	logger *zap.Logger,
) *ServiceV2 {
	// Create service wallet
	serviceWallet := ServiceWallet{
		Address: config.ServiceAddress,
		// Private key should be loaded from secure storage
	}
	
	// Initialize components
	duplicateDetector := NewDuplicateDetector(redisClient, db, logger)
	paymentValidator := NewPaymentValidator(config.ServiceAddress)
	cache := NewPacketCache(redisClient, logger)
	refundService := NewRefundService(db, serviceWallet, logger)
	
	service := &ServiceV2{
		db:                db,
		redisClient:       redisClient,
		secretKey:         config.SecretKey,
		serviceAddress:    config.ServiceAddress,
		chainRPCs:         config.ChainRPCs,
		hermesURL:         config.HermesURL,
		logger:            logger.With(zap.String("component", "clearing_service")),
		duplicateDetector: duplicateDetector,
		paymentValidator:  paymentValidator,
		cache:             cache,
		refundService:     refundService,
	}
	
	// Create execution service
	// TODO: Create actual Hermes client
	var hermesClient HermesClient
	tracker := NewOperationTracker()
	
	service.executionService = NewExecutionServiceV2(
		db,
		redisClient,
		hermesClient,
		refundService,
		tracker,
		logger,
	)
	
	return service
}

// Start initializes background workers
func (s *ServiceV2) Start(ctx context.Context) {
	// Start execution service
	s.executionService.Start(ctx)
	
	// Start refund processor
	go s.refundService.ProcessPendingRefunds(ctx)
	
	// Start duplicate detector cleanup
	go s.duplicateDetector.CleanupOldRecords(ctx)
	
	s.logger.Info("Clearing service started")
}

// GenerateToken creates a new clearing token
func (s *ServiceV2) GenerateToken(ctx context.Context, request ClearingRequest) (*TokenResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "generate_token"),
		zap.String("wallet", request.WalletAddress),
		zap.Int("packet_count", len(request.PacketIdentifiers)),
	)
	
	logger.Info("Generating clearing token")
	
	// Validate request
	if err := s.validateRequest(request); err != nil {
		logger.Error("Request validation failed", zap.Error(err))
		return nil, err
	}
	
	// Calculate fees
	totalPackets := len(request.PacketIdentifiers)
	serviceFee := s.getServiceFee()
	perPacketFee := s.getPerPacketFee()
	totalServiceFee := serviceFee + (perPacketFee * int64(totalPackets))
	
	// Estimate gas
	estimatedGas := s.estimateGas(totalPackets)
	gasPrice := s.getGasPrice(request.ChainID)
	estimatedGasFee := estimatedGas * gasPrice
	
	totalRequired := totalServiceFee + estimatedGasFee
	
	// Create token
	token := &ClearingToken{
		Token:             uuid.New().String(),
		Version:           1,
		RequestType:       "clear_packets",
		TargetIdentifiers: request.TargetIdentifiers,
		WalletAddress:     request.WalletAddress,
		ChainID:           request.ChainID,
		IssuedAt:          time.Now().Unix(),
		ExpiresAt:         time.Now().Add(TokenTTL).Unix(),
		ServiceFee:        fmt.Sprintf("%d", totalServiceFee),
		EstimatedGasFee:   fmt.Sprintf("%d", estimatedGasFee),
		TotalRequired:     fmt.Sprintf("%d", totalRequired),
		AcceptedDenom:     s.getAcceptedDenom(request.ChainID),
		Nonce:             generateNonce(),
	}
	
	// Sign token
	token.Signature = s.signToken(token)
	
	// Store token in Redis
	tokenKey := fmt.Sprintf("token:%s", token.Token)
	tokenData, err := json.Marshal(token)
	if err != nil {
		logger.Error("Failed to marshal token", zap.Error(err))
		return nil, err
	}
	
	if err := s.redisClient.Set(ctx, tokenKey, tokenData, TokenTTL).Err(); err != nil {
		logger.Error("Failed to store token", zap.Error(err))
		return nil, err
	}
	
	// Store packet identifiers
	packetsKey := fmt.Sprintf("packets:%s", token.Token)
	packetsData, err := json.Marshal(request.PacketIdentifiers)
	if err != nil {
		logger.Error("Failed to marshal packets", zap.Error(err))
		return nil, err
	}
	
	if err := s.redisClient.Set(ctx, packetsKey, packetsData, TokenTTL).Err(); err != nil {
		logger.Error("Failed to store packets", zap.Error(err))
		return nil, err
	}
	
	logger.Info("Token generated successfully",
		zap.String("token_id", token.Token),
		zap.Time("expires_at", time.Unix(token.ExpiresAt, 0)),
	)
	
	// Invalidate user's packet cache
	s.cache.InvalidateUserPackets(ctx, request.WalletAddress)
	
	return &TokenResponse{
		Token:         token,
		PaymentMemo:   generatePaymentMemo(token.Token),
		PaymentAmount: token.TotalRequired,
		ExpiresIn:     int(TokenTTL.Seconds()),
	}, nil
}

// VerifyPayment verifies the payment transaction and queues the clearing
func (s *ServiceV2) VerifyPayment(ctx context.Context, tokenID string, txHash string) (*PaymentVerificationResponse, error) {
	logger := s.logger.With(
		zap.String("token_id", tokenID),
		zap.String("tx_hash", txHash),
	)
	
	// Check for duplicate payment
	isDuplicate, err := s.duplicateDetector.CheckDuplicate(ctx, txHash)
	if err != nil {
		logger.Error("Failed to check duplicate", zap.Error(err))
		return nil, err
	}
	
	if isDuplicate {
		// Get previous payment info
		prevInfo, err := s.duplicateDetector.GetPaymentInfo(ctx, txHash)
		if err != nil {
			logger.Error("Failed to get duplicate payment info", zap.Error(err))
		}
		
		if prevInfo != nil && prevInfo.TokenID == tokenID {
			// Same token, return success (idempotent)
			return &PaymentVerificationResponse{
				Success:     true,
				OperationID: prevInfo.OperationID,
				Message:     "Payment already processed",
			}, nil
		}
		
		return nil, ErrDuplicatePayment
	}
	
	// Get token from Redis
	tokenKey := fmt.Sprintf("token:%s", tokenID)
	tokenData, err := s.redisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTokenExpired
		}
		logger.Error("Failed to get token", zap.Error(err))
		return nil, err
	}
	
	var token ClearingToken
	if err := json.Unmarshal([]byte(tokenData), &token); err != nil {
		logger.Error("Failed to unmarshal token", zap.Error(err))
		return nil, err
	}
	
	// Verify token hasn't expired
	if time.Now().Unix() > token.ExpiresAt {
		return nil, ErrTokenExpired
	}
	
	// Get transaction details
	tx, err := s.getTransaction(ctx, token.ChainID, txHash)
	if err != nil {
		logger.Error("Failed to get transaction", zap.Error(err))
		return nil, err
	}
	
	// Validate payment
	if err := s.paymentValidator.ValidatePayment(ctx, &token, tx); err != nil {
		logger.Error("Payment validation failed", zap.Error(err))
		
		// Check if overpayment
		if IsOverpayment(err) {
			// Process payment but mark for partial refund
			overpayment := err.(*ErrOverpayment)
			logger.Info("Overpayment detected, will process partial refund",
				zap.String("paid", overpayment.Paid),
				zap.String("required", overpayment.Required),
			)
		} else {
			return nil, err
		}
	}
	
	// Create clearing operation
	operationID := uuid.New().String()
	operation := &ClearingOperation{
		ID:               operationID,
		TokenID:          tokenID,
		WalletAddress:    token.WalletAddress,
		ChainID:          token.ChainID,
		PaymentTxHash:    txHash,
		PaymentAddress:   tx.FromAddress,
		ServiceFee:       token.ServiceFee,
		EstimatedGasFee:  token.EstimatedGasFee,
		ActualFeePaid:    tx.Amount,
		FeeDenom:         token.AcceptedDenom,
		Status:           "queued",
		CreatedAt:        time.Now(),
	}
	
	if err := s.db.Create(operation).Error; err != nil {
		logger.Error("Failed to create operation", zap.Error(err))
		return nil, err
	}
	
	// Store payment info
	paymentInfo := &PaymentInfo{
		TxHash:        txHash,
		TokenID:       tokenID,
		OperationID:   operationID,
		WalletAddress: token.WalletAddress,
		Amount:        tx.Amount,
		Denom:         token.AcceptedDenom,
		ProcessedAt:   time.Now(),
	}
	
	if err := s.duplicateDetector.StorePaymentInfo(ctx, paymentInfo); err != nil {
		logger.Error("Failed to store payment info", zap.Error(err))
	}
	
	// Queue for execution
	if err := s.redisClient.RPush(ctx, "clearing:execution:queue", tokenID).Err(); err != nil {
		logger.Error("Failed to queue for execution", zap.Error(err))
		return nil, err
	}
	
	// Mark token as used
	s.redisClient.Del(ctx, tokenKey)
	
	logger.Info("Payment verified and clearing queued",
		zap.String("operation_id", operationID),
		zap.String("amount", tx.Amount),
	)
	
	return &PaymentVerificationResponse{
		Success:     true,
		OperationID: operationID,
		Message:     "Payment verified, clearing in progress",
	}, nil
}

// Helper methods

func (s *ServiceV2) validateRequest(request ClearingRequest) error {
	if request.WalletAddress == "" {
		return errors.New("wallet address required")
	}
	
	if len(request.PacketIdentifiers) == 0 {
		return errors.New("no packets to clear")
	}
	
	if len(request.PacketIdentifiers) > 100 {
		return errors.New("too many packets (max 100)")
	}
	
	return nil
}

func (s *ServiceV2) signToken(token *ClearingToken) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	
	// Create signing payload
	payload := fmt.Sprintf("%s:%d:%s:%s:%s",
		token.Token,
		token.IssuedAt,
		token.WalletAddress,
		token.TotalRequired,
		token.Nonce,
	)
	
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *ServiceV2) getServiceFee() int64 {
	// TODO: Make configurable
	return 1000000 // 1 TOKEN
}

func (s *ServiceV2) getPerPacketFee() int64 {
	// TODO: Make configurable
	return 100000 // 0.1 TOKEN
}

func (s *ServiceV2) estimateGas(packetCount int) int64 {
	return int64(BaseGasAmount + (PerPacketGas * packetCount))
}

func (s *ServiceV2) getGasPrice(chainID string) int64 {
	// TODO: Get from chain
	return 25000 // 0.025 TOKEN
}

func (s *ServiceV2) getAcceptedDenom(chainID string) string {
	// TODO: Make configurable per chain
	denoms := map[string]string{
		"osmosis-1":    "uosmo",
		"cosmoshub-4":  "uatom",
		"neutron-1":    "untrn",
	}
	
	if denom, ok := denoms[chainID]; ok {
		return denom
	}
	
	return "uatom" // Default
}

func (s *ServiceV2) getTransaction(ctx context.Context, chainID, txHash string) (*Transaction, error) {
	// TODO: Implement actual transaction fetching
	return &Transaction{
		Hash:        txHash,
		FromAddress: "cosmos1...",
		Amount:      "1100000",
		Messages:    []Message{},
	}, nil
}

func generateNonce() string {
	return uuid.New().String()[:8]
}

func generatePaymentMemo(tokenID string) string {
	return fmt.Sprintf("CLR-%s", tokenID)
}