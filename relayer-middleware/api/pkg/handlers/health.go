package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HealthStatus struct {
	Status           string                 `json:"status"`
	Timestamp        time.Time              `json:"timestamp"`
	Version          string                 `json:"version"`
	Checks           map[string]CheckResult `json:"checks"`
	DegradedServices []string               `json:"degraded_services,omitempty"`
}

type CheckResult struct {
	Status  string        `json:"status"`
	Latency time.Duration `json:"latency_ms"`
	Error   string        `json:"error,omitempty"`
}

type HealthHandler struct {
	db     *gorm.DB
	redis  *redis.Client
	hermes HermesClient
	logger *zap.Logger
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client, hermes HermesClient, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		hermes: hermes,
		logger: logger.With(zap.String("component", "health")),
	}
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   getVersion(),
		Checks:    make(map[string]CheckResult),
	}

	// Check database
	dbCheck := h.checkDatabase()
	health.Checks["database"] = dbCheck

	// Check Redis
	redisCheck := h.checkRedis()
	health.Checks["redis"] = redisCheck

	// Check Hermes
	hermesCheck := h.checkHermes()
	health.Checks["hermes"] = hermesCheck

	// Determine overall status with degraded mode support
	criticalFailure := false
	degradedServices := []string{}

	// Database is critical
	if dbCheck.Status != "healthy" {
		criticalFailure = true
	}

	// Redis is non-critical (can use fallbacks)
	if redisCheck.Status != "healthy" {
		degradedServices = append(degradedServices, "redis")
	}

	// Hermes is critical for clearing
	if hermesCheck.Status != "healthy" {
		criticalFailure = true
	}

	if criticalFailure {
		health.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, health)
	} else if len(degradedServices) > 0 {
		health.Status = "degraded"
		health.DegradedServices = degradedServices
		c.JSON(http.StatusOK, health) // Still return 200 for degraded
	} else {
		c.JSON(http.StatusOK, health)
	}
}

func (h *HealthHandler) GetReadiness(c *gin.Context) {
	// Lighter check for k8s readiness probe
	if err := h.db.Exec("SELECT 1").Error; err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *HealthHandler) GetLiveness(c *gin.Context) {
	// Simple liveness check for k8s
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func (h *HealthHandler) checkDatabase() CheckResult {
	start := time.Now()
	result := CheckResult{Status: "healthy"}

	err := h.db.Exec("SELECT 1").Error
	result.Latency = time.Since(start) / time.Millisecond

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
		h.logger.Error("Database health check failed", zap.Error(err))
	}

	return result
}

func (h *HealthHandler) checkRedis() CheckResult {
	start := time.Now()
	result := CheckResult{Status: "healthy"}

	ctx := context.Background()
	err := h.redis.Ping(ctx).Err()
	result.Latency = time.Since(start) / time.Millisecond

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
		h.logger.Error("Redis health check failed", zap.Error(err))
	}

	return result
}

func (h *HealthHandler) checkHermes() CheckResult {
	start := time.Now()
	result := CheckResult{Status: "healthy"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.hermes.GetVersion(ctx)
	result.Latency = time.Since(start) / time.Millisecond

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
		h.logger.Error("Hermes health check failed", zap.Error(err))
	}

	return result
}

func getVersion() string {
	// This would typically come from build info
	return "1.0.0"
}