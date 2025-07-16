package clearing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DuplicateDetector struct {
	redis       *redis.Client
	db          *gorm.DB
	expiration  time.Duration
	bloomFilter *BloomFilter
	logger      *zap.Logger
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

type PaymentRecord struct {
	TxHash    string    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	Data      string    // JSON encoded PaymentInfo
}

// Bloom filter for efficient duplicate detection
type BloomFilter struct {
	mu     sync.RWMutex
	filter *bloom.BloomFilter
}

var ErrDuplicatePayment = fmt.Errorf("duplicate payment detected")

func NewDuplicateDetector(redis *redis.Client, db *gorm.DB, logger *zap.Logger) *DuplicateDetector {
	return &DuplicateDetector{
		redis:       redis,
		db:          db,
		expiration:  24 * time.Hour, // Keep records for 24 hours
		bloomFilter: NewBloomFilter(1000000), // Expect up to 1M payments
		logger:      logger.With(zap.String("component", "duplicate_detector")),
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
		return true, nil
	}

	return false, nil
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

func (d *DuplicateDetector) GetPaymentInfo(ctx context.Context, txHash string) (*PaymentInfo, error) {
	key := fmt.Sprintf("payment:info:%s", txHash)

	// Try Redis first
	data, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Try database
			return d.getPaymentInfoDB(ctx, txHash)
		}
		return nil, err
	}

	var info PaymentInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (d *DuplicateDetector) getPaymentInfoDB(ctx context.Context, txHash string) (*PaymentInfo, error) {
	var record PaymentRecord
	if err := d.db.Where("tx_hash = ?", txHash).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if record.Data == "" {
		return nil, nil
	}

	var info PaymentInfo
	if err := json.Unmarshal([]byte(record.Data), &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (d *DuplicateDetector) StorePaymentInfo(ctx context.Context, info *PaymentInfo) error {
	key := fmt.Sprintf("payment:info:%s", info.TxHash)

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	// Store in Redis
	if err := d.redis.Set(ctx, key, data, d.expiration).Err(); err != nil {
		d.logger.Warn("Failed to store payment info in Redis, using database", zap.Error(err))
		// Fallback to database
		return d.storePaymentInfoDB(ctx, info)
	}

	// Also update database asynchronously
	go func() {
		dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		d.storePaymentInfoDB(dbCtx, info)
	}()

	return nil
}

func (d *DuplicateDetector) storePaymentInfoDB(ctx context.Context, info *PaymentInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	// Update existing record with payment info
	return d.db.Model(&PaymentRecord{}).
		Where("tx_hash = ?", info.TxHash).
		Update("data", string(data)).Error
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

// Cleanup old payment records periodically
func (d *DuplicateDetector) CleanupOldRecords(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-24 * time.Hour)
			result := d.db.Where("created_at < ?", cutoff).Delete(&PaymentRecord{})
			if result.Error != nil {
				d.logger.Error("Failed to cleanup old payment records", zap.Error(result.Error))
			} else if result.RowsAffected > 0 {
				d.logger.Info("Cleaned up old payment records", zap.Int64("count", result.RowsAffected))
			}
		}
	}
}