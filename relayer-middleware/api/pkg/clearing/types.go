package clearing

import (
	"time"
)

// ClearingRequest represents a request to clear packets
type ClearingRequest struct {
	WalletAddress string          `json:"walletAddress" binding:"required"`
	ChainID       string          `json:"chainId" binding:"required"`
	Type          string          `json:"type" binding:"required,oneof=packet channel bulk"`
	Targets       ClearingTargets `json:"targets" binding:"required"`
}

// ClearingTargets contains the packets or channels to clear
type ClearingTargets struct {
	Packets  []PacketIdentifier `json:"packets,omitempty"`
	Channels []ChannelPair      `json:"channels,omitempty"`
}

// PacketIdentifier uniquely identifies a packet
type PacketIdentifier struct {
	ChainID   string `json:"chainId,omitempty"`     // For execution service
	Chain     string `json:"chain" binding:"required"`
	ChannelID string `json:"channelId,omitempty"`   // For execution service
	Channel   string `json:"channel" binding:"required"`
	PortID    string `json:"portId,omitempty"`      // For execution service
	Sequence  uint64 `json:"sequence" binding:"required"`
}

// ChannelPair represents a source-destination channel pair
type ChannelPair struct {
	SrcChain    string `json:"srcChain" binding:"required"`
	DstChain    string `json:"dstChain" binding:"required"`
	SrcChannel  string `json:"srcChannel" binding:"required"`
	DstChannel  string `json:"dstChannel" binding:"required"`
}

// ClearingToken represents an authorization token for clearing
type ClearingToken struct {
	Token             string           `json:"token"`
	Version           int              `json:"version"`
	RequestType       string           `json:"requestType"`
	TargetIdentifiers ClearingTargets  `json:"targetIdentifiers"`
	WalletAddress     string           `json:"walletAddress"`
	ChainID           string           `json:"chainId"`
	IssuedAt          int64            `json:"issuedAt"`
	ExpiresAt         int64            `json:"expiresAt"`
	ServiceFee        string           `json:"serviceFee"`
	EstimatedGasFee   string           `json:"estimatedGasFee"`
	TotalRequired     string           `json:"totalRequired"`
	AcceptedDenom     string           `json:"acceptedDenom"`
	Nonce             string           `json:"nonce"`
	Signature         string           `json:"signature"`
}

// TokenResponse is the API response for token requests
type TokenResponse struct {
	Token          *ClearingToken `json:"token"`
	PaymentAddress string         `json:"paymentAddress"`
	PaymentMemo    string         `json:"paymentMemo"`
	PaymentAmount  string         `json:"paymentAmount"`
	ExpiresIn      int            `json:"expiresIn"`
	Memo           string         `json:"memo,omitempty"` // Deprecated, use PaymentMemo
}

// PaymentVerificationRequest represents a payment verification request
type PaymentVerificationRequest struct {
	Token  string `json:"token" binding:"required"`
	TxHash string `json:"txHash" binding:"required"`
}

// PaymentVerificationResponse represents the verification result
type PaymentVerificationResponse struct {
	Verified bool   `json:"verified"`
	Status   string `json:"status"` // pending, verified, insufficient, invalid
	Message  string `json:"message,omitempty"`
}

// ClearingStatus represents the current status of a clearing operation
// NOTE: Using the enhanced version from types_v2.go
// type ClearingStatus struct {
// 	Token     string         `json:"token"`
// 	Status    string         `json:"status"` // pending, paid, executing, completed, failed
// 	Payment   PaymentStatus  `json:"payment"`
// 	Execution *ExecutionInfo `json:"execution,omitempty"`
// }

// PaymentStatus contains payment information
type PaymentStatus struct {
	Received bool   `json:"received"`
	TxHash   string `json:"txHash,omitempty"`
	Amount   string `json:"amount,omitempty"`
}

// ExecutionInfo contains clearing execution details
type ExecutionInfo struct {
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	PacketsCleared int        `json:"packetsCleared,omitempty"`
	PacketsFailed  int        `json:"packetsFailed,omitempty"`
	TxHashes       []string   `json:"txHashes,omitempty"`
	Error          string     `json:"error,omitempty"`
}

// WalletAuthRequest represents a wallet authentication request
// NOTE: Using the enhanced version from types_v2.go
// type WalletAuthRequest struct {
// 	WalletAddress string `json:"walletAddress" binding:"required"`
// 	Signature     string `json:"signature" binding:"required"`
// 	Message       string `json:"message" binding:"required"`
// }

// WalletAuthResponse represents the authentication response
// NOTE: Using the enhanced version from types_v2.go
// type WalletAuthResponse struct {
// 	SessionToken string    `json:"sessionToken"`
// 	ExpiresAt    time.Time `json:"expiresAt"`
// }

// UserStatistics represents user clearing statistics
type UserStatistics struct {
	Wallet               string                 `json:"wallet"`
	TotalRequests        int                    `json:"totalRequests"`
	SuccessfulClears     int                    `json:"successfulClears"`
	FailedClears         int                    `json:"failedClears"`
	TotalPacketsCleared  int                    `json:"totalPacketsCleared"`
	TotalFeesPaid        string                 `json:"totalFeesPaid"`
	TotalGasSaved        string                 `json:"totalGasSaved"`
	SuccessRate          float64                `json:"successRate"`
	AvgClearTime         int                    `json:"avgClearTime"` // milliseconds
	MostActiveChannels   []ChannelActivity      `json:"mostActiveChannels"`
	History              []ClearingHistoryItem  `json:"history,omitempty"`
}

// ChannelActivity represents activity on a channel
type ChannelActivity struct {
	Channel string `json:"channel"`
	Count   int    `json:"count"`
}

// ClearingHistoryItem represents a historical clearing operation
type ClearingHistoryItem struct {
	Timestamp       time.Time `json:"timestamp"`
	Type            string    `json:"type"`
	PacketsCleared  int       `json:"packetsCleared"`
	Fee             string    `json:"fee"`
	TxHashes        []string  `json:"txHashes"`
}

// PlatformStatistics represents platform-wide statistics
type PlatformStatistics struct {
	Global      GlobalStats         `json:"global"`
	Daily       DailyStats          `json:"daily"`
	TopChannels []ChannelStats      `json:"topChannels"`
	PeakHours   []HourlyActivity    `json:"peakHours"`
}

// GlobalStats represents all-time statistics
type GlobalStats struct {
	TotalPacketsCleared int     `json:"totalPacketsCleared"`
	TotalUsers          int     `json:"totalUsers"`
	TotalFeesCollected  string  `json:"totalFeesCollected"`
	AvgClearTime        int     `json:"avgClearTime"` // milliseconds
	SuccessRate         float64 `json:"successRate"`
}

// DailyStats represents today's statistics
type DailyStats struct {
	PacketsCleared int    `json:"packetsCleared"`
	ActiveUsers    int    `json:"activeUsers"`
	FeesCollected  string `json:"feesCollected"`
}

// ChannelStats represents statistics for a channel
type ChannelStats struct {
	Channel        string `json:"channel"`
	PacketsCleared int    `json:"packetsCleared"`
	AvgClearTime   int    `json:"avgClearTime"` // milliseconds
}

// HourlyActivity represents activity by hour
type HourlyActivity struct {
	Hour     int `json:"hour"`     // 0-23
	Activity int `json:"activity"` // number of operations
}

// PaymentMemo represents the structured memo for payment transactions
type PaymentMemo struct {
	Version int                    `json:"v"`
	Token   string                 `json:"t"`
	Action  string                 `json:"a"`
	Data    map[string]interface{} `json:"d"`
}

// ClearingOperation represents a clearing operation in the database
type ClearingOperation struct {
	ID               uint      `json:"id"`
	Token            string    `json:"token"`
	WalletAddress    string    `json:"walletAddress"`
	OperationType    string    `json:"operationType"`
	PacketsTargeted  int       `json:"packetsTargeted"`
	PacketsCleared   int       `json:"packetsCleared"`
	PacketsFailed    int       `json:"packetsFailed"`
	StartedAt        time.Time `json:"startedAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
	DurationMs       int       `json:"durationMs,omitempty"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"errorMessage,omitempty"`
	GasUsed          string    `json:"gasUsed"`
	ActualFeePaid    string    `json:"actualFeePaid"`
	PaymentTxHash    string    `json:"paymentTxHash"`
	ExecutionTxHashes []string  `json:"executionTxHashes"`
}