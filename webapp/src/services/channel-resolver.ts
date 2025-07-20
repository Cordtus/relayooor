/**
 * Channel Resolver Service
 * 
 * Fetches IBC channel information dynamically from chain REST APIs
 * to determine counterparty chains and channel mappings.
 */

import { api } from './api'
import { fetchChainRegistry } from './chain-registry'

interface ChannelInfo {
  channelId: string
  clientId: string
  connectionId: string
  counterpartyChannelId: string
  counterpartyClientId: string
  counterpartyConnectionId: string
  counterpartyChainId: string
  sourceChainId: string
  portId: string
}

interface ChainRPCEndpoint {
  chainId: string
  restUrl: string
}

// Cache for channel information to avoid repeated API calls
const channelCache = new Map<string, ChannelInfo>()

// Get chain endpoints from the centralized registry
async function getChainEndpoints(): Promise<ChainRPCEndpoint[]> {
  const registry = await fetchChainRegistry()
  const endpoints: ChainRPCEndpoint[] = []
  
  for (const [chainId, chain] of Object.entries(registry.chains)) {
    if (chain.rest_endpoint) {
      endpoints.push({
        chainId,
        restUrl: chain.rest_endpoint
      })
    }
  }
  
  return endpoints
}

/**
 * Sanitize URL by removing trailing slashes
 */
function sanitizeUrl(url: string): string {
  return url.replace(/\/+$/, '')
}

/**
 * Make API request with error handling
 */
async function fetchData<T>(url: string): Promise<T | null> {
  try {
    const response = await fetch(url)
    if (!response.ok) {
      console.error(`Failed to fetch from ${url}: ${response.status}`)
      return null
    }
    return await response.json()
  } catch (error) {
    console.error(`Error fetching data from ${url}:`, error)
    return null
  }
}

/**
 * Get REST endpoint for a given chain ID
 */
async function getChainEndpoint(chainId: string): Promise<string | null> {
  const endpoints = await getChainEndpoints()
  const endpoint = endpoints.find(e => e.chainId === chainId)
  return endpoint ? endpoint.restUrl : null
}

/**
 * Resolve channel information from source chain and channel ID
 */
export async function resolveChannel(
  sourceChainId: string,
  channelId: string,
  portId: string = 'transfer'
): Promise<ChannelInfo | null> {
  // Check cache first
  const cacheKey = `${sourceChainId}:${channelId}:${portId}`
  if (channelCache.has(cacheKey)) {
    return channelCache.get(cacheKey)!
  }

  const restUrl = await getChainEndpoint(sourceChainId)
  if (!restUrl) {
    console.warn(`No REST endpoint configured for chain ${sourceChainId}`)
    return null
  }

  const baseUrl = sanitizeUrl(restUrl)

  // Fetch channel data
  const channelData = await fetchData<any>(
    `${baseUrl}/ibc/core/channel/v1/channels/${channelId}/ports/${portId}`
  )
  if (!channelData?.channel) {
    return null
  }

  const channel = channelData.channel
  const counterpartyChannelId = channel.counterparty?.channel_id
  const connectionId = channel.connection_hops?.[0]

  if (!connectionId || !counterpartyChannelId) {
    return null
  }

  // Fetch connection data
  const connectionData = await fetchData<any>(
    `${baseUrl}/ibc/core/connection/v1/connections/${connectionId}`
  )
  if (!connectionData?.connection) {
    return null
  }

  const connection = connectionData.connection
  const clientId = connection.client_id
  const counterpartyClientId = connection.counterparty?.client_id
  const counterpartyConnectionId = connection.counterparty?.connection_id

  if (!clientId || !counterpartyClientId || !counterpartyConnectionId) {
    return null
  }

  // Fetch counterparty chain ID
  const clientStateData = await fetchData<any>(
    `${baseUrl}/ibc/core/channel/v1/channels/${channelId}/ports/${portId}/client_state`
  )
  const counterpartyChainId = clientStateData?.identified_client_state?.client_state?.chain_id

  if (!counterpartyChainId) {
    return null
  }

  const channelInfo: ChannelInfo = {
    channelId,
    clientId,
    connectionId,
    counterpartyChannelId,
    counterpartyClientId,
    counterpartyConnectionId,
    counterpartyChainId,
    sourceChainId,
    portId
  }

  // Cache the result
  channelCache.set(cacheKey, channelInfo)

  return channelInfo
}

/**
 * Batch resolve multiple channels
 */
export async function resolveChannels(
  channels: Array<{ sourceChainId: string; channelId: string; portId?: string }>
): Promise<Map<string, ChannelInfo>> {
  const results = new Map<string, ChannelInfo>()

  // Process in parallel with rate limiting
  const batchSize = 5
  for (let i = 0; i < channels.length; i += batchSize) {
    const batch = channels.slice(i, i + batchSize)
    const promises = batch.map(ch =>
      resolveChannel(ch.sourceChainId, ch.channelId, ch.portId)
    )
    
    const batchResults = await Promise.all(promises)
    
    batchResults.forEach((result, index) => {
      if (result) {
        const ch = batch[index]
        const key = `${ch.sourceChainId}:${ch.channelId}`
        results.set(key, result)
      }
    })
  }

  return results
}

/**
 * Get counterparty chain ID for a channel
 */
export async function getCounterpartyChainId(
  sourceChainId: string,
  channelId: string,
  portId: string = 'transfer'
): Promise<string | null> {
  const info = await resolveChannel(sourceChainId, channelId, portId)
  return info?.counterpartyChainId || null
}

/**
 * Clear the channel cache (useful for testing or force refresh)
 */
export function clearChannelCache(): void {
  channelCache.clear()
}

/**
 * Get cached channel info without making API calls
 */
export function getCachedChannelInfo(
  sourceChainId: string,
  channelId: string,
  portId: string = 'transfer'
): ChannelInfo | null {
  const cacheKey = `${sourceChainId}:${channelId}:${portId}`
  return channelCache.get(cacheKey) || null
}

/**
 * Initialize channel resolver with endpoint configuration
 */
export async function initializeChannelResolver(): Promise<void> {
  try {
    // In production, fetch chain endpoints from the API
    const response = await api.get('/chains/endpoints')
    if (response.data?.endpoints) {
      // Clear existing endpoints
      chainEndpoints.length = 0
      
      // Add new endpoints
      response.data.endpoints.forEach((endpoint: ChainRPCEndpoint) => {
        chainEndpoints.push(endpoint)
      })
    }
  } catch (error) {
    console.warn('Failed to fetch chain endpoints, using defaults:', error)
  }
}

// Export types
export type { ChannelInfo, ChainRPCEndpoint }