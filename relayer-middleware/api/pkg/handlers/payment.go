package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	
	"relayooor/api/pkg/chainpulse"
)

// PaymentHandler handles payment-related endpoints for UX improvements
type PaymentHandler struct {
	db              *gorm.DB
	redis           *redis.Client
	logger          *zap.Logger
	chainpulseClient *chainpulse.Client
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(db *gorm.DB, redis *redis.Client, chainpulseClient *chainpulse.Client, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		db:               db,
		redis:            redis,
		logger:           logger.With(zap.String("component", "payment_handler")),
		chainpulseClient: chainpulseClient,
	}
}

// GetPaymentURI handles GET /api/v1/payments/uri
// Returns a Cosmos payment URI for wallet integration
func (h *PaymentHandler) GetPaymentURI(c *gin.Context) {
	tokenID := c.Query("token")
	if tokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_TOKEN",
				"message": "Token ID is required",
			},
		})
		return
	}

	// Get token details from Redis
	tokenKey := fmt.Sprintf("token:%s", tokenID)
	tokenData, err := h.redis.Get(c.Request.Context(), tokenKey).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "TOKEN_NOT_FOUND",
				"message": "Token not found or expired",
			},
		})
		return
	}

	var token map[string]interface{}
	if err := json.Unmarshal([]byte(tokenData), &token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "PARSE_ERROR",
				"message": "Failed to parse token data",
			},
		})
		return
	}

	// Generate payment URI
	serviceAddress := token["service_address"].(string)
	amount := token["total_required"].(string)
	denom := token["accepted_denom"].(string)
	memo := fmt.Sprintf("CLR-%s", tokenID)

	// Build Cosmos URI: cosmos:<address>?amount=<amount>&denom=<denom>&memo=<memo>
	uri := fmt.Sprintf("cosmos:%s?amount=%s&denom=%s&memo=%s",
		serviceAddress, amount, denom, memo)

	// Generate QR code data URL
	qrCode, err := h.generateQRCode(uri)
	if err != nil {
		h.logger.Error("Failed to generate QR code", zap.Error(err))
		qrCode = "" // Don't fail the request
	}

	c.JSON(http.StatusOK, gin.H{
		"uri":             uri,
		"qr_code":         qrCode,
		"payment_address": serviceAddress,
		"amount":          amount,
		"denom":           denom,
		"memo":            memo,
		"expires_at":      token["expires_at"],
		"chain_id":        token["chain_id"],
	})
}

// GetPriceUSD is deprecated - we no longer use price oracles
// This endpoint returns a deprecation notice
func (h *PaymentHandler) GetPriceUSD(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"error": gin.H{
			"code":    "DEPRECATED",
			"message": "Price oracle functionality has been removed. Fees are calculated in native tokens only.",
		},
	})
}

// GetSimplifiedStatus handles GET /api/v1/clearing/simple-status
// Returns simplified status for non-technical users
func (h *PaymentHandler) GetSimplifiedStatus(c *gin.Context) {
	walletAddress := c.Query("wallet")
	if walletAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_WALLET",
				"message": "Wallet address is required",
			},
		})
		return
	}

	// Get stuck packets summary (mock for now)
	summary := h.getStuckPacketsSummary(c.Request.Context(), walletAddress)

	c.JSON(http.StatusOK, gin.H{
		"stuck_count":     summary.TotalCount,
		"total_value":     summary.TotalValue,
		"primary_denom":   summary.PrimaryDenom,
		"chains":          summary.Chains,
		"estimated_fees":  summary.EstimatedFees,
		"potential_savings": summary.PotentialSavings,
		"last_updated":    time.Now().Unix(),
	})
}

// GetFeeBreakdown handles GET /api/v1/fees/breakdown
// Returns detailed fee breakdown with USD estimates
func (h *PaymentHandler) GetFeeBreakdown(c *gin.Context) {
	packetCount := c.DefaultQuery("packets", "1")
	chainID := c.Query("chain")
	
	if chainID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_CHAIN",
				"message": "Chain ID is required",
			},
		})
		return
	}

	// Calculate fees
	breakdown := h.calculateFeeBreakdown(packetCount, chainID)

	c.JSON(http.StatusOK, breakdown)
}

// ValidatePaymentMemo handles POST /api/v1/payments/validate-memo
// Validates a payment memo format
func (h *PaymentHandler) ValidatePaymentMemo(c *gin.Context) {
	var request struct {
		Memo string `json:"memo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
			},
		})
		return
	}

	// Validate memo format
	valid, tokenID, err := h.validateMemo(request.Memo)

	response := gin.H{
		"valid": valid,
		"memo":  request.Memo,
	}

	if valid {
		response["token_id"] = tokenID
		response["message"] = "Valid payment memo"
	} else {
		response["error"] = err.Error()
		response["example"] = "CLR-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods

func (h *PaymentHandler) generateQRCode(uri string) (string, error) {
	// This would use a QR code library
	// For now, return empty string
	return "", nil
}

// fetchPrice is no longer used - removed price oracle dependency

type PacketsSummary struct {
	TotalCount       int                       `json:"total_count"`
	TotalValue       string                    `json:"total_value"`
	PrimaryDenom     string                    `json:"primary_denom"`
	Chains           []ChainSummary            `json:"chains"`
	EstimatedFees    map[string]string         `json:"estimated_fees"`
	PotentialSavings string                    `json:"potential_savings"`
}

type ChainSummary struct {
	ChainID     string `json:"chain_id"`
	ChainName   string `json:"chain_name"`
	PacketCount int    `json:"packet_count"`
	TotalValue  string `json:"total_value"`
	Denom       string `json:"denom"`
}

func (h *PaymentHandler) getStuckPacketsSummary(ctx context.Context, wallet string) *PacketsSummary {
	// Query stuck packets for the user from Chainpulse
	packets, err := h.chainpulseClient.GetPacketsByUser(ctx, wallet)
	if err != nil {
		h.logger.Error("Failed to fetch user packets from Chainpulse", zap.Error(err), zap.String("wallet", wallet))
		return &PacketsSummary{
			TotalCount:   0,
			TotalValue:   "0",
			PrimaryDenom: "",
			Chains:       []ChainSummary{},
			EstimatedFees: map[string]string{
				"service_fee": "0",
				"gas_fee":     "0",
				"total":       "0",
			},
			PotentialSavings: "0",
		}
	}

	// Aggregate packets by chain
	chainMap := make(map[string]*ChainSummary)
	totalValueByDenom := make(map[string]int64)
	var primaryDenom string
	var maxDenomValue int64

	for _, packet := range packets {
		// Skip if not stuck (check if age is significant)
		// Since we don't have a Status field, we'll assume all returned packets are stuck
		
		// Get or create chain summary
		chainSummary, exists := chainMap[packet.Chain]
		if !exists {
			chainSummary = &ChainSummary{
				ChainID:     packet.Chain,
				ChainName:   packet.Chain, // Would need chain name mapping
				PacketCount: 0,
				TotalValue:  "0",
				Denom:       packet.Denom,
			}
			chainMap[packet.Chain] = chainSummary
		}

		// Update counts
		chainSummary.PacketCount++

		// Parse and aggregate amounts
		var amountValue int64
		fmt.Sscanf(packet.Amount, "%d", &amountValue)
		
		var currentTotal int64
		fmt.Sscanf(chainSummary.TotalValue, "%d", &currentTotal)
		
		newTotal := currentTotal + amountValue
		chainSummary.TotalValue = fmt.Sprintf("%d", newTotal)
		
		// Track total by denom
		totalValueByDenom[packet.Denom] += amountValue
		if totalValueByDenom[packet.Denom] > maxDenomValue {
			maxDenomValue = totalValueByDenom[packet.Denom]
			primaryDenom = packet.Denom
		}
	}

	// Convert map to slice
	chains := make([]ChainSummary, 0, len(chainMap))
	for _, chain := range chainMap {
		chains = append(chains, *chain)
	}

	// Calculate total value (using primary denom for simplicity)
	totalValue := fmt.Sprintf("%d", maxDenomValue)

	// Calculate estimated fees based on packet count
	packetCount := len(packets)
	baseFee := int64(1000000)  // 1 TOKEN
	perPacketFee := int64(100000) // 0.1 TOKEN
	gasPrice := int64(25000)    // 0.025 TOKEN
	baseGas := int64(150000)
	perPacketGas := int64(50000)

	serviceFee := baseFee + (perPacketFee * int64(packetCount))
	gasAmount := baseGas + (perPacketGas * int64(packetCount))
	gasFee := gasAmount * gasPrice / 1000000
	totalFee := serviceFee + gasFee

	// Calculate potential savings (vs manual retry)
	manualRetryCost := gasFee * 3 // Assume 3 retries
	savings := manualRetryCost - totalFee

	return &PacketsSummary{
		TotalCount:   packetCount,
		TotalValue:   totalValue,
		PrimaryDenom: primaryDenom,
		Chains:       chains,
		EstimatedFees: map[string]string{
			"service_fee": fmt.Sprintf("%d", serviceFee),
			"gas_fee":     fmt.Sprintf("%d", gasFee),
			"total":       fmt.Sprintf("%d", totalFee),
		},
		PotentialSavings: fmt.Sprintf("%d", savings),
	}
}

func (h *PaymentHandler) calculateFeeBreakdown(packetCount string, chainID string) gin.H {
	// Base fees
	baseFee := int64(1000000)  // 1 TOKEN
	perPacketFee := int64(100000) // 0.1 TOKEN
	gasPrice := int64(25000)    // 0.025 TOKEN
	baseGas := int64(150000)
	perPacketGas := int64(50000)

	// Parse packet count
	count := int64(1)
	fmt.Sscanf(packetCount, "%d", &count)

	// Calculate totals
	serviceFee := baseFee + (perPacketFee * count)
	gasAmount := baseGas + (perPacketGas * count)
	gasFee := gasAmount * gasPrice / 1000000 // Adjust for gas price units
	total := serviceFee + gasFee

	// Get denom for chain
	denom := h.getDenomForChain(chainID)

	return gin.H{
		"service_fee": gin.H{
			"amount":     fmt.Sprintf("%d", serviceFee),
			"denom":      denom,
			"breakdown": gin.H{
				"base_fee":       fmt.Sprintf("%d", baseFee),
				"per_packet_fee": fmt.Sprintf("%d", perPacketFee),
				"packet_count":   count,
			},
		},
		"gas_fee": gin.H{
			"amount":      fmt.Sprintf("%d", gasFee),
			"denom":       denom,
			"gas_amount":  gasAmount,
			"gas_price":   gasPrice,
			"is_estimate": true,
		},
		"total": gin.H{
			"amount":    fmt.Sprintf("%d", total),
			"denom":     denom,
		},
		"comparison": gin.H{
			"manual_retry_cost": fmt.Sprintf("%d", gasFee*3), // 3 retries
			"savings":           fmt.Sprintf("%d", gasFee*3-total),
			"savings_percent":   fmt.Sprintf("%.0f", float64(gasFee*3-total)/float64(gasFee*3)*100),
		},
	}
}

func (h *PaymentHandler) getDenomForChain(chainID string) string {
	denoms := map[string]string{
		"osmosis-1":   "uosmo",
		"cosmoshub-4": "uatom",
		"neutron-1":   "untrn",
	}
	
	if denom, ok := denoms[chainID]; ok {
		return denom
	}
	
	return "uatom" // default
}

func (h *PaymentHandler) validateMemo(memo string) (bool, string, error) {
	if memo == "" {
		return false, "", fmt.Errorf("memo is required")
	}

	if len(memo) > 200 {
		return false, "", fmt.Errorf("memo too long (max 200 characters)")
	}

	// Check if it starts with CLR-
	if !strings.HasPrefix(memo, "CLR-") {
		return false, "", fmt.Errorf("memo must start with 'CLR-'")
	}

	// Extract token ID
	tokenID := strings.TrimPrefix(memo, "CLR-")
	
	// Validate UUID format (simplified)
	if len(tokenID) != 36 {
		return false, "", fmt.Errorf("invalid token format")
	}

	return true, tokenID, nil
}

// RegisterRoutes registers payment handler routes
func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/payments/uri", h.GetPaymentURI)
	router.GET("/prices/:denom", h.GetPriceUSD)
	router.GET("/clearing/simple-status", h.GetSimplifiedStatus)
	router.GET("/fees/breakdown", h.GetFeeBreakdown)
	router.POST("/payments/validate-memo", h.ValidatePaymentMemo)
}