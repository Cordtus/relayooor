import { defineStore } from 'pinia'
import { ref, computed, onMounted, watch } from 'vue'
import { keplrService } from '@/services/keplr'
import { getChainName } from '@/config/chains'

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
  const isKeplrInstalled = ref(false)
  const accountChangeListener = ref<(() => void) | null>(null)

  // Getters
  const shortAddress = computed(() => {
    if (!address.value) return ''
    return `${address.value.slice(0, 10)}...${address.value.slice(-8)}`
  })

  const chainName = computed(() => {
    return chainId.value ? getChainName(chainId.value) : ''
  })

  // Actions
  async function connect(preferredChainId?: string) {
    try {
      const targetChainId = preferredChainId || chainId.value || 'osmosis-1'
      
      if (keplrService.isInstalled()) {
        await keplrService.enable(targetChainId)
        const key = await keplrService.getAccount(targetChainId)
        
        address.value = key.bech32Address
        chainId.value = targetChainId
        isConnected.value = true
        
        // Set up account change listener
        if (!accountChangeListener.value) {
          accountChangeListener.value = keplrService.onAccountChange(() => {
            // Refresh account info when changed
            refreshAccount()
          })
        }
      } else {
        throw new Error('Please install Keplr wallet extension')
      }
    } catch (error) {
      console.error('Failed to connect wallet:', error)
      throw error
    }
  }

  function disconnect() {
    address.value = null
    chainId.value = null
    isConnected.value = false
    
    // Remove account change listener
    if (accountChangeListener.value) {
      accountChangeListener.value()
      accountChangeListener.value = null
    }
  }

  async function switchChain(newChainId: string) {
    if (!isConnected.value) {
      throw new Error('Wallet not connected')
    }

    try {
      await keplrService.enable(newChainId)
      const key = await keplrService.getAccount(newChainId)
      
      address.value = key.bech32Address
      chainId.value = newChainId
    } catch (error) {
      console.error('Failed to switch chain:', error)
      throw error
    }
  }

  async function signMessage(message: string): Promise<string> {
    if (!isConnected.value || !address.value || !chainId.value) {
      throw new Error('Wallet not connected')
    }

    try {
      const result = await keplrService.signMessage(
        chainId.value,
        address.value,
        message
      )
      
      return result.signature
    } catch (error) {
      console.error('Failed to sign message:', error)
      throw error
    }
  }

  // Helper functions
  async function refreshAccount() {
    if (!isConnected.value || !chainId.value) return
    
    try {
      const key = await keplrService.getAccount(chainId.value)
      address.value = key.bech32Address
    } catch (error) {
      console.error('Failed to refresh account:', error)
    }
  }
  
  async function getSupportedChains(): Promise<string[]> {
    try {
      return await keplrService.getSupportedChains()
    } catch (error) {
      console.error('Failed to get supported chains:', error)
      return []
    }
  }
  
  async function sendTokens(
    toAddress: string,
    amount: string,
    denom: string,
    memo?: string
  ) {
    if (!isConnected.value || !address.value || !chainId.value) {
      throw new Error('Wallet not connected')
    }
    
    try {
      return await keplrService.sendTokens(
        chainId.value,
        address.value,
        toAddress,
        amount,
        denom,
        memo
      )
    } catch (error) {
      console.error('Failed to send tokens:', error)
      throw error
    }
  }
  
  async function verifySignature(
    message: string,
    signature: any
  ): Promise<boolean> {
    if (!address.value) {
      throw new Error('No address to verify against')
    }
    
    try {
      return await keplrService.verifyMessage(address.value, message, signature)
    } catch (error) {
      console.error('Failed to verify signature:', error)
      throw error
    }
  }
  
  // Initialize
  onMounted(() => {
    isKeplrInstalled.value = keplrService.isInstalled()
    
    // Auto-connect if previously connected
    const savedConnection = localStorage.getItem('wallet_connection')
    if (savedConnection && isKeplrInstalled.value) {
      const { chainId: savedChainId } = JSON.parse(savedConnection)
      connect(savedChainId).catch(console.error)
    }
  })
  
  // Watch for connection changes to save state
  watch(isConnected, (connected) => {
    if (connected && chainId.value) {
      localStorage.setItem('wallet_connection', JSON.stringify({
        chainId: chainId.value,
        timestamp: Date.now()
      }))
    } else {
      localStorage.removeItem('wallet_connection')
    }
  })

  return {
    // State
    address,
    chainId,
    isConnected,
    isKeplrInstalled,
    
    // Getters
    shortAddress,
    chainName,
    
    // Actions
    connect,
    disconnect,
    switchChain,
    signMessage,
    refreshAccount,
    getSupportedChains,
    sendTokens,
    verifySignature,
  }
})

// Extend window type for Keplr
declare global {
  interface Window {
    keplr?: any
  }
}