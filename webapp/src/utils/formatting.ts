/**
 * Centralized formatting utilities used across the application
 */

/**
 * Format large numbers with K/M/B suffixes
 */
export function formatNumber(num: number, decimals: number = 1): string {
  if (num >= 1_000_000_000) return (num / 1_000_000_000).toFixed(decimals) + 'B'
  if (num >= 1_000_000) return (num / 1_000_000).toFixed(decimals) + 'M'
  if (num >= 1_000) return (num / 1_000).toFixed(decimals) + 'K'
  return num.toString()
}

/**
 * Format numbers with locale-specific separators
 */
export function formatNumberWithCommas(num: number): string {
  return new Intl.NumberFormat().format(num)
}

/**
 * Truncate addresses for display
 */
export function formatAddress(address: string, startChars: number = 10, endChars: number = 4): string {
  if (!address || address.length <= startChars + endChars) return address
  return `${address.slice(0, startChars)}...${address.slice(-endChars)}`
}

/**
 * Format durations in human-readable format
 */
export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

/**
 * Format crypto amounts with proper decimals and symbols
 */
export function formatAmount(amount: string | number, denom: string, decimals: number = 6): string {
  const value = typeof amount === 'string' ? parseFloat(amount) : amount
  const humanValue = value / Math.pow(10, decimals)
  
  // Common denom mappings - should come from config eventually
  const symbols: Record<string, string> = {
    'uatom': 'ATOM',
    'uosmo': 'OSMO',
    'untrn': 'NTRN',
    'ustars': 'STARS',
    'ujuno': 'JUNO',
    'uakt': 'AKT',
    'uscrt': 'SCRT',
    'uluna': 'LUNA',
    'uusd': 'USD',
    'uusdc': 'USDC',
    'uusdt': 'USDT'
  }
  
  const symbol = symbols[denom] || denom.toUpperCase().replace('U', '')
  return `${humanValue.toFixed(2)} ${symbol}`
}

/**
 * Format percentage values
 */
export function formatPercentage(value: number, decimals: number = 1): string {
  return `${value.toFixed(decimals)}%`
}

/**
 * Format success rate with color coding
 */
export function getSuccessRateClass(rate: number): string {
  if (rate >= 95) return 'text-green-600'
  if (rate >= 85) return 'text-yellow-600'
  return 'text-red-600'
}

/**
 * Format timestamps
 */
export function formatTimestamp(date: Date | string | number): string {
  const d = new Date(date)
  return d.toLocaleString()
}

/**
 * Format relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(date: Date | string | number): string {
  const d = new Date(date)
  const now = new Date()
  const seconds = Math.floor((now.getTime() - d.getTime()) / 1000)
  
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}