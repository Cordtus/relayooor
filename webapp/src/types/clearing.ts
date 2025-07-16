export interface Transfer {
  id: string
  srcChain: string
  dstChain: string
  srcChannel: string
  dstChannel: string
  sequence: number
  sender: string
  receiver: string
  amount: string
  denom: string
  memo?: string
  txHash: string
  timestamp: number
  status: 'pending' | 'completed' | 'stuck' | 'failed'
  attempts?: number
  usdValue?: number
}

export interface StuckPacket extends Transfer {
  status: 'stuck'
  stuckSince: number
  estimatedTimeout?: number
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

export interface ClearingStatus {
  token: string
  status: 'pending' | 'paid' | 'executing' | 'completed' | 'failed'
  payment: {
    received: boolean
    txHash?: string
    amount?: string
  }
  execution?: {
    packetsCleared: number
    txHashes: string[]
    completedAt?: number
  }
}

export interface UserStatistics {
  wallet: string
  totalRequests: number
  successfulClears: number
  failedClears: number
  totalPacketsCleared: number
  avgClearTime: number
  successRate: number
  totalFeesSpent: string
  history: ClearingHistoryItem[]
}

export interface ClearingHistoryItem {
  timestamp: Date
  type: string
  packetsCleared: number
  fee: string
  txHashes: string[]
}