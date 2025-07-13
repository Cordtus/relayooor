import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface WalletState {
  address: string | null
  chainId: string | null
  isConnected: boolean
}

export const useWalletStore = defineStore('wallet', () => {
  // State
  const address = ref<string | null>(null)
  const chainId = ref<string | null>(null)
  const isConnected = ref(false)

  // Getters
  const shortAddress = computed(() => {
    if (!address.value) return ''
    return `${address.value.slice(0, 10)}...${address.value.slice(-8)}`
  })

  const chainName = computed(() => {
    const chains: Record<string, string> = {
      'osmosis-1': 'Osmosis',
      'cosmoshub-4': 'Cosmos Hub',
      'neutron-1': 'Neutron',
    }
    return chainId.value ? chains[chainId.value] || chainId.value : ''
  })

  // Actions
  async function connect() {
    // In production, this would interact with Keplr
    // For development, simulate connection
    if (window.keplr) {
      try {
        await window.keplr.enable(chainId.value || 'osmosis-1')
        const key = await window.keplr.getKey(chainId.value || 'osmosis-1')
        
        address.value = key.bech32Address
        isConnected.value = true
      } catch (error) {
        console.error('Failed to connect wallet:', error)
        throw error
      }
    } else {
      // Mock connection for development
      address.value = 'osmo1abc123def456ghi789jkl012mno345pqr678stu'
      chainId.value = 'osmosis-1'
      isConnected.value = true
    }
  }

  function disconnect() {
    address.value = null
    chainId.value = null
    isConnected.value = false
  }

  async function switchChain(newChainId: string) {
    if (!window.keplr) {
      chainId.value = newChainId
      return
    }

    try {
      await window.keplr.enable(newChainId)
      const key = await window.keplr.getKey(newChainId)
      
      address.value = key.bech32Address
      chainId.value = newChainId
    } catch (error) {
      console.error('Failed to switch chain:', error)
      throw error
    }
  }

  async function signMessage(message: string): Promise<string> {
    if (!window.keplr || !address.value || !chainId.value) {
      throw new Error('Wallet not connected')
    }

    try {
      const signature = await window.keplr.signArbitrary(
        chainId.value,
        address.value,
        message
      )
      
      return signature.signature
    } catch (error) {
      console.error('Failed to sign message:', error)
      throw error
    }
  }

  return {
    // State
    address,
    chainId,
    isConnected,
    
    // Getters
    shortAddress,
    chainName,
    
    // Actions
    connect,
    disconnect,
    switchChain,
    signMessage,
  }
})

// Extend window type for Keplr
declare global {
  interface Window {
    keplr?: any
  }
}