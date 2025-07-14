import { api } from './api'

export interface UserTransfer {
  id: string
  channelId: string
  sequence: number
  sourceChain: string
  destinationChain: string
  amount: string
  denom: string
  sender: string
  receiver: string
  status: 'pending' | 'stuck' | 'completed'
  timestamp: string
  txHash: string
  stuckDuration?: string
}

export interface StuckPacket {
  id: string
  channelId: string
  sequence: number
  sourceChain: string
  destinationChain: string
  stuckDuration: string
  amount: string
  denom: string
  sender: string
  receiver: string
  timestamp: string
  txHash?: string
}

export interface ClearPacketRequest {
  packetIds: string[]
  wallet: string
  signature: string
}

export interface ClearPacketResponse {
  status: string
  txHash?: string
  cleared?: string[]
  failed?: string[]
  message?: string
}

export const packetsService = {
  // Get transfers for a specific wallet
  async getUserTransfers(walletAddress: string): Promise<UserTransfer[]> {
    const response = await api.get(`/api/user/${walletAddress}/transfers`)
    return response.data
  },

  // Get stuck packets for a specific wallet
  async getUserStuckPackets(walletAddress: string): Promise<UserTransfer[]> {
    const response = await api.get(`/api/user/${walletAddress}/stuck`)
    return response.data
  },

  // Get all stuck packets (global view)
  async getAllStuckPackets(): Promise<StuckPacket[]> {
    const response = await api.get('/api/packets/stuck')
    return response.data
  },

  // Clear stuck packets
  async clearPackets(request: ClearPacketRequest): Promise<ClearPacketResponse> {
    const response = await api.post('/api/packets/clear', request)
    return response.data
  },

  // Get packet status from the chainpulse integration
  async getPacketDetails(chain: string, channel: string, sequence: number): Promise<any> {
    const response = await api.get(`/api/v1/chainpulse/packets/${chain}/${channel}/${sequence}`)
    return response.data
  },

  // Get channel congestion data
  async getChannelCongestion(): Promise<any[]> {
    const response = await api.get('/api/v1/chainpulse/channels/congestion')
    return response.data
  },

  // Stream stuck packets (SSE)
  subscribeToStuckPackets(callback: (packets: StuckPacket[]) => void): EventSource {
    const eventSource = new EventSource(`${api.defaults.baseURL}/api/v1/packets/stuck/stream`)
    
    eventSource.onmessage = (event) => {
      try {
        const packets = JSON.parse(event.data)
        callback(packets)
      } catch (error) {
        console.error('Failed to parse stuck packets:', error)
      }
    }

    eventSource.onerror = (error) => {
      console.error('EventSource error:', error)
      eventSource.close()
    }

    return eventSource
  },

  // Helper to format stuck duration
  formatStuckDuration(duration?: string): string {
    if (!duration) return 'Unknown'
    
    // Parse duration string (e.g., "2h30m", "45m", "3h")
    const hours = duration.match(/(\d+)h/)?.[1]
    const minutes = duration.match(/(\d+)m/)?.[1]
    
    if (hours && minutes) {
      return `${hours}h ${minutes}m`
    } else if (hours) {
      return `${hours} hours`
    } else if (minutes) {
      const mins = parseInt(minutes)
      if (mins < 60) {
        return `${mins} minutes`
      } else {
        const h = Math.floor(mins / 60)
        const m = mins % 60
        return m > 0 ? `${h}h ${m}m` : `${h} hours`
      }
    }
    
    return duration
  },

  // Helper to get chain display name
  getChainDisplayName(chainId: string): string {
    const names: Record<string, string> = {
      'cosmoshub-4': 'Cosmos Hub',
      'osmosis-1': 'Osmosis',
      'neutron-1': 'Neutron',
      'juno-1': 'Juno',
      'axelar-dojo-1': 'Axelar',
      'stride-1': 'Stride',
      'stargaze-1': 'Stargaze',
      'injective-1': 'Injective'
    }
    return names[chainId] || chainId
  },

  // Helper to format amount with denom
  formatAmount(amount: string, denom: string): string {
    // Convert from minimal denom to display denom
    const decimals: Record<string, number> = {
      uatom: 6,
      uosmo: 6,
      untrn: 6,
      ujuno: 6,
      uaxl: 6,
      ustrd: 6,
      ustars: 6,
      inj: 18
    }
    
    const decimal = decimals[denom] || 6
    const num = parseFloat(amount) / Math.pow(10, decimal)
    
    // Format with appropriate precision
    if (num < 0.01) {
      return `${num.toFixed(6)} ${denom.replace('u', '').toUpperCase()}`
    } else if (num < 1) {
      return `${num.toFixed(4)} ${denom.replace('u', '').toUpperCase()}`
    } else if (num < 1000) {
      return `${num.toFixed(2)} ${denom.replace('u', '').toUpperCase()}`
    } else {
      return `${num.toLocaleString(undefined, { maximumFractionDigits: 2 })} ${denom.replace('u', '').toUpperCase()}`
    }
  }
}