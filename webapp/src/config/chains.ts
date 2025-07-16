/**
 * Chain configuration and utilities
 * Enhanced with more chains and detailed configurations
 */

export interface ChainInfo {
  name: string
  prefix: string
  color: string
  logo?: string
  rpcEndpoint?: string
  restEndpoint?: string
  denom: string
  minimalDenom: string
  decimals: number
  gasPrice: string
  clearingFee?: string
}

// Chain configuration with public endpoints only
// Private endpoints are loaded from environment variables on the backend
export const CHAIN_CONFIG: Record<string, ChainInfo> = {
  'cosmoshub-4': {
    name: 'Cosmos Hub',
    prefix: 'cosmos',
    color: '#2E3148',
    logo: '/images/cosmos-logo.svg',
    rpcEndpoint: 'https://cosmos-rpc.polkachu.com',
    restEndpoint: 'https://cosmos-api.polkachu.com',
    denom: 'ATOM',
    minimalDenom: 'uatom',
    decimals: 6,
    gasPrice: '0.025',
    clearingFee: '0.1'
  },
  'osmosis-1': {
    name: 'Osmosis',
    prefix: 'osmo',
    color: '#5E12A0',
    logo: '/images/osmosis-logo.svg',
    rpcEndpoint: 'https://osmosis-rpc.polkachu.com',
    restEndpoint: 'https://osmosis-api.polkachu.com',
    denom: 'OSMO',
    minimalDenom: 'uosmo',
    decimals: 6,
    gasPrice: '0.025',
    clearingFee: '0.1'
  },
  'neutron-1': {
    name: 'Neutron',
    prefix: 'neutron',
    color: '#000000',
    logo: '/images/neutron-logo.svg',
    rpcEndpoint: 'https://neutron-rpc.polkachu.com',
    restEndpoint: 'https://neutron-api.polkachu.com',
    denom: 'NTRN',
    minimalDenom: 'untrn',
    decimals: 6,
    gasPrice: '0.025',
    clearingFee: '0.1'
  },
  'noble-1': {
    name: 'Noble',
    prefix: 'noble',
    color: '#5B21B6',
    logo: '/images/noble-logo.svg',
    rpcEndpoint: 'https://noble-rpc.polkachu.com',
    restEndpoint: 'https://noble-api.polkachu.com',
    denom: 'USDC',
    minimalDenom: 'uusdc',
    decimals: 6,
    gasPrice: '0.01',
    clearingFee: '0.1'
  },
  'terra2-1': {
    name: 'Terra',
    prefix: 'terra',
    color: '#0052FF',
    logo: '/images/terra-logo.svg',
    rpcEndpoint: 'https://rpc-terra.keplr.app',
    restEndpoint: 'https://api-terra.keplr.app',
    denom: 'LUNA',
    minimalDenom: 'uluna',
    decimals: 6,
    gasPrice: '0.015',
    clearingFee: '0.1'
  },
  'juno-1': {
    name: 'Juno',
    prefix: 'juno',
    color: '#F0827D',
    logo: '/images/juno-logo.svg',
    rpcEndpoint: 'https://rpc-juno.keplr.app',
    restEndpoint: 'https://api-juno.keplr.app',
    denom: 'JUNO',
    minimalDenom: 'ujuno',
    decimals: 6,
    gasPrice: '0.075',
    clearingFee: '0.1'
  },
  'akashnet-2': {
    name: 'Akash',
    prefix: 'akash',
    color: '#FF414C',
    logo: '/images/akash-logo.svg',
    rpcEndpoint: 'https://rpc-akash.keplr.app',
    restEndpoint: 'https://api-akash.keplr.app',
    denom: 'AKT',
    minimalDenom: 'uakt',
    decimals: 6,
    gasPrice: '0.025',
    clearingFee: '0.1'
  },
  'secret-4': {
    name: 'Secret Network',
    prefix: 'secret',
    color: '#1B1B1B',
    logo: '/images/secret-logo.svg',
    rpcEndpoint: 'https://rpc-secret.keplr.app',
    restEndpoint: 'https://api-secret.keplr.app',
    denom: 'SCRT',
    minimalDenom: 'uscrt',
    decimals: 6,
    gasPrice: '0.25',
    clearingFee: '0.1'
  },
  'evmos_9001-2': {
    name: 'Evmos',
    prefix: 'evmos',
    color: '#ED4E33',
    logo: '/images/evmos-logo.svg',
    rpcEndpoint: 'https://rpc-evmos.keplr.app',
    restEndpoint: 'https://api-evmos.keplr.app',
    denom: 'EVMOS',
    minimalDenom: 'aevmos',
    decimals: 18,
    gasPrice: '25000000000',
    clearingFee: '0.1'
  }
}

export type ChainId = keyof typeof CHAIN_CONFIG

// Channel connection mapping
export interface ChannelConnection {
  chain: string
  counterparty: string
  port?: string
  version?: string
  active?: boolean
}

export const CHANNEL_CONNECTIONS: Record<string, Record<string, ChannelConnection>> = {
  'cosmoshub-4': {
    'channel-0': { chain: 'osmosis-1', counterparty: 'channel-141', active: true },
    'channel-1': { chain: 'neutron-1', counterparty: 'channel-569', active: true },
    'channel-207': { chain: 'terra2-1', counterparty: 'channel-0', active: true },
    'channel-192': { chain: 'juno-1', counterparty: 'channel-0', active: true },
    'channel-184': { chain: 'akashnet-2', counterparty: 'channel-17', active: true },
    'channel-476': { chain: 'secret-4', counterparty: 'channel-0', active: true },
    'channel-292': { chain: 'evmos_9001-2', counterparty: 'channel-3', active: true }
  },
  'osmosis-1': {
    'channel-141': { chain: 'cosmoshub-4', counterparty: 'channel-0', active: true },
    'channel-874': { chain: 'neutron-1', counterparty: 'channel-10', active: true },
    'channel-108': { chain: 'terra2-1', counterparty: 'channel-1', active: true },
    'channel-75': { chain: 'juno-1', counterparty: 'channel-47', active: true },
    'channel-2': { chain: 'akashnet-2', counterparty: 'channel-9', active: true },
    'channel-199': { chain: 'secret-4', counterparty: 'channel-1', active: true },
    'channel-188': { chain: 'evmos_9001-2', counterparty: 'channel-0', active: true }
  },
  'neutron-1': {
    'channel-569': { chain: 'cosmoshub-4', counterparty: 'channel-1', active: true },
    'channel-10': { chain: 'osmosis-1', counterparty: 'channel-874', active: true },
    'channel-3': { chain: 'terra2-1', counterparty: 'channel-25', active: true },
    'channel-4': { chain: 'juno-1', counterparty: 'channel-124', active: true }
  },
  'terra2-1': {
    'channel-0': { chain: 'cosmoshub-4', counterparty: 'channel-207', active: true },
    'channel-1': { chain: 'osmosis-1', counterparty: 'channel-108', active: true },
    'channel-25': { chain: 'neutron-1', counterparty: 'channel-3', active: true },
    'channel-33': { chain: 'juno-1', counterparty: 'channel-86', active: true }
  },
  'juno-1': {
    'channel-0': { chain: 'cosmoshub-4', counterparty: 'channel-192', active: true },
    'channel-47': { chain: 'osmosis-1', counterparty: 'channel-75', active: true },
    'channel-124': { chain: 'neutron-1', counterparty: 'channel-4', active: true },
    'channel-86': { chain: 'terra2-1', counterparty: 'channel-33', active: true }
  }
}

// Utility functions
export const getChainName = (chainId: string): string => {
  return CHAIN_CONFIG[chainId]?.name || chainId
}

export const getChainColor = (chainId: string): string => {
  return CHAIN_CONFIG[chainId]?.color || '#6B7280'
}

export const getChainDenom = (chainId: string): string => {
  return CHAIN_CONFIG[chainId]?.denom || 'UNKNOWN'
}

export const getMinimalDenom = (chainId: string): string => {
  return CHAIN_CONFIG[chainId]?.minimalDenom || 'unknown'
}

export const getCounterpartyChain = (srcChain: string, srcChannel: string): string => {
  const connections = CHANNEL_CONNECTIONS[srcChain]
  if (!connections) return 'unknown'
  
  const connection = connections[srcChannel]
  return connection?.chain || 'unknown'
}

export const getCounterpartyChannel = (srcChain: string, srcChannel: string): string => {
  const connections = CHANNEL_CONNECTIONS[srcChain]
  if (!connections) return 'unknown'
  
  const connection = connections[srcChannel]
  return connection?.counterparty || 'unknown'
}

export const getChannelPair = (srcChain: string, srcChannel: string): string => {
  const dstChain = getCounterpartyChain(srcChain, srcChannel)
  const dstChannel = getCounterpartyChannel(srcChain, srcChannel)
  
  if (dstChain === 'unknown' || dstChannel === 'unknown') {
    return `${srcChain}/${srcChannel}`
  }
  
  return `${srcChain}/${srcChannel} ↔ ${dstChain}/${dstChannel}`
}

export const formatAmount = (amount: string | number, denom: string, chainId?: string): string => {
  const decimals = chainId ? CHAIN_CONFIG[chainId]?.decimals || 6 : 6
  const value = typeof amount === 'string' ? parseFloat(amount) : amount
  const formatted = (value / Math.pow(10, decimals)).toFixed(decimals)
  
  // Remove trailing zeros
  const trimmed = formatted.replace(/\.?0+$/, '')
  
  return `${trimmed} ${denom}`
}

export const parseAmount = (amount: string | number, chainId: string): bigint => {
  const decimals = CHAIN_CONFIG[chainId]?.decimals || 6
  const value = typeof amount === 'string' ? parseFloat(amount) : amount
  return BigInt(Math.floor(value * Math.pow(10, decimals)))
}

// Get all supported chains
export const getSupportedChains = (): ChainId[] => {
  return Object.keys(CHAIN_CONFIG) as ChainId[]
}

// Check if a chain is supported
export const isChainSupported = (chainId: string): boolean => {
  return chainId in CHAIN_CONFIG
}

// Get chain by prefix
export const getChainByPrefix = (prefix: string): ChainId | undefined => {
  const entry = Object.entries(CHAIN_CONFIG).find(([_, info]) => info.prefix === prefix)
  return entry?.[0] as ChainId | undefined
}

// Get all active channels for a chain
export const getActiveChannels = (chainId: string): string[] => {
  const connections = CHANNEL_CONNECTIONS[chainId]
  if (!connections) return []
  
  return Object.entries(connections)
    .filter(([_, conn]) => conn.active !== false)
    .map(([channel]) => channel)
}