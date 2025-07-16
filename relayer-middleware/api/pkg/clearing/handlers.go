package clearing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"relayooor/api/pkg/types"
)

// HandlersV2 struct contains all clearing-related handlers with improved error handling
type HandlersV2 struct {
	service    *ServiceV2
	logger     *zap.Logger
	wsManager  *WebSocketManager
}

// NewHandlersV2 creates new clearing handlers with improved error handling
func NewHandlersV2(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) *HandlersV2 {
	config := Config{
		SecretKey:      getEnvOrDefault("CLEARING_SECRET_KEY", "default-secret-key"),
		ServiceAddress: getEnvOrDefault("SERVICE_WALLET_ADDRESS", "cosmos1service..."),
		HermesURL:      getEnvOrDefault("HERMES_REST_URL", "http://localhost:5185"),
		ChainRPCs:      parseChainRPCs(),
	}

	service := NewServiceV2(db, redisClient, config, logger)
	
	// Start service background workers
	ctx := context.Background()
	service.Start(ctx)
	
	// Create WebSocket manager
	wsManager := NewWebSocketManager(redisClient, logger)
	
	return &HandlersV2{
		service:   service,
		logger:    logger.With(zap.String("component", "handlers")),
		wsManager: wsManager,
	}
}

// RequestToken handles POST /api/v1/clearing/request-token with improved error handling
func (h *HandlersV2) RequestToken(c *gin.Context) {
	logger := h.logger.With(
		zap.String("handler", "RequestToken"),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request ClearingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: sanitizeError(err),
			},
		})
		return
	}

	// Log request details (with sensitive data masked)
	logger.Info("Token request received",
		zap.String("wallet", maskWallet(request.WalletAddress)),
		zap.Int("packet_count", len(request.Targets.Packets)),
		zap.String("chain_id", request.ChainID),
	)

	// Generate token
	response, err := h.service.GenerateToken(c.Request.Context(), request)
	if err != nil {
		logger.Error("Failed to generate token", zap.Error(err))
		
		// Map specific errors to appropriate HTTP status codes
		status, errorCode := mapErrorToResponse(err)
		c.JSON(status, ErrorResponse{
			Error: ErrorDetail{
				Code:    errorCode,
				Message: getUserFriendlyMessage(err),
				Details: sanitizeError(err),
			},
		})
		return
	}

	logger.Info("Token generated successfully",
		zap.String("token_id", response.Token.Token),
		zap.Time("expires_at", time.Unix(response.Token.ExpiresAt, 0)),
	)

	c.JSON(http.StatusOK, response)
}

// VerifyPayment handles POST /api/v1/clearing/verify-payment with duplicate detection
func (h *HandlersV2) VerifyPayment(c *gin.Context) {
	logger := h.logger.With(
		zap.String("handler", "VerifyPayment"),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request PaymentVerificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: sanitizeError(err),
			},
		})
		return
	}

	logger.Info("Payment verification request",
		zap.String("token", request.Token[:8]+"..."),
		zap.String("tx_hash", request.TxHash),
	)

	// Verify payment
	response, err := h.service.VerifyPayment(c.Request.Context(), request.Token, request.TxHash)
	if err != nil {
		// Handle specific errors
		switch {
		case errors.Is(err, ErrDuplicatePayment):
			logger.Warn("Duplicate payment detected", zap.String("tx_hash", request.TxHash))
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DUPLICATE_PAYMENT",
					Message: "This payment has already been processed",
					Details: "Transaction hash: " + request.TxHash,
				},
			})
			return
			
		case errors.Is(err, ErrTokenExpired):
			logger.Warn("Expired token used", zap.String("token", request.Token))
			c.JSON(http.StatusGone, ErrorResponse{
				Error: ErrorDetail{
					Code:    "TOKEN_EXPIRED",
					Message: "The clearing token has expired",
					Details: "Please request a new token",
				},
			})
			return
			
		case errors.Is(err, ErrInvalidToken):
			logger.Warn("Invalid token used", zap.String("token", request.Token))
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: ErrorDetail{
					Code:    "INVALID_TOKEN",
					Message: "The provided token is invalid",
				},
			})
			return
			
		default:
			logger.Error("Payment verification failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrorDetail{
					Code:    "VERIFICATION_FAILED",
					Message: "Failed to verify payment",
					Details: sanitizeError(err),
				},
			})
			return
		}
	}

	logger.Info("Payment verified successfully",
		zap.String("token", request.Token),
		zap.Bool("verified", response.Verified),
		zap.String("status", response.Status),
	)

	// Generate operation ID
	operationID := uuid.New().String()

	// Broadcast payment verification via WebSocket
	h.wsManager.Broadcast(request.Token, WebSocketMessage{
		Type:      "payment_verified",
		Token:     request.Token,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"operation_id": operationID,
			"status":       response.Status,
		},
	})

	c.JSON(http.StatusOK, response)
}

// GetStatus handles GET /api/v1/clearing/status/:token with real-time updates
func (h *HandlersV2) GetStatus(c *gin.Context) {
	token := c.Param("token")
	
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_TOKEN",
				Message: "Token is required",
			},
		})
		return
	}

	// Check if client wants SSE
	if c.GetHeader("Accept") == "text/event-stream" {
		h.handleStatusSSE(c, token)
		return
	}

	// Get status
	status, err := h.service.GetStatus(c.Request.Context(), token)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: ErrorDetail{
					Code:    "TOKEN_NOT_FOUND",
					Message: "Token not found or expired",
				},
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "STATUS_ERROR",
				Message: "Failed to get clearing status",
				Details: sanitizeError(err),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, status)
}

// handleStatusSSE handles Server-Sent Events for real-time status updates
func (h *HandlersV2) handleStatusSSE(c *gin.Context, token string) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create SSE writer
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "SSE_NOT_SUPPORTED",
				Message: "Server-sent events not supported",
			},
		})
		return
	}

	// Subscribe to updates
	updates := h.wsManager.Subscribe(token)
	defer h.wsManager.Unsubscribe(token, updates)

	// Send initial status
	status, err := h.service.GetStatus(c.Request.Context(), token)
	if err == nil {
		data, _ := json.Marshal(status)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	// Send updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update := <-updates:
			data, _ := json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case <-ticker.C:
			// Send heartbeat
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}

// WalletSignIn handles POST /api/v1/auth/wallet-sign with improved security
func (h *HandlersV2) WalletSignIn(c *gin.Context) {
	logger := h.logger.With(
		zap.String("handler", "WalletSignIn"),
		zap.String("request_id", c.GetString("request_id")),
	)

	var request WalletAuthRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: sanitizeError(err),
			},
		})
		return
	}

	logger.Info("Wallet sign-in attempt",
		zap.String("wallet", maskWallet(request.WalletAddress)),
		zap.String("chain", request.Chain),
	)

	// Verify signature
	if err := h.verifyWalletSignature(request); err != nil {
		logger.Warn("Invalid signature",
			zap.String("wallet", maskWallet(request.WalletAddress)),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_SIGNATURE",
				Message: "Invalid wallet signature",
			},
		})
		return
	}

	// Generate session token
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(SessionTTL)
	
	// Store session with additional metadata
	sessionKey := fmt.Sprintf("clearing:session:%s", sessionToken)
	sessionData := SessionData{
		Wallet:    request.WalletAddress,
		Chain:     request.Chain,
		ExpiresAt: expiresAt.Unix(),
		CreatedAt: time.Now().Unix(),
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	sessionJSON, _ := json.Marshal(sessionData)
	
	if err := h.service.redisClient.Set(c.Request.Context(), sessionKey, sessionJSON, SessionTTL).Err(); err != nil {
		logger.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "SESSION_ERROR",
				Message: "Failed to create session",
			},
		})
		return
	}

	logger.Info("Wallet signed in successfully",
		zap.String("wallet", maskWallet(request.WalletAddress)),
		zap.String("session", sessionToken[:8]+"..."),
	)

	c.JSON(http.StatusOK, WalletAuthResponse{
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
		Wallet:       request.WalletAddress,
	})
}

// GetUserStatistics handles GET /api/v1/users/statistics with caching
func (h *HandlersV2) GetUserStatistics(c *gin.Context) {
	wallet := c.GetString("wallet")
	if wallet == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	// Try cache first
	cacheKey := fmt.Sprintf("cache:user_stats:%s", wallet)
	if cached, err := h.service.redisClient.Get(c.Request.Context(), cacheKey).Result(); err == nil {
		var stats UserStatistics
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			c.JSON(http.StatusOK, stats)
			return
		}
	}

	// Get fresh statistics
	stats, err := h.getUserStatistics(c.Request.Context(), wallet)
	if err != nil {
		h.logger.Error("Failed to get user statistics",
			zap.String("wallet", maskWallet(wallet)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "STATS_ERROR",
				Message: "Failed to retrieve statistics",
			},
		})
		return
	}

	// Cache for 1 minute
	if data, err := json.Marshal(stats); err == nil {
		h.service.redisClient.Set(c.Request.Context(), cacheKey, data, time.Minute)
	}

	c.JSON(http.StatusOK, stats)
}

// GetOperations handles GET /api/v1/clearing/operations with pagination
func (h *HandlersV2) GetOperations(c *gin.Context) {
	wallet := c.GetString("wallet")
	if wallet == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	// Parse pagination parameters
	var pagination types.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_PAGINATION",
				Message: "Invalid pagination parameters",
				Details: sanitizeError(err),
			},
		})
		return
	}
	
	// Validate pagination
	if err := pagination.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_PAGINATION",
				Message: "Invalid pagination parameters",
				Details: err.Error(),
			},
		})
		return
	}

	// Get operations from database
	var operations []ClearingOperation
	var total int64

	// Build query with sorting
	sortOrder := types.BuildSortOrder(pagination.SortBy, pagination.SortDir)

	// Count total
	if err := h.service.db.Model(&ClearingOperation{}).
		Where("wallet_address = ?", wallet).
		Count(&total).Error; err != nil {
		h.logger.Error("Failed to count operations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "DB_ERROR",
				Message: "Failed to retrieve operations",
			},
		})
		return
	}

	// Get paginated results
	if err := h.service.db.
		Where("wallet_address = ?", wallet).
		Order(sortOrder).
		Offset(pagination.Offset()).
		Limit(pagination.PageSize).
		Find(&operations).Error; err != nil {
		h.logger.Error("Failed to get operations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "DB_ERROR",
				Message: "Failed to retrieve operations",
			},
		})
		return
	}

	// Calculate pagination response
	paginationResp := types.CalculatePaginationResponse(
		pagination.Page,
		pagination.PageSize,
		total,
	)
	
	c.JSON(http.StatusOK, OperationsResponse{
		Operations: operations,
		Pagination: paginationResp,
	})
}

// Helper functions

func (h *HandlersV2) verifyWalletSignature(request WalletAuthRequest) error {
	// TODO: Implement actual signature verification based on chain type
	// This would use the appropriate library for each chain (cosmos-sdk, etc.)
	
	// For now, basic validation
	expectedMessage := fmt.Sprintf("Relayooor Authentication Request\n\nWallet: %s\nTimestamp: %d", 
		request.WalletAddress, request.Timestamp)
	
	if request.Message != expectedMessage {
		return errors.New("message mismatch")
	}
	
	// Verify timestamp is recent (within 5 minutes)
	if time.Now().Unix()-request.Timestamp > 300 {
		return errors.New("signature expired")
	}
	
	return nil
}

func (h *HandlersV2) getUserStatistics(ctx context.Context, wallet string) (*UserStatistics, error) {
	var stats UserStatistics
	
	// Get aggregated stats from database
	var result struct {
		TotalRequests       int64
		SuccessfulClears    int64
		FailedClears        int64
		TotalPacketsCleared int64
		AvgClearTime        float64
		TotalFeesSpent      string
	}
	
	err := h.service.db.Model(&ClearingOperation{}).
		Select(`
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_clears,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_clears,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN packets_cleared END), 0) as total_packets_cleared,
			COALESCE(AVG(CASE WHEN status = 'completed' THEN 
				EXTRACT(EPOCH FROM (completed_at - created_at)) * 1000 
			END), 0) as avg_clear_time,
			COALESCE(SUM(CAST(actual_fee_paid AS NUMERIC)), 0)::TEXT as total_fees_spent
		`).
		Where("wallet_address = ?", wallet).
		Scan(&result).Error
	
	if err != nil {
		return nil, err
	}
	
	stats.Wallet = wallet
	stats.TotalRequests = int(result.TotalRequests)
	stats.SuccessfulClears = int(result.SuccessfulClears)
	stats.FailedClears = int(result.FailedClears)
	stats.TotalPacketsCleared = int(result.TotalPacketsCleared)
	stats.AvgClearTime = int(result.AvgClearTime)
	stats.TotalFeesPaid = result.TotalFeesSpent
	
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulClears) / float64(stats.TotalRequests)
	}
	
	// Get recent history
	var recentOps []ClearingOperation
	err = h.service.db.
		Where("wallet_address = ?", wallet).
		Order("created_at DESC").
		Limit(10).
		Find(&recentOps).Error
	
	if err == nil {
		stats.History = make([]ClearingHistoryItem, len(recentOps))
		for i, op := range recentOps {
			stats.History[i] = ClearingHistoryItem{
				Timestamp:      op.StartedAt,
				Type:           op.OperationType,
				PacketsCleared: op.PacketsCleared,
				Fee:            op.ActualFeePaid,
				TxHashes:       op.ExecutionTxHashes,
			}
		}
	}
	
	return &stats, nil
}

// Utility functions

func mapErrorToResponse(err error) (int, string) {
	switch {
	case errors.Is(err, ErrTokenExpired):
		return http.StatusGone, "TOKEN_EXPIRED"
	case errors.Is(err, ErrInvalidToken):
		return http.StatusBadRequest, "INVALID_TOKEN"
	case errors.Is(err, ErrDuplicatePayment):
		return http.StatusConflict, "DUPLICATE_PAYMENT"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}

func getUserFriendlyMessage(err error) string {
	switch {
	case errors.Is(err, ErrTokenExpired):
		return "The clearing token has expired. Please request a new one."
	case errors.Is(err, ErrInvalidToken):
		return "The provided token is invalid."
	case errors.Is(err, ErrDuplicatePayment):
		return "This payment has already been processed."
	default:
		return "An unexpected error occurred. Please try again."
	}
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	// Remove sensitive information from error messages
	msg := err.Error()
	// Remove file paths
	msg = strings.ReplaceAll(msg, "/Users/", "")
	msg = strings.ReplaceAll(msg, "/home/", "")
	// Remove potential secrets
	msg = strings.ReplaceAll(msg, "secret", "***")
	msg = strings.ReplaceAll(msg, "password", "***")
	msg = strings.ReplaceAll(msg, "key", "***")
	return msg
}

func maskWallet(wallet string) string {
	if len(wallet) < 10 {
		return wallet
	}
	return wallet[:6] + "..." + wallet[len(wallet)-4:]
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseChainRPCs() map[string]string {
	rpcs := make(map[string]string)
	// Parse from environment or use defaults
	rpcs["osmosis-1"] = getEnvOrDefault("RPC_OSMOSIS", "https://rpc.osmosis.zone")
	rpcs["cosmoshub-4"] = getEnvOrDefault("RPC_COSMOSHUB", "https://rpc.cosmos.network")
	rpcs["neutron-1"] = getEnvOrDefault("RPC_NEUTRON", "https://rpc.neutron.org")
	return rpcs
}

// RegisterRoutes registers all clearing routes with middleware
func (h *HandlersV2) RegisterRoutes(router *gin.RouterGroup) {
	// Apply request ID middleware
	router.Use(RequestIDMiddleware())
	
	// Apply logging middleware
	router.Use(LoggingMiddleware(h.logger))
	
	// Public endpoints
	public := router.Group("/")
	{
		public.POST("/clearing/request-token", h.RequestToken)
		public.POST("/clearing/verify-payment", h.VerifyPayment)
		public.GET("/clearing/status/:token", h.GetStatus)
		public.POST("/auth/wallet-sign", h.WalletSignIn)
		public.GET("/statistics/platform", h.GetPlatformStatistics)
		
		// WebSocket endpoint
		public.GET("/ws", h.wsManager.HandleWebSocket)
	}
	
	// Protected endpoints (require session)
	protected := router.Group("/")
	protected.Use(h.authMiddleware())
	{
		protected.GET("/users/statistics", h.GetUserStatistics)
		protected.GET("/clearing/operations", h.GetOperations)
	}
}

// authMiddleware checks for valid session token with improved security
func (h *HandlersV2) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: ErrorDetail{
					Code:    "MISSING_AUTH",
					Message: "Authorization header required",
				},
			})
			c.Abort()
			return
		}
		
		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
		sessionKey := fmt.Sprintf("clearing:session:%s", sessionToken)
		
		// Get session data
		sessionData, err := h.service.redisClient.Get(c.Request.Context(), sessionKey).Result()
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: ErrorDetail{
					Code:    "INVALID_SESSION",
					Message: "Invalid or expired session",
				},
			})
			c.Abort()
			return
		}
		
		var session SessionData
		if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: ErrorDetail{
					Code:    "INVALID_SESSION",
					Message: "Invalid session data",
				},
			})
			c.Abort()
			return
		}
		
		// Check expiry
		if time.Now().Unix() > session.ExpiresAt {
			// Clean up expired session
			h.service.redisClient.Del(c.Request.Context(), sessionKey)
			
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: ErrorDetail{
					Code:    "SESSION_EXPIRED",
					Message: "Session has expired",
				},
			})
			c.Abort()
			return
		}
		
		// Store wallet and session info in context
		c.Set("wallet", session.Wallet)
		c.Set("session", session)
		c.Set("session_token", sessionToken)
		
		// Update session last activity
		go func() {
			ctx := context.Background()
			h.service.redisClient.Expire(ctx, sessionKey, SessionTTL)
		}()
		
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs all requests with structured logging
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", c.GetString("request_id")),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		)
	}
}

// GetPlatformStatistics handles GET /api/v1/statistics/platform with caching
func (h *HandlersV2) GetPlatformStatistics(c *gin.Context) {
	// Try cache first
	cacheKey := "cache:platform_stats"
	if cached, err := h.service.redisClient.Get(c.Request.Context(), cacheKey).Result(); err == nil {
		var stats PlatformStatistics
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			c.JSON(http.StatusOK, stats)
			return
		}
	}

	// Calculate fresh statistics
	stats, err := h.calculatePlatformStatistics(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to calculate platform statistics", zap.Error(err))
		// Return cached data if available, even if expired
		c.JSON(http.StatusOK, getDefaultPlatformStats())
		return
	}

	// Cache for 5 minutes
	if data, err := json.Marshal(stats); err == nil {
		h.service.redisClient.Set(c.Request.Context(), cacheKey, data, 5*time.Minute)
	}

	c.JSON(http.StatusOK, stats)
}

func (h *HandlersV2) calculatePlatformStatistics(ctx context.Context) (*PlatformStatistics, error) {
	// Implementation would aggregate from database
	// For now, return enhanced mock data
	return &PlatformStatistics{
		Global: GlobalStats{
			TotalPacketsCleared: 15420,
			TotalUsers:          342,
			TotalFeesCollected:  "1542000000",
			AvgClearTime:        4500,
			SuccessRate:         0.973,
		},
		Daily: DailyStats{
			PacketsCleared: 287,
			ActiveUsers:    42,
			FeesCollected:  "28700000",
		},
		TopChannels: []ChannelStats{
			{
				Channel:        "osmosis-1/channel-0->cosmoshub-4/channel-141",
				PacketsCleared: 5420,
				AvgClearTime:   3200,
			},
		},
		PeakHours: []HourlyActivity{
			{Hour: 14, Activity: 89},
			{Hour: 15, Activity: 92},
		},
	}, nil
}

func getDefaultPlatformStats() *PlatformStatistics {
	return &PlatformStatistics{
		Global: GlobalStats{
			TotalPacketsCleared: 0,
			TotalUsers:          0,
			TotalFeesCollected:  "0",
			AvgClearTime:        0,
			SuccessRate:         0,
		},
	}
}