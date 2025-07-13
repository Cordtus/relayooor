package clearing

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// PacketCache provides caching for packet data with stampede prevention
type PacketCache struct {
	redis         *redis.Client
	logger        *zap.Logger
	refreshLocks  map[string]*sync.Mutex
	refreshMutex  sync.RWMutex
	ttl           time.Duration
	gracePeriod   time.Duration
}

// CachedPacketData represents cached packet information
type CachedPacketData struct {
	Packets    []PacketIdentifier `json:"packets"`
	UpdatedAt  time.Time          `json:"updated_at"`
	ExpiresAt  time.Time          `json:"expires_at"`
	Version    int                `json:"version"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits       int64     `json:"hits"`
	Misses     int64     `json:"misses"`
	Refreshes  int64     `json:"refreshes"`
	LastReset  time.Time `json:"last_reset"`
}

// NewPacketCache creates a new packet cache
func NewPacketCache(redis *redis.Client, logger *zap.Logger) *PacketCache {
	return &PacketCache{
		redis:        redis,
		logger:       logger.With(zap.String("component", "packet_cache")),
		refreshLocks: make(map[string]*sync.Mutex),
		ttl:          5 * time.Minute,
		gracePeriod:  30 * time.Second,
	}
}

// GetCacheKey returns a cache key for a given type and identifier
func (c *PacketCache) GetCacheKey(cacheType, identifier string) string {
	return fmt.Sprintf("cache:%s:%s", cacheType, identifier)
}

// GetUserPackets retrieves cached packets for a user with stampede prevention
func (c *PacketCache) GetUserPackets(ctx context.Context, walletAddress string) ([]PacketIdentifier, error) {
	key := fmt.Sprintf("cache:user_packets:%s", walletAddress)
	
	// Try to get from cache
	data, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cached CachedPacketData
		if err := json.Unmarshal([]byte(data), &cached); err == nil {
			// Check if cache is still valid
			if time.Now().Before(cached.ExpiresAt) {
				c.incrementStat(ctx, "hits")
				return cached.Packets, nil
			}
			
			// Cache expired but within grace period - return stale data while refreshing
			if time.Now().Before(cached.ExpiresAt.Add(c.gracePeriod)) {
				go c.refreshUserPackets(context.Background(), walletAddress)
				c.incrementStat(ctx, "hits_stale")
				return cached.Packets, nil
			}
		}
	}
	
	c.incrementStat(ctx, "misses")
	
	// Cache miss or expired - need to fetch fresh data
	return c.refreshUserPackets(ctx, walletAddress)
}

// refreshUserPackets refreshes the cache with stampede prevention
func (c *PacketCache) refreshUserPackets(ctx context.Context, walletAddress string) ([]PacketIdentifier, error) {
	// Get or create mutex for this key to prevent stampede
	c.refreshMutex.Lock()
	mu, exists := c.refreshLocks[walletAddress]
	if !exists {
		mu = &sync.Mutex{}
		c.refreshLocks[walletAddress] = mu
	}
	c.refreshMutex.Unlock()
	
	// Try to acquire lock
	locked := mu.TryLock()
	if !locked {
		// Another goroutine is already refreshing
		// Wait a bit and try to get from cache again
		time.Sleep(100 * time.Millisecond)
		
		key := fmt.Sprintf("cache:user_packets:%s", walletAddress)
		data, err := c.redis.Get(ctx, key).Result()
		if err == nil {
			var cached CachedPacketData
			if err := json.Unmarshal([]byte(data), &cached); err == nil {
				return cached.Packets, nil
			}
		}
		
		// Still no data, wait for the refresh
		mu.Lock()
		defer mu.Unlock()
		
		// Try cache one more time
		data, err = c.redis.Get(ctx, key).Result()
		if err == nil {
			var cached CachedPacketData
			if err := json.Unmarshal([]byte(data), &cached); err == nil {
				return cached.Packets, nil
			}
		}
		
		return nil, fmt.Errorf("failed to refresh cache")
	}
	defer mu.Unlock()
	
	c.logger.Info("Refreshing user packets cache",
		zap.String("wallet", maskWallet(walletAddress)),
	)
	
	// Fetch fresh data (this would call the actual data source)
	packets, err := c.fetchUserPackets(ctx, walletAddress)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	cached := CachedPacketData{
		Packets:   packets,
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
		Version:   1,
	}
	
	key := fmt.Sprintf("cache:user_packets:%s", walletAddress)
	data, err := json.Marshal(cached)
	if err != nil {
		return packets, nil // Return data even if cache update fails
	}
	
	// Set with extended TTL to support grace period
	if err := c.redis.Set(ctx, key, data, c.ttl+c.gracePeriod).Err(); err != nil {
		c.logger.Error("Failed to update cache",
			zap.String("wallet", maskWallet(walletAddress)),
			zap.Error(err),
		)
	}
	
	c.incrementStat(ctx, "refreshes")
	
	// Clean up lock after a delay
	go func() {
		time.Sleep(5 * time.Second)
		c.refreshMutex.Lock()
		delete(c.refreshLocks, walletAddress)
		c.refreshMutex.Unlock()
	}()
	
	return packets, nil
}

// fetchUserPackets fetches packets from the data source
func (c *PacketCache) fetchUserPackets(ctx context.Context, walletAddress string) ([]PacketIdentifier, error) {
	// This would be implemented to fetch from Chainpulse or database
	// For now, return mock data
	return []PacketIdentifier{
		{
			Chain:    "osmosis-1",
			Channel:  "channel-0",
			Sequence: 12345,
		},
	}, nil
}

// triggerCacheRefresh triggers an asynchronous cache refresh
func (c *PacketCache) triggerCacheRefresh(ctx context.Context, walletAddress string) {
	c.logger.Debug("Triggering cache refresh", zap.String("wallet", maskWallet(walletAddress)))
	_, err := c.refreshUserPackets(ctx, walletAddress)
	if err != nil {
		c.logger.Error("Failed to refresh cache", 
			zap.String("wallet", maskWallet(walletAddress)),
			zap.Error(err),
		)
	}
}

// InvalidateUserPackets invalidates cached packets for a user
func (c *PacketCache) InvalidateUserPackets(ctx context.Context, walletAddress string) error {
	key := fmt.Sprintf("cache:user_packets:%s", walletAddress)
	return c.redis.Del(ctx, key).Err()
}

// GetChannelPackets retrieves cached packets for a channel
func (c *PacketCache) GetChannelPackets(ctx context.Context, srcChain, srcChannel string) ([]PacketIdentifier, error) {
	key := fmt.Sprintf("cache:channel_packets:%s:%s", srcChain, srcChannel)
	
	// Similar implementation to GetUserPackets
	data, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cached CachedPacketData
		if err := json.Unmarshal([]byte(data), &cached); err == nil {
			if time.Now().Before(cached.ExpiresAt) {
				c.incrementStat(ctx, "hits")
				return cached.Packets, nil
			}
		}
	}
	
	c.incrementStat(ctx, "misses")
	
	// Fetch fresh data
	packets, err := c.fetchChannelPackets(ctx, srcChain, srcChannel)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	cached := CachedPacketData{
		Packets:   packets,
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(c.ttl),
		Version:   1,
	}
	
	cacheData, err := json.Marshal(cached)
	if err == nil {
		c.redis.Set(ctx, key, cacheData, c.ttl+c.gracePeriod)
	}
	
	return packets, nil
}

// fetchChannelPackets fetches packets from the data source
func (c *PacketCache) fetchChannelPackets(ctx context.Context, srcChain, srcChannel string) ([]PacketIdentifier, error) {
	// This would be implemented to fetch from Chainpulse
	return []PacketIdentifier{}, nil
}

// WarmCache pre-populates cache for active channels
func (c *PacketCache) WarmCache(ctx context.Context, channels []ChannelPair) {
	c.logger.Info("Warming packet cache", zap.Int("channels", len(channels)))
	
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent warming
	
	for _, channel := range channels {
		wg.Add(1)
		go func(ch ChannelPair) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			if _, err := c.GetChannelPackets(ctx, ch.SrcChain, ch.SrcChannel); err != nil {
				c.logger.Error("Failed to warm cache for channel",
					zap.String("channel", ch.SrcChannel),
					zap.Error(err),
				)
			}
		}(channel)
	}
	
	wg.Wait()
	c.logger.Info("Cache warming completed")
}

// GetStats returns cache statistics
func (c *PacketCache) GetStats(ctx context.Context) (*CacheStats, error) {
	stats := &CacheStats{
		LastReset: time.Now(),
	}
	
	// Get stats from Redis
	hits, _ := c.redis.Get(ctx, "cache:stats:hits").Int64()
	misses, _ := c.redis.Get(ctx, "cache:stats:misses").Int64()
	refreshes, _ := c.redis.Get(ctx, "cache:stats:refreshes").Int64()
	
	stats.Hits = hits
	stats.Misses = misses
	stats.Refreshes = refreshes
	
	return stats, nil
}

// ResetStats resets cache statistics
func (c *PacketCache) ResetStats(ctx context.Context) error {
	keys := []string{
		"cache:stats:hits",
		"cache:stats:misses",
		"cache:stats:refreshes",
		"cache:stats:hits_stale",
	}
	
	for _, key := range keys {
		if err := c.redis.Del(ctx, key).Err(); err != nil {
			return err
		}
	}
	
	return nil
}

// incrementStat increments a cache statistic
func (c *PacketCache) incrementStat(ctx context.Context, stat string) {
	key := fmt.Sprintf("cache:stats:%s", stat)
	c.redis.Incr(ctx, key)
}

// SetTTL updates the cache TTL
func (c *PacketCache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// SetGracePeriod updates the grace period for stale cache
func (c *PacketCache) SetGracePeriod(period time.Duration) {
	c.gracePeriod = period
}

// PrefetchUserPackets prefetches packets for multiple users
func (c *PacketCache) PrefetchUserPackets(ctx context.Context, wallets []string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent prefetches
	
	for _, wallet := range wallets {
		wg.Add(1)
		go func(w string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Check if already cached
			key := fmt.Sprintf("cache:user_packets:%s", w)
			exists, _ := c.redis.Exists(ctx, key).Result()
			if exists == 0 {
				// Not cached, fetch and cache
				c.GetUserPackets(ctx, w)
			}
		}(wallet)
	}
	
	wg.Wait()
}

// ClearAll clears all cached data
func (c *PacketCache) ClearAll(ctx context.Context) error {
	// Get all cache keys
	keys, err := c.redis.Keys(ctx, "cache:*").Result()
	if err != nil {
		return err
	}
	
	if len(keys) == 0 {
		return nil
	}
	
	// Delete all cache keys
	return c.redis.Del(ctx, keys...).Err()
}