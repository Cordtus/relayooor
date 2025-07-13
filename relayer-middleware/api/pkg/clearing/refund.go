package clearing

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RefundService struct {
	db            *gorm.DB
	serviceWallet ServiceWallet
	logger        *zap.Logger
}

type ServiceWallet struct {
	Address    string
	PrivateKey string // This should be encrypted in production
}

type RefundableOperation struct {
	ID            string    `gorm:"primaryKey"`
	OperationID   string    `gorm:"index"`
	WalletAddress string
	RefundAddress string
	ChainID       string
	AmountPaid    string
	Denom         string
	RefundReason  string
	RefundStatus  string `gorm:"index"` // pending, processing, completed, failed, manual_required
	RefundTxHash  string
	ErrorMessage  string
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}

var (
	ErrInsufficientRefundBalance = fmt.Errorf("insufficient balance for refund")
)

func NewRefundService(db *gorm.DB, wallet ServiceWallet, logger *zap.Logger) *RefundService {
	return &RefundService{
		db:            db,
		serviceWallet: wallet,
		logger:        logger.With(zap.String("component", "refund")),
	}
}

func (s *RefundService) ProcessRefund(ctx context.Context, operationID string, reason string) error {
	logger := s.logger.With(
		zap.String("operation_id", operationID),
		zap.String("reason", reason),
	)

	// Get operation details
	var operation ClearingOperation
	if err := s.db.Where("id = ?", operationID).First(&operation).Error; err != nil {
		logger.Error("Failed to find operation", zap.Error(err))
		return fmt.Errorf("operation not found: %w", err)
	}

	// Check if already refunded
	if operation.RefundStatus == "completed" {
		logger.Info("Operation already refunded")
		return nil
	}

	// Create refund record
	refund := RefundableOperation{
		ID:            generateID(),
		OperationID:   operationID,
		WalletAddress: operation.WalletAddress,
		RefundAddress: operation.PaymentAddress, // Refund to payment source
		ChainID:       operation.ChainID,
		AmountPaid:    operation.ActualFeePaid,
		Denom:         operation.FeeDenom,
		RefundReason:  reason,
		RefundStatus:  "processing",
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.db.Create(&refund).Error; err != nil {
		logger.Error("Failed to create refund record", zap.Error(err))
		return err
	}

	// Calculate refund amount (minus network fees)
	refundAmount, err := s.calculateRefundAmount(operation)
	if err != nil {
		logger.Error("Failed to calculate refund amount", zap.Error(err))
		return err
	}

	// Check service wallet balance before refund
	balance, err := s.getServiceWalletBalance(ctx, refund.Denom)
	if err != nil {
		logger.Error("Failed to check service wallet balance", zap.Error(err))
		return err
	}

	if balance.IsLT(refundAmount.Amount) {
		logger.Error("Insufficient balance for refund",
			zap.String("required", refundAmount.String()),
			zap.String("available", balance.String()),
		)

		// Alert operators
		s.alertOperators("Insufficient balance for refund", map[string]interface{}{
			"operation_id": operationID,
			"required":     refundAmount.String(),
			"available":    balance.String(),
		})

		// Mark for manual processing
		s.db.Model(&refund).Updates(map[string]interface{}{
			"refund_status": "manual_required",
			"error_message": "Insufficient service wallet balance",
		})

		return ErrInsufficientRefundBalance
	}

	// Execute refund transaction
	txHash, err := s.executeRefund(ctx, refund, refundAmount)
	if err != nil {
		logger.Error("Failed to execute refund", zap.Error(err))

		// Update refund status to failed
		s.db.Model(&refund).Updates(map[string]interface{}{
			"refund_status": "failed",
			"error_message": err.Error(),
		})

		return err
	}

	// Update refund record
	now := time.Now().UTC()
	if err := s.db.Model(&refund).Updates(map[string]interface{}{
		"refund_status": "completed",
		"refund_tx_hash": txHash,
		"processed_at":   &now,
	}).Error; err != nil {
		logger.Error("Failed to update refund record", zap.Error(err))
		return err
	}

	// Update operation status
	if err := s.db.Model(&operation).Updates(map[string]interface{}{
		"refund_status":  "completed",
		"refund_tx_hash": txHash,
		"refund_reason":  reason,
	}).Error; err != nil {
		logger.Error("Failed to update operation", zap.Error(err))
	}

	logger.Info("Refund processed successfully",
		zap.String("tx_hash", txHash),
		zap.String("amount", refundAmount.String()),
	)

	return nil
}

func (s *RefundService) calculateRefundAmount(operation ClearingOperation) (sdk.Coin, error) {
	// Parse the amount paid
	paidAmount, err := sdk.ParseCoinNormalized(operation.ActualFeePaid + operation.FeeDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Estimate network fee for refund transaction
	estimatedGas := sdk.NewInt(80000) // Typical gas for bank send
	gasPrice := sdk.NewDecFromStr("0.025") // Get from chain config
	networkFee := estimatedGas.ToDec().Mul(gasPrice).TruncateInt()

	// Calculate refund amount (paid - network fee)
	refundAmount := paidAmount.Amount.Sub(networkFee)

	// Ensure we don't refund negative amounts
	if refundAmount.IsNegative() {
		return sdk.Coin{}, fmt.Errorf("refund amount would be negative after fees")
	}

	return sdk.NewCoin(paidAmount.Denom, refundAmount), nil
}

func (s *RefundService) getServiceWalletBalance(ctx context.Context, denom string) (sdk.Int, error) {
	// TODO: Implement actual balance check via RPC
	// This is a placeholder
	return sdk.NewInt(1000000000), nil
}

func (s *RefundService) executeRefund(ctx context.Context, refund RefundableOperation, amount sdk.Coin) (string, error) {
	// Create refund message
	msg := &banktypes.MsgSend{
		FromAddress: s.serviceWallet.Address,
		ToAddress:   refund.RefundAddress,
		Amount:      sdk.NewCoins(amount),
	}

	// Add memo explaining refund
	memo := fmt.Sprintf("Refund for clearing operation %s: %s",
		truncateID(refund.OperationID), refund.RefundReason)

	// TODO: Implement actual transaction signing and broadcasting
	// This is a placeholder
	return generateTxHash(), nil
}

func (s *RefundService) alertOperators(subject string, data map[string]interface{}) {
	// TODO: Implement operator alerting (email, Slack, etc.)
	s.logger.Error("OPERATOR ALERT",
		zap.String("subject", subject),
		zap.Any("data", data),
	)
}

// Background worker to process pending refunds
func (s *RefundService) ProcessPendingRefunds(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping refund processor")
			return
		case <-ticker.C:
			s.processBatch(ctx)
		}
	}
}

func (s *RefundService) processBatch(ctx context.Context) {
	var pendingRefunds []RefundableOperation

	// Get pending refunds older than 5 minutes (to avoid race conditions)
	cutoff := time.Now().UTC().Add(-5 * time.Minute)
	if err := s.db.Where("refund_status = ? AND created_at < ?", "pending", cutoff).
		Limit(10).Find(&pendingRefunds).Error; err != nil {
		s.logger.Error("Failed to fetch pending refunds", zap.Error(err))
		return
	}

	for _, refund := range pendingRefunds {
		if err := s.ProcessRefund(ctx, refund.OperationID, refund.RefundReason); err != nil {
			s.logger.Error("Failed to process refund",
				zap.String("operation_id", refund.OperationID),
				zap.Error(err),
			)
		}
	}
}

func generateID() string {
	// TODO: Implement proper ID generation
	return fmt.Sprintf("ref_%d", time.Now().UnixNano())
}

func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func generateTxHash() string {
	// TODO: This is a placeholder
	return fmt.Sprintf("0x%064x", time.Now().UnixNano())
}