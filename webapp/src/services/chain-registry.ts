/**
 * Chain Registry Service
 * Fetches chain configuration from the backend API instead of hardcoding
 */

import { api } from '@/api/client'

export interface ChainConfig {
  chain_id: string
  chain_name: string
  address_prefix: string
  rpc_endpoint?: string
  rest_endpoint?: string
  ws_endpoint?: string
  grpc_endpoint?: string
  explorer: string
  logo?: string
}

export interface ChannelConfig {
  source_chain: string
  source_channel: string
  dest_chain: string
  dest_channel: string
  source_port: string
  dest_port: string
  status: string
}

export interface ChainRegistry {
  chains: Record<string, ChainConfig>
  channels: ChannelConfig[]
}

// Cache for chain registry data
let registryCache: ChainRegistry | null = null
let cacheExpiry: number = 0

/**
 * Fetch chain registry from backend API
 * Includes caching to avoid repeated API calls
 */
export async function fetchChainRegistry(): Promise<ChainRegistry> {
  const now = Date.now()
  
  // Return cached data if still valid (1 hour cache)
  if (registryCache && cacheExpiry > now) {
    return registryCache
  }
  
  try {
    const response = await api.get('/api/config')
    registryCache = response.data as ChainRegistry
    cacheExpiry = now + 3600000 // 1 hour cache
    return registryCache
  } catch (error) {
    console.error('Failed to fetch chain registry:', error)
    // Return empty registry as fallback
    return { chains: {}, channels: [] }
  }
}

/**
 * Get chain configuration by ID
 */
export async function getChainConfig(chainId: string): Promise<ChainConfig | undefined> {
  const registry = await fetchChainRegistry()
  return registry.chains[chainId]
}

/**
 * Get all chains
 */
export async function getAllChains(): Promise<ChainConfig[]> {
  const registry = await fetchChainRegistry()
  return Object.values(registry.chains)
}

/**
 * Find destination chain for a channel
 */
export async function getDestinationChain(
  sourceChain: string,
  sourceChannel: string
): Promise<string | undefined> {
  const registry = await fetchChainRegistry()
  const channel = registry.channels.find(
    ch => ch.source_chain === sourceChain && ch.source_channel === sourceChannel
  )
  return channel?.dest_chain
}

/**
 * Get channel information
 */
export async function getChannelInfo(
  sourceChain: string,
  sourceChannel: string
): Promise<ChannelConfig | undefined> {
  const registry = await fetchChainRegistry()
  return registry.channels.find(
    ch => ch.source_chain === sourceChain && ch.source_channel === sourceChannel
  )
}

/**
 * Invalidate the cache (useful after configuration changes)
 */
export function invalidateCache(): void {
  registryCache = null
  cacheExpiry = 0
}