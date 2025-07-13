package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"relayooor/api/pkg/logging"
)

// Sensitive fields that should not be logged
var sensitiveFields = map[string]bool{
	"password":    true,
	"private_key": true,
	"mnemonic":    true,
	"secret":      true,
	"token":       true,
	"signature":   true,
}

func sanitizePath(path string) string {
	// Remove potential secrets from URLs
	if strings.Contains(path, "/tokens/") {
		// Replace token IDs with placeholder
		re := regexp.MustCompile(`/tokens/[^/]+`)
		path = re.ReplaceAllString(path, "/tokens/[REDACTED]")
	}
	return path
}

func sanitizeQueryParams(query string) string {
	params, _ := url.ParseQuery(query)
	for key := range params {
		if sensitiveFields[strings.ToLower(key)] {
			params[key] = []string{"[REDACTED]"}
		}
	}
	return params.Encode()
}

func sanitizeError(err string) string {
	// Remove sensitive patterns from error messages
	for field := range sensitiveFields {
		re := regexp.MustCompile(fmt.Sprintf(`(?i)%s[=:]\s*\S+`, field))
		err = re.ReplaceAllString(err, fmt.Sprintf("%s=[REDACTED]", field))
	}
	return err
}

func hashWallet(wallet string) string {
	hash := sha256.Sum256([]byte(wallet))
	return fmt.Sprintf("%x", hash[:8])
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := sanitizePath(c.Request.URL.Path)
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			// Sanitize query parameters
			sanitizedQuery := sanitizeQueryParams(raw)
			path = path + "?" + sanitizedQuery
		}

		fields := []zap.Field{
			zap.String("client_ip", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if errorMessage != "" {
			// Sanitize error messages
			fields = append(fields, zap.String("error", sanitizeError(errorMessage)))
		}

		// Add request ID if present
		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Add user context if authenticated (hash wallet for privacy)
		if walletAddress := c.GetString("wallet_address"); walletAddress != "" {
			fields = append(fields, zap.String("wallet_hash", hashWallet(walletAddress)))
		}

		switch {
		case statusCode >= 500:
			logging.Error("Server error", fields...)
		case statusCode >= 400:
			logging.Warn("Client error", fields...)
		case statusCode >= 300:
			logging.Info("Redirection", fields...)
		default:
			logging.Info("Request completed", fields...)
		}
	}
}