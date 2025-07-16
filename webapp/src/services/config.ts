import { api } from './api'

export interface ChainConfig {
  chain_id: string
  chain_name: string
  address_prefix: string
  explorer: string
  logo?: string
  rpc_endpoint?: string
  api_endpoint?: string
  ws_endpoint?: string
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
  api_version: string
  clearingFees?: Record<string, string>
  denoms?: Record<string, string>
}

class ConfigService {
  private registry: ChainRegistry | null = null
  private loadPromise: Promise<ChainRegistry> | null = null

  async loadRegistry(): Promise<ChainRegistry> {
    // If already loaded, return cached
    if (this.registry) {
      return this.registry
    }

    // If currently loading, return existing promise
    if (this.loadPromise) {
      return this.loadPromise
    }

    // Start loading
    this.loadPromise = api.get<ChainRegistry>('/config')
      .then(response => {
        this.registry = response.data
        return this.registry
      })
      .catch(error => {
        console.error('Failed to load chain registry:', error)
        // Return fallback configuration
        return this.getFallbackRegistry()
      })

    return this.loadPromise
  }

  async getChain(chainId: string): Promise<ChainConfig | undefined> {
    const registry = await this.loadRegistry()
    return registry.chains[chainId]
  }

  async getChainByPrefix(prefix: string): Promise<ChainConfig | undefined> {
    const registry = await this.loadRegistry()
    return Object.values(registry.chains).find(chain => chain.address_prefix === prefix)
  }

  async getChannels(chainId?: string): Promise<ChannelConfig[]> {
    const registry = await this.loadRegistry()
    if (!chainId) {
      return registry.channels
    }
    return registry.channels.filter(
      channel => channel.source_chain === chainId || channel.dest_chain === chainId
    )
  }

  async getAllChains(): Promise<ChainConfig[]> {
    const registry = await this.loadRegistry()
    return Object.values(registry.chains)
  }

  getExplorerUrl(chainId: string, txHash: string): string {
    const chain = this.registry?.chains[chainId]
    if (!chain) {
      // Fallback to mintscan
      return `https://www.mintscan.io/${chainId}/txs/${txHash}`
    }
    return `${chain.explorer}/${txHash}`
  }

  async getRegistry(): Promise<ChainRegistry> {
    // Return the full registry
    return this.loadRegistry()
  }

  private getFallbackRegistry(): ChainRegistry {
    // Minimal fallback configuration
    return {
      chains: {
        'cosmoshub-4': {
          chain_id: 'cosmoshub-4',
          chain_name: 'Cosmos Hub',
          address_prefix: 'cosmos',
          explorer: 'https://www.mintscan.io/cosmos/txs',
          logo: '/images/chains/cosmos.svg'
        },
        'osmosis-1': {
          chain_id: 'osmosis-1',
          chain_name: 'Osmosis',
          address_prefix: 'osmo',
          explorer: 'https://www.mintscan.io/osmosis/txs',
          logo: '/images/chains/osmosis.svg'
        },
        'neutron-1': {
          chain_id: 'neutron-1',
          chain_name: 'Neutron',
          address_prefix: 'neutron',
          explorer: 'https://www.mintscan.io/neutron/txs',
          logo: '/images/chains/neutron.svg'
        },
        'noble-1': {
          chain_id: 'noble-1',
          chain_name: 'Noble',
          address_prefix: 'noble',
          explorer: 'https://www.mintscan.io/noble/txs',
          logo: '/images/chains/noble.svg'
        }
      },
      channels: [],
      api_version: '1.0'
    }
  }

  // Clear cache (useful for testing or when config might have changed)
  clearCache() {
    this.registry = null
    this.loadPromise = null
  }
}

export const configService = new ConfigService()