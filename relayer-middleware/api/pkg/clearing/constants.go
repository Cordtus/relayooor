package clearing

import "time"

// Token and session constants
const (
	TokenTTL    = 5 * time.Minute  // Token validity duration
	SessionTTL  = 24 * time.Hour   // Session validity duration
)

// Gas estimation constants
const (
	BaseGasAmount = 200000  // Base gas for clearing operation
	PerPacketGas  = 50000   // Additional gas per packet
)

// Fee constants (can be overridden by environment variables)
const (
	DefaultServiceFee    = 1000000  // 1 TOKEN
	DefaultPerPacketFee  = 100000   // 0.1 TOKEN per packet
)

// Pagination defaults
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// WebSocket constants
const (
	MaxConnectionsPerIP = 5
	WriteTimeout        = 10 * time.Second
	PongTimeout         = 60 * time.Second
	PingPeriod          = (PongTimeout * 9) / 10
)