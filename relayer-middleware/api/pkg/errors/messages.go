package errors

import (
	"fmt"
	"strings"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	// Payment errors
	ErrInsufficientBalance ErrorCode = "INSUFFICIENT_BALANCE"
	ErrPaymentTimeout      ErrorCode = "PAYMENT_TIMEOUT"
	ErrInvalidAmount       ErrorCode = "INVALID_AMOUNT"
	ErrDuplicatePayment    ErrorCode = "DUPLICATE_PAYMENT"
	
	// Token errors
	ErrTokenExpired    ErrorCode = "TOKEN_EXPIRED"
	ErrTokenNotFound   ErrorCode = "TOKEN_NOT_FOUND"
	ErrInvalidToken    ErrorCode = "INVALID_TOKEN"
	
	// Channel errors
	ErrChannelClosed      ErrorCode = "CHANNEL_CLOSED"
	ErrChannelUnavailable ErrorCode = "CHANNEL_UNAVAILABLE"
	
	// Service errors
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrInternalError      ErrorCode = "INTERNAL_ERROR"
	
	// Validation errors
	ErrInvalidRequest     ErrorCode = "INVALID_REQUEST"
	ErrMissingParameter   ErrorCode = "MISSING_PARAMETER"
	ErrInvalidParameter   ErrorCode = "INVALID_PARAMETER"
)

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    ErrorCode              `json:"code"`
	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Action  string                 `json:"action"`
	Icon    string                 `json:"icon,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// UserError represents an error that should be shown to users
type UserError struct {
	ErrorInfo
	HTTPStatus int    `json:"-"`
	LogMessage string `json:"-"`
}

// Error implements the error interface
func (e *UserError) Error() string {
	return e.Message
}

// ErrorMessages contains user-friendly error messages
var ErrorMessages = map[ErrorCode]ErrorInfo{
	ErrInsufficientBalance: {
		Code:    ErrInsufficientBalance,
		Title:   "Insufficient Balance",
		Message: "You need at least {{amount}} {{denom}} to complete this transaction",
		Action:  "Add funds to your wallet and try again",
		Icon:    "wallet-alert",
	},
	ErrPaymentTimeout: {
		Code:    ErrPaymentTimeout,
		Title:   "Payment Timeout",
		Message: "Your payment wasn't received within the time limit",
		Action:  "Start a new clearing request",
		Icon:    "clock-alert",
	},
	ErrTokenExpired: {
		Code:    ErrTokenExpired,
		Title:   "Request Expired",
		Message: "Your clearing request has expired for security reasons",
		Action:  "Start a new clearing request",
		Icon:    "clock-alert",
	},
	ErrChannelClosed: {
		Code:    ErrChannelClosed,
		Title:   "Channel Unavailable",
		Message: "The IBC channel for this transfer is temporarily closed",
		Action:  "Try again later or contact support",
		Icon:    "channel-alert",
	},
	ErrInvalidAmount: {
		Code:    ErrInvalidAmount,
		Title:   "Payment Amount Incorrect",
		Message: "The payment amount doesn't match the required fee",
		Action:  "Send exactly {{required}} {{denom}}",
		Icon:    "amount-alert",
	},
	ErrDuplicatePayment: {
		Code:    ErrDuplicatePayment,
		Title:   "Duplicate Payment",
		Message: "This payment has already been processed",
		Action:  "Check your clearing status or start a new request",
		Icon:    "duplicate-alert",
	},
	ErrServiceUnavailable: {
		Code:    ErrServiceUnavailable,
		Title:   "Service Temporarily Unavailable",
		Message: "Our clearing service is temporarily unavailable",
		Action:  "Please try again in a few minutes",
		Icon:    "service-alert",
	},
	ErrRateLimitExceeded: {
		Code:    ErrRateLimitExceeded,
		Title:   "Too Many Requests",
		Message: "You've made too many requests. Please slow down",
		Action:  "Wait a few minutes before trying again",
		Icon:    "rate-limit-alert",
	},
}

// NewUserError creates a new user-friendly error
func NewUserError(code ErrorCode, httpStatus int, context map[string]interface{}) *UserError {
	template, exists := ErrorMessages[code]
	if !exists {
		template = ErrorInfo{
			Code:    ErrInternalError,
			Title:   "Something Went Wrong",
			Message: "An unexpected error occurred",
			Action:  "Please try again or contact support",
			Icon:    "error-alert",
		}
	}

	// Interpolate context into message and action
	message := interpolateTemplate(template.Message, context)
	action := interpolateTemplate(template.Action, context)

	return &UserError{
		ErrorInfo: ErrorInfo{
			Code:    template.Code,
			Title:   template.Title,
			Message: message,
			Action:  action,
			Icon:    template.Icon,
			Context: context,
		},
		HTTPStatus: httpStatus,
		LogMessage: fmt.Sprintf("%s: %s", template.Code, message),
	}
}

// interpolateTemplate replaces {{key}} with values from context
func interpolateTemplate(template string, context map[string]interface{}) string {
	if context == nil {
		return template
	}

	result := template
	for key, value := range context {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// GetHTTPStatus returns the appropriate HTTP status for an error code
func GetHTTPStatus(code ErrorCode) int {
	statusMap := map[ErrorCode]int{
		ErrInsufficientBalance: 400,
		ErrPaymentTimeout:      408,
		ErrInvalidAmount:       400,
		ErrDuplicatePayment:    409,
		ErrTokenExpired:        410,
		ErrTokenNotFound:       404,
		ErrInvalidToken:        400,
		ErrChannelClosed:       503,
		ErrChannelUnavailable:  503,
		ErrServiceUnavailable:  503,
		ErrRateLimitExceeded:   429,
		ErrInternalError:       500,
		ErrInvalidRequest:      400,
		ErrMissingParameter:    400,
		ErrInvalidParameter:    400,
	}

	if status, ok := statusMap[code]; ok {
		return status
	}

	return 500
}

// FormatError formats an error for API response
func FormatError(err error) map[string]interface{} {
	if userErr, ok := err.(*UserError); ok {
		return map[string]interface{}{
			"error": map[string]interface{}{
				"code":    userErr.Code,
				"title":   userErr.Title,
				"message": userErr.Message,
				"action":  userErr.Action,
				"icon":    userErr.Icon,
			},
		}
	}

	// Generic error response
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    ErrInternalError,
			"title":   "Something Went Wrong",
			"message": "An unexpected error occurred",
			"action":  "Please try again or contact support",
		},
	}
}

// IsUserError checks if an error is a UserError
func IsUserError(err error) bool {
	_, ok := err.(*UserError)
	return ok
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if userErr, ok := err.(*UserError); ok {
		return userErr.Code
	}
	return ErrInternalError
}