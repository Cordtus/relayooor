import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

export interface Metric {
  stuckPackets: number
  stuckPacketsTrend?: number
  activeChannels: number
  packetFlowRate: number
  successRate: number
  successRateTrend?: number
}

export interface Channel {
  channelId: string
  counterpartyChannelId: string
  sourceChain: string
  destinationChain: string
  state: string
  pendingPackets: number
  totalPackets: number
}

export interface StuckPacket {
  id: string
  channelId: string
  sequence: number
  sourceChain: string
  destinationChain: string
  stuckDuration: string
  sender?: string
}

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

export interface ClearPacketRequest {
  packetIds: string[]
  wallet: string
  signature: string
}

export interface ClearPacketResponse {
  status: string
  txHash?: string
  cleared: string[]
  failed: string[]
  message?: string
}

export const fetchMetrics = async (): Promise<Metric> => {
  const { data } = await api.get('/metrics')
  return data
}

export const fetchChannels = async (): Promise<Channel[]> => {
  const { data } = await api.get('/channels')
  return data
}

export const fetchStuckPackets = async (walletAddress?: string): Promise<StuckPacket[]> => {
  const params = walletAddress ? { wallet: walletAddress } : {}
  const { data } = await api.get('/packets/stuck', { params })
  return data
}

export const fetchUserTransfers = async (walletAddress: string): Promise<UserTransfer[]> => {
  const { data } = await api.get(`/user/${walletAddress}/transfers`)
  return data
}

export const fetchUserStuckPackets = async (walletAddress: string): Promise<UserTransfer[]> => {
  const { data } = await api.get(`/user/${walletAddress}/stuck`)
  return data
}

export const clearPackets = async (request: ClearPacketRequest): Promise<ClearPacketResponse> => {
  const { data } = await api.post('/packets/clear', request)
  return data
}