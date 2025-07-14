<template>
  <button
    @click="handleConnect"
    class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
    :disabled="connecting"
  >
    <Wallet class="w-4 h-4 mr-2" />
    <span v-if="connecting">Connecting...</span>
    <span v-else>{{ wallet.isConnected ? formatAddress(wallet.address) : 'Connect Wallet' }}</span>
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Wallet } from 'lucide-vue-next'
import { useWalletStore } from '@/stores/wallet'
import { keplrService } from '@/services/keplr'
import { toast } from 'vue-sonner'

const wallet = useWalletStore()
const connecting = ref(false)

async function handleConnect() {
  if (connecting.value) return
  
  try {
    connecting.value = true
    
    if (!wallet.isConnected) {
      // Check if Keplr is installed
      if (!keplrService.isInstalled()) {
        toast.error('Please install Keplr wallet extension')
        return
      }
      
      // Enable chains
      const chainIds = ['osmosis-1', 'cosmoshub-4']
      for (const chainId of chainIds) {
        try {
          await keplrService.enable(chainId)
        } catch (err) {
          console.warn(`Failed to enable ${chainId}:`, err)
        }
      }
      
      // Get account info from the first available chain
      let accountInfo = null
      for (const chainId of chainIds) {
        try {
          const key = await keplrService.getAccount(chainId)
          accountInfo = {
            address: key.bech32Address,
            name: key.name,
            algo: key.algo,
            pubKey: key.pubKey
          }
          
          // Store in wallet store
          await wallet.connect(chainId)
          
          toast.success(`Connected to ${accountInfo.name}`)
          break
        } catch (err) {
          console.warn(`Failed to get key for ${chainId}:`, err)
        }
      }
      
      if (!accountInfo) {
        throw new Error('Failed to connect to any supported chain')
      }
      
      // Listen for account changes
      window.addEventListener('keplr_keystorechange', handleAccountChange)
      
    } else {
      // Disconnect
      await wallet.disconnect()
      window.removeEventListener('keplr_keystorechange', handleAccountChange)
      toast.info('Wallet disconnected')
    }
  } catch (error) {
    console.error('Wallet connection error:', error)
    toast.error(error instanceof Error ? error.message : 'Failed to connect wallet')
  } finally {
    connecting.value = false
  }
}

async function handleAccountChange() {
  // Handle account change event
  if (wallet.isConnected) {
    try {
      await wallet.refreshAccount()
      toast.info('Account changed')
    } catch (error) {
      console.error('Error handling account change:', error)
    }
  }
}

function formatAddress(addr: string | null): string {
  if (!addr) return ''
  if (addr.length < 15) return addr
  return `${addr.slice(0, 10)}...${addr.slice(-4)}`
}

// Cleanup on unmount
import { onUnmounted } from 'vue'
onUnmounted(() => {
  window.removeEventListener('keplr_keystorechange', handleAccountChange)
})
</script>