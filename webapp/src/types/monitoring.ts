// Comprehensive types for all Chainpulse metrics and data

export interface ChainMetrics {
  chainId: string
  chainName: string
  totalTxs: number
  totalPackets: number
  reconnects: number
  timeouts: number
  errors: number
  status: 'connected' | 'disconnected' | 'error'
  lastUpdate: Date
}

export interface IBCPacket {
  chain_id: string
  src_channel: string
  src_port: string
  dst_channel: string
  dst_port: string
  sequence?: number
  signer: string
  memo: string
  effected: boolean
  timestamp: Date
}

export interface RelayerMetrics {
  address: string
  totalPackets: number
  effectedPackets: number
  uneffectedPackets: number
  frontrunCount: number
  successRate: number
  memo: string
  software: string
  version: string
}

export interface ChannelMetrics {
  srcChain: string
  dstChain: string
  srcChannel: string
  dstChannel: string
  srcPort: string
  dstPort: string
  totalPackets: number
  effectedPackets: number
  uneffectedPackets: number
  successRate: number
  avgProcessingTime?: number
  status: 'active' | 'congested' | 'idle'
}

export interface FrontrunEvent {
  chain_id: string
  channel: string
  signer: string
  frontrunned_by: string
  count: number
  timestamp: Date
}

export interface StuckPacket {
  srcChain: string
  dstChain: string
  srcChannel: string
  sequence: number
  stuckDuration: number // in seconds
  estimatedValue?: string
  sender?: string
  receiver?: string
}

export interface SystemMetrics {
  totalChains: number
  totalTransactions: number
  totalPackets: number
  totalErrors: number
  uptime: number
  lastSync: Date
}

export interface MetricsSnapshot {
  system: SystemMetrics
  chains: ChainMetrics[]
  relayers: RelayerMetrics[]
  channels: ChannelMetrics[]
  recentPackets: IBCPacket[]
  stuckPackets: StuckPacket[]
  frontrunEvents: FrontrunEvent[]
  timestamp: Date
}