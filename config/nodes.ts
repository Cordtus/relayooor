/**
 * Secure node configuration loader
 * Loads node endpoints from environment variables
 * NEVER hardcode private endpoints in this file!
 */

export interface NodeEndpoints {
  name: string
  rpc: string
  api: string
  grpc: string
  ws: string
  backupRpc?: string
  healthy: boolean
  active: boolean
}

// Load endpoint from environment with fallback to public endpoint
function getEndpoint(envVar: string, fallback?: string): string {
  const value = process.env[envVar]
  if (!value && !fallback) {
    console.warn(`Warning: ${envVar} not set and no fallback provided`)
    return ''
  }
  return value || fallback || ''
}

// Node configurations loaded from environment
export const NODES: Record<string, NodeEndpoints> = {
  'cosmoshub-4': {
    name: 'Cosmos Hub',
    rpc: getEndpoint('COSMOS_RPC_URL'),
    api: getEndpoint('COSMOS_API_URL'),
    grpc: getEndpoint('COSMOS_GRPC_URL'),
    ws: getEndpoint('COSMOS_WS_URL'),
    backupRpc: getEndpoint('COSMOS_BACKUP_RPC', 'https://cosmos-rpc.polkachu.com'),
    healthy: true,
    active: true
  },
  
  'osmosis-1': {
    name: 'Osmosis',
    rpc: getEndpoint('OSMOSIS_RPC_URL'),
    api: getEndpoint('OSMOSIS_API_URL'),
    grpc: getEndpoint('OSMOSIS_GRPC_URL'),
    ws: getEndpoint('OSMOSIS_WS_URL'),
    backupRpc: getEndpoint('OSMOSIS_BACKUP_RPC', 'https://osmosis-rpc.polkachu.com'),
    healthy: true,
    active: true
  },
  
  'neutron-1': {
    name: 'Neutron',
    rpc: getEndpoint('NEUTRON_RPC_URL'),
    api: getEndpoint('NEUTRON_API_URL'),
    grpc: getEndpoint('NEUTRON_GRPC_URL'),
    ws: getEndpoint('NEUTRON_WS_URL'),
    backupRpc: getEndpoint('NEUTRON_BACKUP_RPC', 'https://neutron-rpc.polkachu.com'),
    healthy: true,
    active: true
  },
  
  'noble-1': {
    name: 'Noble',
    rpc: getEndpoint('NOBLE_RPC_URL'),
    api: getEndpoint('NOBLE_API_URL'),
    grpc: getEndpoint('NOBLE_GRPC_URL'),
    ws: getEndpoint('NOBLE_WS_URL'),
    backupRpc: getEndpoint('NOBLE_BACKUP_RPC', 'https://noble-rpc.polkachu.com'),
    healthy: true,
    active: true
  }
}

// Get node endpoints for a specific chain
export function getNodeEndpoints(chainId: string): NodeEndpoints | undefined {
  return NODES[chainId]
}

// Get RPC endpoint for a chain
export function getRpcEndpoint(chainId: string, useBackup = false): string | undefined {
  const node = NODES[chainId]
  if (!node) return undefined
  
  // If private endpoint is not configured, always use backup
  if (!node.rpc || useBackup) {
    return node.backupRpc
  }
  
  return node.healthy ? node.rpc : node.backupRpc
}

// Get API endpoint for a chain
export function getApiEndpoint(chainId: string): string | undefined {
  const node = NODES[chainId]
  if (!node || !node.api) return undefined
  return node.healthy ? node.api : undefined
}

// Get gRPC endpoint for a chain
export function getGrpcEndpoint(chainId: string): string | undefined {
  const node = NODES[chainId]
  if (!node || !node.grpc) return undefined
  return node.healthy ? node.grpc : undefined
}

// Get WebSocket endpoint for a chain
export function getWebSocketEndpoint(chainId: string): string | undefined {
  const node = NODES[chainId]
  if (!node || !node.ws) return undefined
  return node.healthy ? node.ws : undefined
}

// Get all configured chains
export function getConfiguredChains(): string[] {
  return Object.entries(NODES)
    .filter(([_, node]) => node.rpc && node.api && node.ws)
    .map(([chainId]) => chainId)
}

// Check if private endpoints are configured
export function hasPrivateEndpoints(): boolean {
  return Object.values(NODES).some(node => 
    node.rpc && !node.rpc.includes('polkachu.com')
  )
}

// Generate chainpulse configuration
export function generateChainpulseConfig(): string {
  let output = '# Auto-generated chainpulse configuration\n\n'
  
  output += '[database]\n'
  output += 'path = "/data/chainpulse.db"\n\n'
  
  output += '[metrics]\n'
  output += 'enabled = true\n'
  output += 'port = 3001\n'
  output += 'stuck_packets = true\n'
  output += 'populate_on_start = true\n\n'
  
  // Add only chains with configured WebSocket endpoints
  const configuredChains = Object.entries(NODES)
    .filter(([_, node]) => node.ws && node.active)
  
  for (const [chainId, node] of configuredChains) {
    output += `# ${node.name}\n`
    output += `[chains.${chainId}]\n`
    output += `url = "${node.ws}"\n`
    output += `comet_version = "0.37"\n\n`
  }
  
  return output
}

// Export convenience getters (will be empty if env vars not set)
export const COSMOS_RPC = NODES['cosmoshub-4']?.rpc
export const COSMOS_API = NODES['cosmoshub-4']?.api
export const COSMOS_GRPC = NODES['cosmoshub-4']?.grpc
export const COSMOS_WS = NODES['cosmoshub-4']?.ws

export const OSMOSIS_RPC = NODES['osmosis-1']?.rpc
export const OSMOSIS_API = NODES['osmosis-1']?.api
export const OSMOSIS_GRPC = NODES['osmosis-1']?.grpc
export const OSMOSIS_WS = NODES['osmosis-1']?.ws

export const NEUTRON_RPC = NODES['neutron-1']?.rpc
export const NEUTRON_API = NODES['neutron-1']?.api
export const NEUTRON_GRPC = NODES['neutron-1']?.grpc
export const NEUTRON_WS = NODES['neutron-1']?.ws