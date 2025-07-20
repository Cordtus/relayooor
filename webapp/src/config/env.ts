// Environment configuration with sensible defaults
// All values can be overridden via environment variables

export const config = {
  // API Configuration
  apiUrl: import.meta.env.VITE_API_URL || '',  // Empty string = use relative URLs
  apiVersion: import.meta.env.VITE_API_VERSION || 'v1',
  
  // Service URLs
  chainpulseUrl: import.meta.env.VITE_CHAINPULSE_URL || 'http://localhost:3001',
  hermesUrl: import.meta.env.VITE_HERMES_URL || 'http://localhost:5185',
  
  // WebSocket Configuration
  wsProtocol: window.location.protocol === 'https:' ? 'wss:' : 'ws:',
  wsHost: import.meta.env.VITE_WS_HOST || window.location.host,
  
  // Feature Flags
  features: {
    debugMode: import.meta.env.VITE_DEBUG === 'true',
    analyticsEnabled: import.meta.env.VITE_ANALYTICS_ENABLED !== 'false',
  },
  
  // Chain Configuration
  defaultChain: import.meta.env.VITE_DEFAULT_CHAIN || 'osmosis-1',
  
  // Get full API URL
  getApiUrl(): string {
    if (this.apiUrl) {
      return this.apiUrl
    }
    // Use relative URL (works with proxy)
    return '/api'
  },
  
  // Get WebSocket URL
  getWsUrl(): string {
    return `${this.wsProtocol}//${this.wsHost}/ws`
  },
  
  // Get service URL with proper base
  getServiceUrl(service: 'chainpulse' | 'hermes'): string {
    switch (service) {
      case 'chainpulse':
        return this.chainpulseUrl
      case 'hermes':
        return this.hermesUrl
      default:
        return ''
    }
  }
}

// Export individual values for convenience
export const API_URL = config.getApiUrl()
export const WS_URL = config.getWsUrl()
export const DEFAULT_CHAIN = config.defaultChain
export const DEBUG_MODE = config.features.debugMode