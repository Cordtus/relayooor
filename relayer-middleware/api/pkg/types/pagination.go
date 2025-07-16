package types

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
)

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int    `json:"page" form:"page" binding:"min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	SortBy   string `json:"sort_by" form:"sort_by"`
	SortDir  string `json:"sort_dir" form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// PaginationResponse contains pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// CursorRequest represents cursor-based pagination parameters
type CursorRequest struct {
	Cursor string `json:"cursor" form:"cursor"`
	Limit  int    `json:"limit" form:"limit" binding:"min=1,max=100"`
}

// CursorResponse contains cursor pagination metadata
type CursorResponse struct {
	NextCursor string `json:"next_cursor"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Count      int    `json:"count"`
}

// Cursor represents an encoded cursor for pagination
type Cursor struct {
	Timestamp time.Time `json:"t"`
	ID        string    `json:"id"`
	Snapshot  time.Time `json:"s"`
}

// NewPaginationRequest creates a new pagination request with defaults
func NewPaginationRequest() PaginationRequest {
	return PaginationRequest{
		Page:     1,
		PageSize: 20,
		SortBy:   "created_at",
		SortDir:  "desc",
	}
}

// Offset calculates the database offset
func (p PaginationRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Validate validates and normalizes pagination parameters
func (p *PaginationRequest) Validate() error {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.SortDir == "" {
		p.SortDir = "desc"
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	return nil
}

// CalculatePaginationResponse calculates pagination metadata
func CalculatePaginationResponse(page, pageSize int, totalItems int64) PaginationResponse {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	
	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// EncodeCursor encodes a cursor with timestamp and ID
func EncodeCursor(timestamp time.Time, id string, snapshot time.Time) string {
	cursor := Cursor{
		Timestamp: timestamp,
		ID:        id,
		Snapshot:  snapshot,
	}
	
	data, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a cursor string
func DecodeCursor(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, errors.New("empty cursor")
	}
	
	data, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	
	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}
	
	return &cursor, nil
}

// ValidateCursorRequest validates cursor pagination parameters
func (c *CursorRequest) Validate() error {
	if c.Limit < 1 {
		c.Limit = 20
	}
	if c.Limit > 100 {
		c.Limit = 100
	}
	return nil
}

// AllowedSortColumns defines which columns can be used for sorting
var AllowedSortColumns = map[string]bool{
	"created_at":      true,
	"completed_at":    true,
	"service_fee":     true,
	"packets_cleared": true,
	"status":          true,
	"wallet_address":  true,
}

// SanitizeSortColumn ensures the sort column is allowed
func SanitizeSortColumn(column string) string {
	if !AllowedSortColumns[column] {
		return "created_at"
	}
	return column
}

// BuildSortOrder builds a safe ORDER BY clause
func BuildSortOrder(sortBy, sortDir string) string {
	column := SanitizeSortColumn(sortBy)
	
	// Ensure sort direction is valid
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}
	
	return fmt.Sprintf("%s %s", column, sortDir)
}