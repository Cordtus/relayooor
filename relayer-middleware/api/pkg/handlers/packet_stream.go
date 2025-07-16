package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"relayooor/api/pkg/clearing"
	"relayooor/api/pkg/types"
)

// PacketStreamHandler handles cursor-based pagination for packet streams
type PacketStreamHandler struct {
	db     *gorm.DB
	cache  *clearing.PacketCache
	logger *zap.Logger
}

// NewPacketStreamHandler creates a new packet stream handler
func NewPacketStreamHandler(db *gorm.DB, cache *clearing.PacketCache, logger *zap.Logger) *PacketStreamHandler {
	return &PacketStreamHandler{
		db:     db,
		cache:  cache,
		logger: logger.With(zap.String("component", "packet_stream")),
	}
}

// GetStuckPacketsStream handles GET /api/v1/packets/stuck/stream with cursor-based pagination
func (h *PacketStreamHandler) GetStuckPacketsStream(c *gin.Context) {
	walletAddress := c.Query("wallet")
	if walletAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_WALLET",
				"message": "Wallet address is required",
			},
		})
		return
	}

	// Parse cursor request
	var cursorReq types.CursorRequest
	if err := c.ShouldBindQuery(&cursorReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid cursor parameters",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate cursor request
	if err := cursorReq.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	// Get packets with cursor
	packets, nextCursor, err := h.getPacketsWithCursor(c, walletAddress, cursorReq.Cursor, cursorReq.Limit)
	if err != nil {
		h.logger.Error("Failed to get packets with cursor",
			zap.String("wallet", walletAddress),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "FETCH_ERROR",
				"message": "Failed to fetch packets",
			},
		})
		return
	}

	// Check if client returned 304 Not Modified
	if c.Writer.Status() == http.StatusNotModified {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"packets":     packets,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
		"count":       len(packets),
	})
}

// getPacketsWithCursor fetches packets using cursor-based pagination
func (h *PacketStreamHandler) getPacketsWithCursor(c *gin.Context, wallet, cursor string, limit int) ([]clearing.PacketIdentifier, string, error) {
	// Create snapshot timestamp for consistent results
	snapshotTime := time.Now()

	// Try cache first if no cursor (first page)
	if cursor == "" {
		cached, found, err := h.cache.GetUserPackets(c.Request.Context(), wallet)
		if err == nil && found {
			// Generate ETag for cached data
			etag := h.generateETag(wallet, snapshotTime)
			h.setETagHeaders(c, etag)

			// Check if client has cached version
			if clientETag := c.GetHeader("If-None-Match"); clientETag == etag {
				c.Status(http.StatusNotModified)
				return nil, "", nil
			}

			// Return cached data if within limit
			if len(cached) <= limit {
				return cached, "", nil
			}

			// Truncate to limit and create cursor
			packets := cached[:limit]
			lastPacket := packets[len(packets)-1]
			nextCursor := types.EncodeCursor(time.Now(), fmt.Sprintf("%d", lastPacket.Sequence), snapshotTime)
			return packets, nextCursor, nil
		}
	}

	// Decode cursor if provided
	var cursorData *types.Cursor
	if cursor != "" {
		var err error
		cursorData, err = types.DecodeCursor(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		// Use cursor's snapshot time for consistency
		snapshotTime = cursorData.Snapshot
	}

	// For now, return mock data (would query from database/chainpulse)
	packets := h.generateMockPackets(wallet, cursorData, limit+1)

	// Generate ETag
	etag := h.generateETag(wallet, snapshotTime)
	h.setETagHeaders(c, etag)

	// Check if client has cached version
	if clientETag := c.GetHeader("If-None-Match"); clientETag == etag {
		c.Status(http.StatusNotModified)
		return nil, "", nil
	}

	// Check if there are more results
	var nextCursor string
	if len(packets) > limit {
		// Remove extra item
		packets = packets[:limit]

		// Create cursor from last item
		lastPacket := packets[len(packets)-1]
		nextCursor = types.EncodeCursor(time.Now(), fmt.Sprintf("%d", lastPacket.Sequence), snapshotTime)
	}

	// Update cache if this is the first page
	if cursor == "" && len(packets) > 0 {
		go func() {
			ctx := c.Request.Context()
			if err := h.cache.SetUserPackets(ctx, wallet, packets); err != nil {
				h.logger.Error("Failed to cache user packets", zap.Error(err))
			}
		}()
	}

	return packets, nextCursor, nil
}

// generateETag generates an ETag for cache validation
func (h *PacketStreamHandler) generateETag(wallet string, snapshot time.Time) string {
	data := fmt.Sprintf("%s:%d", wallet, snapshot.Unix())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf(`"%x"`, hash[:8])
}

// setETagHeaders sets ETag and cache control headers
func (h *PacketStreamHandler) setETagHeaders(c *gin.Context, etag string) {
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, max-age=60")
}

// generateMockPackets generates mock packet data for testing
func (h *PacketStreamHandler) generateMockPackets(wallet string, cursor *types.Cursor, limit int) []clearing.PacketIdentifier {
	// In production, this would query from database
	packets := make([]clearing.PacketIdentifier, 0, limit)

	startSeq := uint64(1)
	if cursor != nil {
		// Parse sequence from cursor ID
		if seq, err := strconv.ParseUint(cursor.ID, 10, 64); err == nil {
			startSeq = seq + 1
		}
	}

	for i := 0; i < limit && i < 10; i++ {
		packets = append(packets, clearing.PacketIdentifier{
			Chain:    "osmosis-1",
			Channel:  "channel-0",
			Port:     "transfer",
			Sequence: startSeq + uint64(i),
		})
	}

	return packets
}

// GetChannelPacketsStream handles channel-specific packet streams
func (h *PacketStreamHandler) GetChannelPacketsStream(c *gin.Context) {
	srcChain := c.Query("src_chain")
	srcChannel := c.Query("src_channel")

	if srcChain == "" || srcChannel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_PARAMS",
				"message": "Source chain and channel are required",
			},
		})
		return
	}

	// Parse cursor request
	var cursorReq types.CursorRequest
	if err := c.ShouldBindQuery(&cursorReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid cursor parameters",
			},
		})
		return
	}

	// Validate
	if err := cursorReq.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	// Try cache first
	cacheKey := fmt.Sprintf("%s:%s", srcChain, srcChannel)
	if cursorReq.Cursor == "" {
		cached, found, err := h.cache.GetChannelPackets(c.Request.Context(), srcChain, srcChannel)
		if err == nil && found {
			// Return cached data
			if len(cached) <= cursorReq.Limit {
				c.JSON(http.StatusOK, gin.H{
					"packets":     cached,
					"next_cursor": "",
					"has_more":    false,
					"count":       len(cached),
				})
				return
			}

			// Paginate cached data
			packets := cached[:cursorReq.Limit]
			lastPacket := packets[len(packets)-1]
			nextCursor := types.EncodeCursor(time.Now(), fmt.Sprintf("%d", lastPacket.Sequence), time.Now())

			c.JSON(http.StatusOK, gin.H{
				"packets":     packets,
				"next_cursor": nextCursor,
				"has_more":    true,
				"count":       len(packets),
			})
			return
		}
	}

	// Fetch from data source (mock for now)
	packets := h.generateMockChannelPackets(srcChain, srcChannel, cursorReq.Cursor, cursorReq.Limit+1)

	var nextCursor string
	if len(packets) > cursorReq.Limit {
		packets = packets[:cursorReq.Limit]
		lastPacket := packets[len(packets)-1]
		nextCursor = types.EncodeCursor(time.Now(), fmt.Sprintf("%d", lastPacket.Sequence), time.Now())
	}

	c.JSON(http.StatusOK, gin.H{
		"packets":     packets,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
		"count":       len(packets),
	})
}

// generateMockChannelPackets generates mock channel packet data
func (h *PacketStreamHandler) generateMockChannelPackets(chain, channel, cursor string, limit int) []clearing.PacketIdentifier {
	packets := make([]clearing.PacketIdentifier, 0, limit)

	startSeq := uint64(1000)
	if cursor != "" {
		if cursorData, err := types.DecodeCursor(cursor); err == nil {
			if seq, err := strconv.ParseUint(cursorData.ID, 10, 64); err == nil {
				startSeq = seq + 1
			}
		}
	}

	for i := 0; i < limit && i < 20; i++ {
		packets = append(packets, clearing.PacketIdentifier{
			Chain:    chain,
			Channel:  channel,
			Port:     "transfer",
			Sequence: startSeq + uint64(i),
		})
	}

	return packets
}