import { api } from './api'

// Types
export interface ClearingRequest {
  walletAddress: string
  chainId: string
  type: 'packet' | 'channel' | 'bulk'
  targets: {
    packets?: PacketIdentifier[]
    channels?: ChannelPair[]
  }
}

export interface PacketIdentifier {
  chain: string
  channel: string
  sequence: number
}

export interface ChannelPair {
  srcChain: string
  dstChain: string
  srcChannel: string
  dstChannel: string
}

export interface ClearingToken {
  token: string
  version: number
  requestType: string
  targetIdentifiers: any
  walletAddress: string
  chainId: string
  issuedAt: number
  expiresAt: number
  serviceFee: string
  estimatedGasFee: string
  totalRequired: string
  acceptedDenom: string
  nonce: string
  signature: string
}

export interface TokenResponse {
  token: ClearingToken
  paymentAddress: string
  memo: string
}

export interface PaymentVerificationResponse {
  verified: boolean
  status: 'pending' | 'verified' | 'insufficient' | 'invalid'
  message?: string
}

export interface ClearingStatus {
  token: string
  status: 'pending' | 'paid' | 'executing' | 'completed' | 'failed'
  payment: {
    received: boolean
    txHash?: string
    amount?: string
  }
  execution?: {
    startedAt?: string
    completedAt?: string
    packetsCleared?: number
    packetsFailed?: number
    txHashes?: string[]
    error?: string
  }
}

export interface UserStatistics {
  wallet: string
  totalRequests: number
  successfulClears: number
  failedClears: number
  totalPacketsCleared: number
  totalFeesPaid: string
  totalGasSaved: string
  successRate: number
  avgClearTime: number
  mostActiveChannels: Array<{
    channel: string
    count: number
  }>
  history?: Array<{
    timestamp: string
    type: string
    packetsCleared: number
    fee: string
    txHashes: string[]
  }>
}

export interface PlatformStatistics {
  global: {
    totalPacketsCleared: number
    totalUsers: number
    totalFeesCollected: string
    avgClearTime: number
    successRate: number
  }
  daily: {
    packetsCleared: number
    activeUsers: number
    feesCollected: string
  }
  topChannels: Array<{
    channel: string
    packetsCleared: number
    avgClearTime: number
  }>
  peakHours: Array<{
    hour: number
    activity: number
  }>
}

// Service class
export class ClearingService {
  private baseUrl = '/api'
  private wsConnection: WebSocket | null = null
  private statusCallbacks: Map<string, (status: ClearingStatus) => void> = new Map()

  // Request clearing token
  async requestToken(request: ClearingRequest): Promise<TokenResponse> {
    const response = await fetch(`${this.baseUrl}/clearing/request-token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to request token')
    }

    return response.json()
  }

  // Verify payment
  async verifyPayment(token: string, txHash: string): Promise<PaymentVerificationResponse> {
    const response = await fetch(`${this.baseUrl}/clearing/verify-payment`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ token, txHash }),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to verify payment')
    }

    return response.json()
  }

  // Get clearing status
  async getStatus(token: string): Promise<ClearingStatus> {
    const response = await fetch(`${this.baseUrl}/clearing/status/${token}`)

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to get status')
    }

    return response.json()
  }

  // Authenticate with wallet signature
  async authenticateWallet(
    walletAddress: string,
    signature: string,
    message: string
  ): Promise<{ sessionToken: string; expiresAt: string }> {
    const response = await fetch(`${this.baseUrl}/auth/wallet-sign`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ walletAddress, signature, message }),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to authenticate')
    }

    const data = await response.json()
    
    // Store session token
    localStorage.setItem('clearing_session_token', data.sessionToken)
    localStorage.setItem('clearing_session_expires', data.expiresAt)
    
    return data
  }

  // Get user statistics
  async getUserStatistics(): Promise<UserStatistics> {
    const response = await fetch(`${this.baseUrl}/users/statistics`, {
      headers: {
        'Authorization': `Bearer ${this.getSessionToken()}`,
      },
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to get statistics')
    }

    return response.json()
  }

  // Get platform statistics
  async getPlatformStatistics(): Promise<PlatformStatistics> {
    const response = await fetch(`${this.baseUrl}/statistics/platform`)

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error?.message || 'Failed to get platform statistics')
    }

    return response.json()
  }

  // Subscribe to real-time updates for a token
  subscribeToUpdates(token: string, callback: (status: ClearingStatus) => void): () => void {
    this.statusCallbacks.set(token, callback)
    
    if (!this.wsConnection) {
      this.connectWebSocket()
    }

    // Subscribe to token updates
    this.sendWebSocketMessage({
      type: 'subscribe',
      token,
    })

    // Return unsubscribe function
    return () => {
      this.statusCallbacks.delete(token)
      
      // Unsubscribe from token updates
      this.sendWebSocketMessage({
        type: 'unsubscribe',
        token,
      })
    }
  }

  // Poll for status updates (fallback for WebSocket)
  async pollForCompletion(
    token: string,
    interval = 2000,
    maxAttempts = 150
  ): Promise<ClearingStatus> {
    let attempts = 0

    while (attempts < maxAttempts) {
      const status = await this.getStatus(token)
      
      if (status.status === 'completed' || status.status === 'failed') {
        return status
      }

      await new Promise(resolve => setTimeout(resolve, interval))
      attempts++
    }

    throw new Error('Clearing operation timed out')
  }

  // Format amount for display
  formatAmount(amount: string, denom: string): string {
    const value = BigInt(amount)
    const decimals = this.getDenomDecimals(denom)
    const divisor = BigInt(10 ** decimals)
    
    const whole = value / divisor
    const fraction = value % divisor
    
    const formattedWhole = whole.toString()
    const formattedFraction = fraction.toString().padStart(decimals, '0').replace(/0+$/, '')
    
    if (formattedFraction) {
      return `${formattedWhole}.${formattedFraction} ${this.getDenomDisplay(denom)}`
    }
    
    return `${formattedWhole} ${this.getDenomDisplay(denom)}`
  }

  // Generate auth message for wallet signing
  generateAuthMessage(walletAddress: string): string {
    const timestamp = Date.now()
    const nonce = Math.random().toString(36).substring(2, 15)
    
    return `Relayooor Authentication Request\n\n` +
           `Wallet: ${walletAddress}\n` +
           `Timestamp: ${timestamp}\n` +
           `Nonce: ${nonce}\n\n` +
           `By signing this message, you authorize Relayooor to access your packet clearing history.`
  }

  // Helper methods

  private getSessionToken(): string | null {
    const token = localStorage.getItem('clearing_session_token')
    const expires = localStorage.getItem('clearing_session_expires')
    
    if (!token || !expires) {
      return null
    }
    
    // Check if expired
    if (new Date(expires) < new Date()) {
      localStorage.removeItem('clearing_session_token')
      localStorage.removeItem('clearing_session_expires')
      return null
    }
    
    return token
  }

  private connectWebSocket() {
    const wsUrl = window.location.protocol === 'https:' 
      ? `wss://${window.location.host}/api/ws/clearing-updates`
      : `ws://${window.location.host}/api/ws/clearing-updates`

    this.wsConnection = new WebSocket(wsUrl)
    
    this.wsConnection.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        
        if (data.type === 'clearing_update' && data.token) {
          const callback = this.statusCallbacks.get(data.token)
          if (callback) {
            callback(data.execution)
          }
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }
    
    this.wsConnection.onerror = (error) => {
      console.error('WebSocket error:', error)
      
      // Fallback to polling
      this.statusCallbacks.forEach((callback, token) => {
        this.pollForCompletion(token).then(callback).catch(console.error)
      })
    }
    
    this.wsConnection.onclose = () => {
      this.wsConnection = null
      
      // Reconnect after delay
      setTimeout(() => {
        if (this.statusCallbacks.size > 0) {
          this.connectWebSocket()
        }
      }, 5000)
    }
  }

  private sendWebSocketMessage(message: any) {
    if (this.wsConnection && this.wsConnection.readyState === WebSocket.OPEN) {
      this.wsConnection.send(JSON.stringify(message))
    }
  }

  private getDenomDecimals(denom: string): number {
    // Map of known denoms to decimals
    const decimals: Record<string, number> = {
      'uosmo': 6,
      'uatom': 6,
      'uion': 6,
      'ustars': 6,
      'uakt': 6,
      'untrn': 6,
    }
    
    return decimals[denom] || 6
  }

  private getDenomDisplay(denom: string): string {
    // Map of denoms to display names
    const display: Record<string, string> = {
      'uosmo': 'OSMO',
      'uatom': 'ATOM',
      'uion': 'ION',
      'ustars': 'STARS',
      'uakt': 'AKT',
      'untrn': 'NTRN',
    }
    
    return display[denom] || denom.toUpperCase().replace('U', '')
  }
}

// Export singleton instance
export const clearingService = new ClearingService()