/**
 * Application-wide constants and configuration values
 */

// Refresh intervals (in milliseconds)
export const REFRESH_INTERVALS = {
  REAL_TIME: 1000,      // 1 second
  FREQUENT: 5000,       // 5 seconds
  NORMAL: 10000,        // 10 seconds
  RELAXED: 30000,       // 30 seconds
  SLOW: 60000,          // 1 minute
  VERY_SLOW: 300000     // 5 minutes
} as const

// API timeouts (in milliseconds)
export const API_TIMEOUTS = {
  DEFAULT: 10000,       // 10 seconds
  LONG: 30000,          // 30 seconds
  UPLOAD: 120000        // 2 minutes
} as const

// UI thresholds
export const UI_THRESHOLDS = {
  // Success rate thresholds (percentage)
  SUCCESS_RATE: {
    EXCELLENT: 95,
    GOOD: 85,
    POOR: 70
  },
  
  // Stuck packet age thresholds (seconds)
  STUCK_PACKET_AGE: {
    WARNING: 900,       // 15 minutes
    CRITICAL: 3600      // 1 hour
  },
  
  // Performance thresholds
  PERFORMANCE: {
    MIN_PACKETS_FOR_STATS: 100,
    MIN_RELAYER_ACTIVITY: 10
  }
} as const

// Pagination
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
  AVAILABLE_PAGE_SIZES: [10, 20, 50, 100]
} as const

// Cache durations (in seconds)
export const CACHE_DURATIONS = {
  CHAIN_CONFIG: 3600,   // 1 hour
  METRICS: 60,          // 1 minute
  USER_DATA: 300,       // 5 minutes
  STATIC_DATA: 86400    // 24 hours
} as const

// Token configuration
export const TOKEN_CONFIG = {
  EXPIRY_MINUTES: 5,
  VERSION: 1
} as const

// WebSocket events
export const WS_EVENTS = {
  CLEARING_UPDATE: 'clearing:update',
  CLEARING_COMPLETE: 'clearing:complete',
  CLEARING_ERROR: 'clearing:error',
  CONNECTION_STATUS: 'connection:status'
} as const

// Time range options
export const TIME_RANGES = [
  { value: '1h', label: 'Last Hour' },
  { value: '24h', label: 'Last 24 Hours' },
  { value: '7d', label: 'Last 7 Days' },
  { value: '30d', label: 'Last 30 Days' },
  { value: '90d', label: 'Last 90 Days' }
] as const

// Analytics constants
export const ANALYTICS = {
  DEFAULT_TIME_RANGE: '7d',
  GROWTH_RATE_DAILY: 0.02, // 2% daily growth for projections
  DEFAULT_PEAK_TIME: '14:00-18:00 UTC'
} as const