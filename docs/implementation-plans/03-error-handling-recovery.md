# Error Handling & Recovery Implementation Plan

## 1. Automatic Refund Mechanism

### Refund Service Implementation
```go
// relayer-middleware/api/pkg/clearing/refund.go
package clearing

import (
    "context"
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "go.uber.org/zap"
)

type RefundService struct {
    db            *gorm.DB
    cosmosClient  client.Context
    serviceWallet ServiceWallet
    logger        *zap.Logger
}

type RefundableOperation struct {
    OperationID   string
    WalletAddress string
    RefundAddress string
    ChainID       string
    AmountPaid    string
    Denom         string
    RefundReason  string
    RefundStatus  string // pending, processing, completed, failed
    RefundTxHash  string
    CreatedAt     time.Time
    ProcessedAt   *time.Time
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
        "processed_at": &now,
    }).Error; err != nil {
        logger.Error("Failed to update refund record", zap.Error(err))
        return err
    }
    
    // Update operation status
    if err := s.db.Model(&operation).Updates(map[string]interface{}{
        "refund_status": "completed",
        "refund_tx_hash": txHash,
        "refund_reason": reason,
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
    
    // Sign and broadcast transaction
    txBuilder := s.cosmosClient.TxConfig.NewTxBuilder()
    txBuilder.SetMsgs(msg)
    txBuilder.SetMemo(memo)
    
    // Sign transaction
    // ... signing logic ...
    
    // Broadcast transaction
    txResponse, err := s.cosmosClient.BroadcastTx(txBytes)
    if err != nil {
        return "", fmt.Errorf("failed to broadcast refund: %w", err)
    }
    
    return txResponse.TxHash, nil
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
```

### Integration with Clearing Service
```go
// Update execution.go to trigger refunds on failure
func (es *ExecutionService) handleClearingFailure(ctx context.Context, operationID string, err error) {
    logger := es.logger.With(
        zap.String("operation_id", operationID),
        zap.Error(err),
    )
    
    // Determine if failure is refundable
    refundReason := es.determineRefundReason(err)
    if refundReason == "" {
        logger.Info("Failure not eligible for refund")
        return
    }
    
    // Mark operation for refund
    if err := es.db.Model(&ClearingOperation{}).
        Where("id = ?", operationID).
        Updates(map[string]interface{}{
            "refund_status": "pending",
            "refund_reason": refundReason,
        }).Error; err != nil {
        logger.Error("Failed to mark operation for refund", zap.Error(err))
    }
}

func (es *ExecutionService) determineRefundReason(err error) string {
    switch {
    case errors.Is(err, ErrChannelClosed):
        return "Channel closed during clearing"
    case errors.Is(err, ErrHermesUnavailable):
        return "Clearing service temporarily unavailable"
    case errors.Is(err, ErrInsufficientGas):
        return "Insufficient gas for clearing transaction"
    default:
        return ""
    }
}
```

## 2. Retry Logic with Exponential Backoff

### Retry Implementation
```go
// relayer-middleware/api/pkg/retry/retry.go
package retry

import (
    "context"
    "fmt"
    "math"
    "math/rand"
    "time"
    
    "go.uber.org/zap"
)

type Config struct {
    MaxAttempts     int
    InitialInterval time.Duration
    MaxInterval     time.Duration
    Multiplier      float64
    RandomFactor    float64 // Jitter factor (0.0 to 1.0)
}

type Retrier struct {
    config Config
    logger *zap.Logger
}

func NewRetrier(config Config, logger *zap.Logger) *Retrier {
    return &Retrier{
        config: config,
        logger: logger,
    }
}

func DefaultConfig() Config {
    return Config{
        MaxAttempts:     3,
        InitialInterval: 1 * time.Second,
        MaxInterval:     30 * time.Second,
        Multiplier:      2.0,
        RandomFactor:    0.1,
    }
}

type OperationState struct {
    mu        sync.Mutex
    attempts  map[string]int
    versions  map[string]int
}

func (r *Retrier) Do(ctx context.Context, operation string, fn func() error) error {
    // Check if operation is idempotent
    operationID := fmt.Sprintf("%s:%s", operation, generateOperationID())
    if !r.isIdempotent(operation) {
        // Track operation state
        r.state.mu.Lock()
        if r.state.attempts[operationID] > 0 {
            r.state.mu.Unlock()
            return ErrNonIdempotentRetry
        }
        r.state.attempts[operationID]++
        r.state.mu.Unlock()
    }
    
    var lastErr error
    
    for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
        // Check context
        if err := ctx.Err(); err != nil {
            return fmt.Errorf("context cancelled: %w", err)
        }
        
        // Execute function
        if err := fn(); err != nil {
            lastErr = err
            
            // Check if error is retryable
            if !isRetryable(err) {
                r.logger.Debug("Error is not retryable",
                    zap.String("operation", operation),
                    zap.Error(err),
                )
                return err
            }
            
            // Calculate next interval
            interval := r.calculateInterval(attempt)
            
            r.logger.Info("Operation failed, retrying",
                zap.String("operation", operation),
                zap.Int("attempt", attempt+1),
                zap.Duration("retry_in", interval),
                zap.Error(err),
            )
            
            // Wait before retry
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(interval):
                continue
            }
        }
        
        // Success
        if attempt > 0 {
            r.logger.Info("Operation succeeded after retry",
                zap.String("operation", operation),
                zap.Int("attempts", attempt+1),
            )
        }
        return nil
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}

func (r *Retrier) isIdempotent(operation string) bool {
    idempotentOps := map[string]bool{
        "get_packets":     true,
        "check_status":    true,
        "verify_payment":  true,
        "clear_packets":   false, // Not idempotent
        "process_refund":  false, // Not idempotent
    }
    
    return idempotentOps[operation]
}

func (r *Retrier) calculateInterval(attempt int) time.Duration {
    // Exponential backoff
    interval := float64(r.config.InitialInterval) * math.Pow(r.config.Multiplier, float64(attempt))
    
    // Add jitter
    if r.config.RandomFactor > 0 {
        delta := interval * r.config.RandomFactor
        minInterval := interval - delta
        maxInterval := interval + delta
        
        // Generate random value in range
        rand.Seed(time.Now().UnixNano())
        interval = minInterval + (rand.Float64() * (maxInterval - minInterval))
    }
    
    // Cap at max interval
    if interval > float64(r.config.MaxInterval) {
        interval = float64(r.config.MaxInterval)
    }
    
    return time.Duration(interval)
}

// Determine if error is retryable
func isRetryable(err error) bool {
    // Define retryable errors
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        return true
    case errors.Is(err, ErrHermesTemporaryFailure):
        return true
    case errors.Is(err, ErrNetworkTimeout):
        return true
    case errors.Is(err, ErrRPCUnavailable):
        return true
    default:
        // Check for specific error messages
        errMsg := err.Error()
        retryableMessages := []string{
            "connection refused",
            "connection reset",
            "timeout",
            "temporary failure",
            "unavailable",
        }
        
        for _, msg := range retryableMessages {
            if strings.Contains(strings.ToLower(errMsg), msg) {
                return true
            }
        }
        
        return false
    }
}
```

### Integration with Clearing Execution
```go
// Update execution.go to use retry logic
func (es *ExecutionService) clearPackets(ctx context.Context, packets []PacketIdentifier) (*ClearingResult, error) {
    retrier := retry.NewRetrier(retry.Config{
        MaxAttempts:     3,
        InitialInterval: 2 * time.Second,
        MaxInterval:     30 * time.Second,
        Multiplier:      2.0,
        RandomFactor:    0.2,
    }, es.logger)
    
    var result *ClearingResult
    
    err := retrier.Do(ctx, "clear_packets", func() error {
        // Group packets by channel
        channelGroups := es.groupPacketsByChannel(packets)
        
        for channel, channelPackets := range channelGroups {
            // Clear packets for this channel
            resp, err := es.hermesClient.ClearPackets(ctx, &hermes.ClearPacketsRequest{
                Chain:   channel.ChainID,
                Channel: channel.ChannelID,
                Port:    channel.PortID,
                Sequences: channelPackets,
            })
            
            if err != nil {
                return fmt.Errorf("failed to clear packets on channel %s: %w", channel.ChannelID, err)
            }
            
            result = resp
        }
        
        return nil
    })
    
    return result, err
}
```

## 3. Circuit Breaker for Hermes

### Circuit Breaker Implementation
```go
// relayer-middleware/api/pkg/circuitbreaker/breaker.go
package circuitbreaker

import (
    "fmt"
    "sync"
    "time"
    
    "go.uber.org/zap"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    name            string
    maxFailures     int
    resetTimeout    time.Duration
    halfOpenCalls   int
    
    mu              sync.Mutex
    state           State
    failures        int
    lastFailureTime time.Time
    successCount    int
    
    logger          *zap.Logger
}

func New(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        name:          name,
        maxFailures:   maxFailures,
        resetTimeout:  resetTimeout,
        halfOpenCalls: 3, // Allow 3 calls in half-open state
        state:         StateClosed,
        logger:        logging.With(zap.String("component", "circuit_breaker"), zap.String("name", name)),
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    
    // Check if we should transition from Open to Half-Open
    if cb.state == StateOpen {
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.state = StateHalfOpen
            cb.successCount = 0
            cb.logger.Info("Circuit breaker transitioning to half-open")
        }
    }
    
    state := cb.state
    cb.mu.Unlock()
    
    // Reject calls if circuit is open
    if state == StateOpen {
        return ErrCircuitOpen
    }
    
    // Execute the function
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.onFailure()
        return err
    }
    
    cb.onSuccess()
    return nil
}

func (cb *CircuitBreaker) onFailure() {
    cb.failures++
    cb.lastFailureTime = time.Now()
    
    switch cb.state {
    case StateClosed:
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
            cb.logger.Error("Circuit breaker opened",
                zap.Int("failures", cb.failures),
                zap.Duration("reset_timeout", cb.resetTimeout),
            )
        }
    
    case StateHalfOpen:
        // Single failure in half-open state reopens the circuit
        cb.state = StateOpen
        cb.failures = cb.maxFailures
        cb.logger.Warn("Circuit breaker reopened from half-open state")
    }
}

func (cb *CircuitBreaker) onSuccess() {
    switch cb.state {
    case StateClosed:
        cb.failures = 0
    
    case StateHalfOpen:
        cb.successCount++
        if cb.successCount >= cb.halfOpenCalls {
            // Gradual ramp-up to prevent thundering herd
            cb.state = StateClosed
            cb.failures = 0
            cb.successCount = 0
            cb.maxFailures = cb.baseMaxFailures // Reset to base threshold
            cb.logger.Info("Circuit breaker closed after successful recovery")
        }
    }
}

// Adaptive circuit breaker that adjusts thresholds
type AdaptiveCircuitBreaker struct {
    *CircuitBreaker
    baseMaxFailures int
    timeOfDay       map[int]int // Hour -> failure threshold
}

func (acb *AdaptiveCircuitBreaker) updateThresholds() {
    hour := time.Now().Hour()
    
    // Adjust thresholds based on time of day
    if threshold, ok := acb.timeOfDay[hour]; ok {
        acb.maxFailures = threshold
    } else {
        // Default: higher threshold during peak hours
        if hour >= 9 && hour <= 17 {
            acb.maxFailures = acb.baseMaxFailures * 2
        } else {
            acb.maxFailures = acb.baseMaxFailures
        }
    }
}

func (cb *CircuitBreaker) onSuccess() {
    switch cb.state {
    case StateClosed:
        cb.failures = 0
    
    case StateHalfOpen:
        cb.successCount++
        if cb.successCount >= cb.halfOpenCalls {
            cb.state = StateClosed
            cb.failures = 0
            cb.successCount = 0
            cb.logger.Info("Circuit breaker closed after successful recovery")
        }
    }
}

func (cb *CircuitBreaker) State() State {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    return cb.state
}

var ErrCircuitOpen = fmt.Errorf("circuit breaker is open")
```

### Hermes Client Wrapper with Circuit Breaker
```go
// relayer-middleware/api/pkg/hermes/client_wrapper.go
package hermes

import (
    "context"
    "time"
    
    "relayooor/api/pkg/circuitbreaker"
)

type CircuitBreakerClient struct {
    client  HermesClient
    breaker *circuitbreaker.CircuitBreaker
}

func NewCircuitBreakerClient(client HermesClient) *CircuitBreakerClient {
    return &CircuitBreakerClient{
        client: client,
        breaker: circuitbreaker.New(
            "hermes",
            5,                    // Open after 5 failures
            30 * time.Second,     // Try again after 30 seconds
        ),
    }
}

func (c *CircuitBreakerClient) ClearPackets(ctx context.Context, req *ClearPacketsRequest) (*ClearPacketsResponse, error) {
    var resp *ClearPacketsResponse
    var err error
    
    circuitErr := c.breaker.Execute(func() error {
        resp, err = c.client.ClearPackets(ctx, req)
        return err
    })
    
    if circuitErr != nil {
        if circuitErr == circuitbreaker.ErrCircuitOpen {
            return nil, ErrHermesUnavailable
        }
        return nil, circuitErr
    }
    
    return resp, nil
}

func (c *CircuitBreakerClient) GetVersion(ctx context.Context) (*VersionResponse, error) {
    var resp *VersionResponse
    var err error
    
    circuitErr := c.breaker.Execute(func() error {
        resp, err = c.client.GetVersion(ctx)
        return err
    })
    
    if circuitErr != nil {
        if circuitErr == circuitbreaker.ErrCircuitOpen {
            return nil, ErrHermesUnavailable
        }
        return nil, circuitErr
    }
    
    return resp, nil
}

func (c *CircuitBreakerClient) HealthCheck() error {
    return c.breaker.Execute(func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        _, err := c.client.GetVersion(ctx)
        return err
    })
}
```

## 4. Payment Validation

### Enhanced Payment Validation
```go
// relayer-middleware/api/pkg/clearing/payment_validator.go
package clearing

import (
    "context"
    "fmt"
    "math/big"
    "strings"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type PaymentValidator struct {
    tolerance *big.Int // Allow small variance for gas estimation
}

func NewPaymentValidator() *PaymentValidator {
    // Allow 1% tolerance for gas estimation variance
    return &PaymentValidator{
        tolerance: big.NewInt(1), // 1%
    }
}

func (v *PaymentValidator) ValidatePayment(ctx context.Context, token *ClearingToken, tx *Transaction) error {
    // Extract all payments from transaction (handle multiple)
    payments, err := v.extractAllPayments(tx)
    if err != nil {
        return fmt.Errorf("failed to extract payments: %w", err)
    }
    
    // Check for multiple payments to same token
    relevantPayments := []Payment{}
    for _, payment := range payments {
        if payment.ToAddress == expectedServiceAddress {
            relevantPayments = append(relevantPayments, payment)
        }
    }
    
    if len(relevantPayments) == 0 {
        return ErrNoPaymentFound
    }
    
    // Handle multiple payments (aggregate)
    totalPaid := new(big.Int)
    seenDenoms := make(map[string]bool)
    
    for _, payment := range relevantPayments {
        if payment.Denom != token.AcceptedDenom {
            return fmt.Errorf("invalid payment denomination: expected %s, got %s", 
                token.AcceptedDenom, payment.Denom)
        }
        seenDenoms[payment.Denom] = true
        
        amount := new(big.Int)
        amount.SetString(payment.Amount, 10)
        totalPaid.Add(totalPaid, amount)
    }
    
    // Validate total amount with tolerance
    required := new(big.Int)
    required.SetString(token.TotalRequired, 10)
    
    // Calculate tolerance amount
    toleranceAmount := new(big.Int)
    toleranceAmount.Mul(required, v.tolerance)
    toleranceAmount.Div(toleranceAmount, big.NewInt(100))
    
    // Check if payment is within tolerance
    diff := new(big.Int)
    diff.Sub(totalPaid, required)
    diff.Abs(diff)
    
    if diff.Cmp(toleranceAmount) > 0 {
        // Check if overpayment (for refund)
        if totalPaid.Cmp(required) > 0 {
            return &ErrOverpayment{
                Required: required.String(),
                Paid:     totalPaid.String(),
                Denom:    token.AcceptedDenom,
            }
        }
        
        return &ErrUnderpayment{
            Required: required.String(),
            Paid:     totalPaid.String(),
            Denom:    token.AcceptedDenom,
        }
    }
    
    return nil
}

func (v *PaymentValidator) extractPaymentAmount(tx *Transaction) (string, string, error) {
    // Parse transaction to find bank send message
    for _, msg := range tx.Messages {
        if msg.Type == "/cosmos.bank.v1beta1.MsgSend" {
            // Decode message
            var sendMsg banktypes.MsgSend
            if err := proto.Unmarshal(msg.Value, &sendMsg); err != nil {
                continue
            }
            
            // Check if this is our payment
            if sendMsg.ToAddress == expectedServiceAddress {
                if len(sendMsg.Amount) != 1 {
                    return "", "", fmt.Errorf("payment must contain exactly one coin")
                }
                
                coin := sendMsg.Amount[0]
                return coin.Amount.String(), coin.Denom, nil
            }
        }
    }
    
    return "", "", fmt.Errorf("payment not found in transaction")
}

// Custom error types for better handling
type ErrOverpayment struct {
    Required string
    Paid     string
    Denom    string
}

func (e *ErrOverpayment) Error() string {
    return fmt.Sprintf("overpayment detected: required %s %s, paid %s %s",
        e.Required, e.Denom, e.Paid, e.Denom)
}

type ErrUnderpayment struct {
    Required string
    Paid     string
    Denom    string
}

func (e *ErrUnderpayment) Error() string {
    return fmt.Sprintf("insufficient payment: required %s %s, paid %s %s",
        e.Required, e.Denom, e.Paid, e.Denom)
}
```

## 5. Duplicate Payment Detection

### Duplicate Detection Implementation
```go
// relayer-middleware/api/pkg/clearing/duplicate_detector.go
package clearing

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type DuplicateDetector struct {
    redis      *redis.Client
    expiration time.Duration
}

func NewDuplicateDetector(redis *redis.Client) *DuplicateDetector {
    return &DuplicateDetector{
        redis:      redis,
        expiration: 24 * time.Hour, // Keep records for 24 hours
    }
}

func (d *DuplicateDetector) CheckDuplicate(ctx context.Context, txHash string) (bool, error) {
    key := fmt.Sprintf("payment:tx:%s", txHash)
    
    // Try Redis first
    set, err := d.redis.SetNX(ctx, key, time.Now().Unix(), d.expiration).Result()
    if err != nil {
        // Redis failure - fallback to database
        d.logger.Warn("Redis unavailable, using database fallback", zap.Error(err))
        return d.checkDuplicateDB(ctx, txHash)
    }
    
    // If set is false, key already existed (duplicate)
    if !set {
        // Also update bloom filter
        d.bloomFilter.Add([]byte(txHash))
    }
    
    return !set, nil
}

func (d *DuplicateDetector) checkDuplicateDB(ctx context.Context, txHash string) (bool, error) {
    // Quick bloom filter check first
    if d.bloomFilter.Test([]byte(txHash)) {
        // Might be duplicate, check database
        var count int64
        err := d.db.Model(&PaymentRecord{}).
            Where("tx_hash = ? AND created_at > ?", txHash, time.Now().Add(-24*time.Hour)).
            Count(&count).Error
        
        if err != nil {
            return false, err
        }
        
        if count > 0 {
            return true, nil
        }
    }
    
    // Not duplicate, add to bloom filter and database
    d.bloomFilter.Add([]byte(txHash))
    
    // Store in database
    record := &PaymentRecord{
        TxHash:    txHash,
        CreatedAt: time.Now(),
    }
    
    if err := d.db.Create(record).Error; err != nil {
        // Check if duplicate key error
        if strings.Contains(err.Error(), "duplicate key") {
            return true, nil
        }
        return false, err
    }
    
    return false, nil
}

// Bloom filter for efficient duplicate detection
type BloomFilter struct {
    mu     sync.RWMutex
    filter *bloom.BloomFilter
}

func NewBloomFilter(expectedItems int) *BloomFilter {
    // Create bloom filter with 0.1% false positive rate
    filter := bloom.NewWithEstimates(uint(expectedItems), 0.001)
    
    return &BloomFilter{
        filter: filter,
    }
}

func (bf *BloomFilter) Add(data []byte) {
    bf.mu.Lock()
    defer bf.mu.Unlock()
    bf.filter.Add(data)
}

func (bf *BloomFilter) Test(data []byte) bool {
    bf.mu.RLock()
    defer bf.mu.RUnlock()
    return bf.filter.Test(data)
}

func (d *DuplicateDetector) GetPaymentInfo(ctx context.Context, txHash string) (*PaymentInfo, error) {
    key := fmt.Sprintf("payment:info:%s", txHash)
    
    var info PaymentInfo
    if err := d.redis.Get(ctx, key).Scan(&info); err != nil {
        if err == redis.Nil {
            return nil, nil
        }
        return nil, err
    }
    
    return &info, nil
}

func (d *DuplicateDetector) StorePaymentInfo(ctx context.Context, info *PaymentInfo) error {
    key := fmt.Sprintf("payment:info:%s", info.TxHash)
    
    return d.redis.Set(ctx, key, info, d.expiration).Err()
}

type PaymentInfo struct {
    TxHash        string    `json:"tx_hash"`
    TokenID       string    `json:"token_id"`
    OperationID   string    `json:"operation_id"`
    WalletAddress string    `json:"wallet_address"`
    Amount        string    `json:"amount"`
    Denom         string    `json:"denom"`
    ProcessedAt   time.Time `json:"processed_at"`
}
```

### Integration in Payment Verification
```go
// Update service.go VerifyPayment method
func (s *Service) VerifyPayment(ctx context.Context, tokenID string, txHash string) (*PaymentVerificationResponse, error) {
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
    
    // Continue with normal verification...
    // ... existing verification logic ...
    
    // Store payment info after successful verification
    paymentInfo := &PaymentInfo{
        TxHash:        txHash,
        TokenID:       tokenID,
        OperationID:   operation.ID,
        WalletAddress: token.WalletAddress,
        Amount:        actualAmount,
        Denom:         token.AcceptedDenom,
        ProcessedAt:   time.Now().UTC(),
    }
    
    if err := s.duplicateDetector.StorePaymentInfo(ctx, paymentInfo); err != nil {
        logger.Error("Failed to store payment info", zap.Error(err))
    }
    
    return response, nil
}
```

## Testing Considerations

1. **Refund Tests**
   - Test automatic refund triggers
   - Verify refund amount calculations
   - Test refund transaction execution
   - Handle refund failures

2. **Retry Logic Tests**
   - Test exponential backoff calculation
   - Verify jitter implementation
   - Test context cancellation
   - Check max attempts limit

3. **Circuit Breaker Tests**
   - Test state transitions
   - Verify failure counting
   - Test half-open recovery
   - Check concurrent access

4. **Payment Validation Tests**
   - Test exact amount matching
   - Test tolerance calculations
   - Test overpayment detection
   - Test underpayment detection

5. **Duplicate Detection Tests**
   - Test concurrent duplicate checks
   - Verify Redis key expiration
   - Test idempotent handling
   - Check edge cases