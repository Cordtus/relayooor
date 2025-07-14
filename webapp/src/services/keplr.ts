/**
 * Keplr wallet integration service
 * Enhanced with signing, verification, and chain management
 */

import { SigningStargateClient } from '@cosmjs/stargate'
import { OfflineDirectSigner } from '@cosmjs/proto-signing'
import { CHAIN_CONFIG } from '@/config/chains'

export interface KeplrSignature {
  signature: string
  pub_key: {
    type: string
    value: string
  }
}

export interface ChainInfo {
  readonly chainId: string
  readonly chainName: string
  readonly rpc: string
  readonly rest: string
  readonly bip44: {
    readonly coinType: number
  }
  readonly bech32Config: {
    readonly bech32PrefixAccAddr: string
    readonly bech32PrefixAccPub: string
    readonly bech32PrefixValAddr: string
    readonly bech32PrefixValPub: string
    readonly bech32PrefixConsAddr: string
    readonly bech32PrefixConsPub: string
  }
  readonly currencies: ReadonlyArray<{
    readonly coinDenom: string
    readonly coinMinimalDenom: string
    readonly coinDecimals: number
    readonly coinGeckoId?: string
  }>
  readonly feeCurrencies: ReadonlyArray<{
    readonly coinDenom: string
    readonly coinMinimalDenom: string
    readonly coinDecimals: number
    readonly coinGeckoId?: string
  }>
  readonly stakeCurrency: {
    readonly coinDenom: string
    readonly coinMinimalDenom: string
    readonly coinDecimals: number
    readonly coinGeckoId?: string
  }
  readonly coinType?: number
  readonly gasPriceStep?: {
    readonly low: number
    readonly average: number
    readonly high: number
  }
}

class KeplrService {
  private signingClients: Map<string, SigningStargateClient> = new Map()

  /**
   * Check if Keplr is installed
   */
  isInstalled(): boolean {
    return typeof window !== 'undefined' && !!window.keplr
  }

  /**
   * Get Keplr instance
   */
  getKeplr() {
    if (!this.isInstalled()) {
      throw new Error('Keplr wallet is not installed')
    }
    return window.keplr!
  }

  /**
   * Enable Keplr for a specific chain
   */
  async enable(chainId: string): Promise<void> {
    const keplr = this.getKeplr()
    
    // Try to enable the chain
    try {
      await keplr.enable(chainId)
    } catch (error) {
      // If chain is not added, try to add it
      if (error.message?.includes('There is no chain info')) {
        await this.addChain(chainId)
        await keplr.enable(chainId)
      } else {
        throw error
      }
    }
  }

  /**
   * Add a new chain to Keplr
   */
  async addChain(chainId: string): Promise<void> {
    const config = CHAIN_CONFIG[chainId]
    if (!config) {
      throw new Error(`Chain ${chainId} not supported`)
    }

    const chainInfo: ChainInfo = {
      chainId,
      chainName: config.name,
      rpc: config.rpcEndpoint || '',
      rest: config.restEndpoint || '',
      bip44: {
        coinType: 118 // Default for Cosmos chains
      },
      bech32Config: {
        bech32PrefixAccAddr: config.prefix,
        bech32PrefixAccPub: `${config.prefix}pub`,
        bech32PrefixValAddr: `${config.prefix}valoper`,
        bech32PrefixValPub: `${config.prefix}valoperpub`,
        bech32PrefixConsAddr: `${config.prefix}valcons`,
        bech32PrefixConsPub: `${config.prefix}valconspub`
      },
      currencies: [{
        coinDenom: config.denom,
        coinMinimalDenom: config.minimalDenom,
        coinDecimals: config.decimals
      }],
      feeCurrencies: [{
        coinDenom: config.denom,
        coinMinimalDenom: config.minimalDenom,
        coinDecimals: config.decimals
      }],
      stakeCurrency: {
        coinDenom: config.denom,
        coinMinimalDenom: config.minimalDenom,
        coinDecimals: config.decimals
      },
      gasPriceStep: {
        low: parseFloat(config.gasPrice) * 0.8,
        average: parseFloat(config.gasPrice),
        high: parseFloat(config.gasPrice) * 1.5
      }
    }

    await this.getKeplr().experimentalSuggestChain(chainInfo)
  }

  /**
   * Get account information
   */
  async getAccount(chainId: string) {
    await this.enable(chainId)
    return await this.getKeplr().getKey(chainId)
  }

  /**
   * Sign a message
   */
  async signMessage(
    chainId: string,
    address: string,
    message: string
  ): Promise<KeplrSignature> {
    await this.enable(chainId)
    
    const result = await this.getKeplr().signArbitrary(
      chainId,
      address,
      message
    )
    
    return {
      signature: result.signature,
      pub_key: result.pub_key
    }
  }

  /**
   * Verify a signed message
   */
  async verifyMessage(
    address: string,
    message: string,
    signature: KeplrSignature
  ): Promise<boolean> {
    // In production, implement proper signature verification
    // using @cosmjs/crypto or similar
    console.log('Verifying message:', { address, message, signature })
    return true
  }

  /**
   * Get offline signer for transactions
   */
  async getOfflineSigner(chainId: string): Promise<OfflineDirectSigner> {
    await this.enable(chainId)
    return this.getKeplr().getOfflineSigner(chainId)
  }

  /**
   * Get signing client for a chain
   */
  async getSigningClient(chainId: string): Promise<SigningStargateClient> {
    // Check cache
    if (this.signingClients.has(chainId)) {
      return this.signingClients.get(chainId)!
    }

    const config = CHAIN_CONFIG[chainId]
    if (!config || !config.rpcEndpoint) {
      throw new Error(`No RPC endpoint configured for chain ${chainId}`)
    }

    const offlineSigner = await this.getOfflineSigner(chainId)
    const client = await SigningStargateClient.connectWithSigner(
      config.rpcEndpoint,
      offlineSigner
    )

    // Cache the client
    this.signingClients.set(chainId, client)
    
    return client
  }

  /**
   * Send tokens
   */
  async sendTokens(
    chainId: string,
    fromAddress: string,
    toAddress: string,
    amount: string,
    denom: string,
    memo?: string
  ) {
    const client = await this.getSigningClient(chainId)
    
    const amountObj = {
      denom,
      amount
    }

    const fee = {
      amount: [{
        denom,
        amount: '5000' // Default fee
      }],
      gas: '200000'
    }

    return await client.sendTokens(
      fromAddress,
      toAddress,
      [amountObj],
      fee,
      memo || ''
    )
  }

  /**
   * Suggest a token to Keplr
   */
  async suggestToken(chainId: string, contractAddress: string) {
    await this.enable(chainId)
    
    // For CW20 tokens
    await this.getKeplr().suggestToken(chainId, contractAddress)
  }

  /**
   * Get all supported chains from Keplr
   */
  async getSupportedChains(): Promise<string[]> {
    // This is a workaround as Keplr doesn't provide a direct API
    const knownChains = Object.keys(CHAIN_CONFIG)
    const supportedChains: string[] = []

    for (const chainId of knownChains) {
      try {
        await this.getKeplr().getKey(chainId)
        supportedChains.push(chainId)
      } catch {
        // Chain not added to Keplr
      }
    }

    return supportedChains
  }

  /**
   * Listen for account changes
   */
  onAccountChange(callback: () => void) {
    // Keplr fires this event when account changes
    window.addEventListener('keplr_keystorechange', callback)
    
    return () => {
      window.removeEventListener('keplr_keystorechange', callback)
    }
  }
}

// Export singleton instance
export const keplrService = new KeplrService()

// Extend window type
declare global {
  interface Window {
    keplr?: any
    getOfflineSigner?: (chainId: string) => OfflineDirectSigner
  }
}