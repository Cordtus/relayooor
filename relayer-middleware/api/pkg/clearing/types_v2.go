package clearing

import (
	"time"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SessionData represents session information
type SessionData struct {
	Wallet    string `json:"wallet"`
	Chain     string `json:"chain"`
	ExpiresAt int64  `json:"expires_at"`
	CreatedAt int64  `json:"created_at"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

// OperationsResponse represents paginated operations response
type OperationsResponse struct {
	Operations []ClearingOperation        `json:"operations"`
	Pagination types.PaginationResponse `json:"pagination"`
}

// ClearingStatus represents the real-time status of a clearing operation
type ClearingStatus struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Progress  int       `json:"progress"`
	UpdatedAt time.Time `json:"updated_at"`
	TxHashes  []string  `json:"tx_hashes,omitempty"`
}

// ClearingResult represents the result of a clearing execution
type ClearingResult struct {
	Success   bool      `json:"success"`
	TxHashes  []string  `json:"tx_hashes"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// QueuedOperation represents an operation in the execution queue
type QueuedOperation struct {
	ID               string              `json:"id"`
	TokenID          string              `json:"token_id"`
	Packets          []PacketIdentifier  `json:"packets"`
	CreatedAt        time.Time           `json:"created_at"`
	ProcessingStarted *time.Time         `json:"processing_started,omitempty"`
}

// ChannelKey represents a unique channel identifier
type ChannelKey struct {
	ChainID   string
	ChannelID string
	PortID    string
}

// Transaction represents a blockchain transaction
type Transaction struct {
	Hash        string    `json:"hash"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      string    `json:"amount"`
	Memo        string    `json:"memo"`
	Messages    []Message `json:"messages"`
	Height      int64     `json:"height"`
	Timestamp   time.Time `json:"timestamp"`
}

// Message represents a transaction message
type Message struct {
	Type   string                 `json:"type"`
	Sender string                 `json:"sender"`
	Data   map[string]interface{} `json:"data"`
}

// ErrOverpayment represents an overpayment error
type ErrOverpayment struct {
	Required string
	Paid     string
	Excess   string
}

func (e *ErrOverpayment) Error() string {
	return "overpayment detected"
}

// IsOverpayment checks if an error is an overpayment error
func IsOverpayment(err error) bool {
	_, ok := err.(*ErrOverpayment)
	return ok
}

// ServiceWallet represents the service wallet for refunds
type ServiceWallet struct {
	Address    string
	PrivateKey string // In production, use secure key management
}

// RefundRequest represents a refund request
type RefundRequest struct {
	OperationID   string    `json:"operation_id"`
	WalletAddress string    `json:"wallet_address"`
	Amount        string    `json:"amount"`
	Denom         string    `json:"denom"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

// RefundResult represents the result of a refund operation
type RefundResult struct {
	Success  bool      `json:"success"`
	TxHash   string    `json:"tx_hash,omitempty"`
	Error    string    `json:"error,omitempty"`
	RefundAt time.Time `json:"refund_at"`
}

// WalletAuthRequest represents an enhanced wallet authentication request
type WalletAuthRequest struct {
	WalletAddress string `json:"wallet_address"`
	Message       string `json:"message"`
	Signature     string `json:"signature"`
	Chain         string `json:"chain"`
	Timestamp     int64  `json:"timestamp"`
}

// WalletAuthResponse represents an enhanced authentication response
type WalletAuthResponse struct {
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Wallet       string    `json:"wallet"`
}

// OperationTracker tracks active operations
type OperationTracker struct {
	operations map[string]*ActiveOperation
	mu         sync.RWMutex
}

// NewOperationTracker creates a new operation tracker
func NewOperationTracker() *OperationTracker {
	return &OperationTracker{
		operations: make(map[string]*ActiveOperation),
	}
}

// Add adds an operation to tracking
func (t *OperationTracker) Add(op *ActiveOperation) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.operations[op.ID] = op
}

// Remove removes an operation from tracking
func (t *OperationTracker) Remove(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.operations, id)
}

// GetAll returns all active operations
func (t *OperationTracker) GetAll() []*ActiveOperation {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	ops := make([]*ActiveOperation, 0, len(t.operations))
	for _, op := range t.operations {
		ops = append(ops, op)
	}
	return ops
}

// Count returns the number of active operations
func (t *OperationTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.operations)
}

// Additional imports
import (
	"sync"
	"relayooor/api/pkg/types"
)