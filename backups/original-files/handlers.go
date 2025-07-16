package clearing

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Handlers struct contains all clearing-related handlers
type Handlers struct {
	service          *Service
	executionService *ExecutionService
}

// NewHandlers creates new clearing handlers
func NewHandlers(redisClient *redis.Client) *Handlers {
	service := NewService(redisClient)
	executionService := NewExecutionService(service, 2)
	
	// Start execution workers
	go executionService.Start(context.Background())
	
	return &Handlers{
		service:          service,
		executionService: executionService,
	}
}

// RequestToken handles POST /api/v1/clearing/request-token
func (h *Handlers) RequestToken(c *gin.Context) {
	var request ClearingRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}
	
	// Generate token
	response, err := h.service.GenerateToken(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "TOKEN_GENERATION_FAILED",
				"message": "Failed to generate clearing token",
				"details": err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// VerifyPayment handles POST /api/v1/clearing/verify-payment
func (h *Handlers) VerifyPayment(c *gin.Context) {
	var request PaymentVerificationRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}
	
	// Verify payment
	response, err := h.service.VerifyPayment(c.Request.Context(), request.Token, request.TxHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "VERIFICATION_FAILED",
				"message": "Failed to verify payment",
				"details": err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetStatus handles GET /api/v1/clearing/status/:token
func (h *Handlers) GetStatus(c *gin.Context) {
	token := c.Param("token")
	
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_TOKEN",
				"message": "Token is required",
			},
		})
		return
	}
	
	// Get status
	status, err := h.service.GetStatus(c.Request.Context(), token)
	if err != nil {
		// Check if token not found
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "TOKEN_NOT_FOUND",
					"message": "Token not found or expired",
				},
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "STATUS_ERROR",
				"message": "Failed to get clearing status",
				"details": err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, status)
}

// WalletSignIn handles POST /api/v1/auth/wallet-sign
func (h *Handlers) WalletSignIn(c *gin.Context) {
	var request WalletAuthRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}
	
	// Verify signature (simplified for now)
	// In production, this would verify against the actual chain signature
	if !h.verifyWalletSignature(request) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "INVALID_SIGNATURE",
				"message": "Invalid wallet signature",
			},
		})
		return
	}
	
	// Generate session token
	sessionToken := generateSessionToken()
	expiresAt := time.Now().Add(SessionTTL)
	
	// Store session
	sessionKey := fmt.Sprintf("clearing:session:%s", sessionToken)
	sessionData := map[string]interface{}{
		"wallet":    request.WalletAddress,
		"expiresAt": expiresAt.Unix(),
	}
	sessionJSON, _ := json.Marshal(sessionData)
	
	if err := h.service.redisClient.Set(c.Request.Context(), sessionKey, sessionJSON, SessionTTL).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "SESSION_ERROR",
				"message": "Failed to create session",
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, WalletAuthResponse{
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
	})
}

// GetUserStatistics handles GET /api/v1/users/statistics
func (h *Handlers) GetUserStatistics(c *gin.Context) {
	// Get wallet from session
	wallet := h.getWalletFromSession(c)
	if wallet == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}
	
	// Get user statistics
	statsKey := fmt.Sprintf("clearing:user:%s:stats", wallet)
	statsData, err := h.service.redisClient.Get(c.Request.Context(), statsKey).Result()
	
	var stats UserStatistics
	if err == nil {
		json.Unmarshal([]byte(statsData), &stats)
	}
	
	// Set wallet if empty
	if stats.Wallet == "" {
		stats.Wallet = wallet
	}
	
	// Get recent history (last 10 operations)
	opsKey := fmt.Sprintf("clearing:user:%s:operations", wallet)
	recentOps, _ := h.service.redisClient.LRange(c.Request.Context(), opsKey, 0, 9).Result()
	
	history := []ClearingHistoryItem{}
	for _, opToken := range recentOps {
		opKey := fmt.Sprintf("clearing:operation:%s", opToken)
		opData, err := h.service.redisClient.Get(c.Request.Context(), opKey).Result()
		if err != nil {
			continue
		}
		
		var op ClearingOperation
		if err := json.Unmarshal([]byte(opData), &op); err != nil {
			continue
		}
		
		historyItem := ClearingHistoryItem{
			Timestamp:      op.StartedAt,
			Type:           op.OperationType,
			PacketsCleared: op.PacketsCleared,
			Fee:            op.ActualFeePaid,
			TxHashes:       op.ExecutionTxHashes,
		}
		
		history = append(history, historyItem)
	}
	
	stats.History = history
	
	c.JSON(http.StatusOK, stats)
}

// GetPlatformStatistics handles GET /api/v1/statistics/platform
func (h *Handlers) GetPlatformStatistics(c *gin.Context) {
	// This would aggregate statistics from all operations
	// For now, return mock data
	stats := PlatformStatistics{
		Global: GlobalStats{
			TotalPacketsCleared: 15420,
			TotalUsers:          342,
			TotalFeesCollected:  "1542000000", // 1542 TOKENS
			AvgClearTime:        4500,         // 4.5 seconds
			SuccessRate:         0.973,
		},
		Daily: DailyStats{
			PacketsCleared: 287,
			ActiveUsers:    42,
			FeesCollected:  "28700000", // 28.7 TOKENS
		},
		TopChannels: []ChannelStats{
			{
				Channel:        "osmosis-1/channel-0->cosmoshub-4/channel-141",
				PacketsCleared: 5420,
				AvgClearTime:   3200,
			},
			{
				Channel:        "osmosis-1/channel-1->akash-2/channel-9",
				PacketsCleared: 3180,
				AvgClearTime:   4100,
			},
		},
		PeakHours: []HourlyActivity{
			{Hour: 14, Activity: 89},
			{Hour: 15, Activity: 92},
			{Hour: 16, Activity: 87},
			{Hour: 9, Activity: 76},
			{Hour: 10, Activity: 71},
		},
	}
	
	c.JSON(http.StatusOK, stats)
}

// GetOperations handles GET /api/v1/clearing/operations
func (h *Handlers) GetOperations(c *gin.Context) {
	// Get wallet from session
	wallet := h.getWalletFromSession(c)
	if wallet == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}
	
	// Get user's operations
	opsKey := fmt.Sprintf("clearing:user:%s:operations", wallet)
	limit := int64(50) // Default limit
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	
	opTokens, err := h.service.redisClient.LRange(c.Request.Context(), opsKey, 0, limit-1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_ERROR",
				"message": "Failed to fetch operations",
			},
		})
		return
	}
	
	operations := []ClearingOperation{}
	for _, token := range opTokens {
		opKey := fmt.Sprintf("clearing:operation:%s", token)
		opData, err := h.service.redisClient.Get(c.Request.Context(), opKey).Result()
		if err != nil {
			continue
		}
		
		var op ClearingOperation
		if err := json.Unmarshal([]byte(opData), &op); err != nil {
			continue
		}
		
		operations = append(operations, op)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"operations": operations,
		"total":      len(operations),
	})
}

// Helper functions

func (h *Handlers) verifyWalletSignature(request WalletAuthRequest) bool {
	// Simplified verification for development
	// In production, this would verify the actual cryptographic signature
	expectedMessage := fmt.Sprintf("Relayooor Authentication Request\n\nWallet: %s", request.WalletAddress)
	return strings.Contains(request.Message, expectedMessage)
}

func (h *Handlers) getWalletFromSession(c *gin.Context) string {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	
	sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
	sessionKey := fmt.Sprintf("clearing:session:%s", sessionToken)
	
	sessionData, err := h.service.redisClient.Get(c.Request.Context(), sessionKey).Result()
	if err != nil {
		return ""
	}
	
	var session map[string]interface{}
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return ""
	}
	
	// Check expiry
	if expiresAt, ok := session["expiresAt"].(float64); ok {
		if time.Now().Unix() > int64(expiresAt) {
			return ""
		}
	}
	
	if wallet, ok := session["wallet"].(string); ok {
		return wallet
	}
	
	return ""
}

func generateSessionToken() string {
	return uuid.New().String()
}

// RegisterRoutes registers all clearing routes
func (h *Handlers) RegisterRoutes(router *gin.RouterGroup) {
	// Public endpoints
	router.POST("/clearing/request-token", h.RequestToken)
	router.POST("/clearing/verify-payment", h.VerifyPayment)
	router.GET("/clearing/status/:token", h.GetStatus)
	router.POST("/auth/wallet-sign", h.WalletSignIn)
	router.GET("/statistics/platform", h.GetPlatformStatistics)
	
	// Protected endpoints (require session)
	protected := router.Group("/")
	protected.Use(h.authMiddleware())
	{
		protected.GET("/users/statistics", h.GetUserStatistics)
		protected.GET("/clearing/operations", h.GetOperations)
	}
}

// authMiddleware checks for valid session token
func (h *Handlers) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		wallet := h.getWalletFromSession(c)
		if wallet == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}
		
		// Store wallet in context for handlers
		c.Set("wallet", wallet)
		c.Next()
	}
}

// Missing imports
import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	
	"github.com/google/uuid"
)