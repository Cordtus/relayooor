<template>
  <div class="space-y-6">
    <!-- Info Banner -->
    <div class="bg-blue-50 border-l-4 border-blue-400 p-4">
      <div class="flex">
        <div class="flex-shrink-0">
          <Info class="h-5 w-5 text-blue-400" />
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-blue-800">Secure Packet Clearing Service</h3>
          <p class="mt-2 text-sm text-blue-700">
            Clear your stuck IBC transfers through our secure relayer service. Simply select the packets you want to clear, 
            pay the service fee, and we'll handle the rest. All transactions are verified on-chain for maximum security.
          </p>
        </div>
      </div>
    </div>

    <!-- Wallet Connection Check -->
    <div v-if="!walletStore.isConnected" class="bg-white shadow rounded-lg p-8 text-center">
      <Wallet class="h-16 w-16 text-gray-400 mx-auto mb-4" />
      <h2 class="text-xl font-semibold text-gray-900 mb-2">Connect Your Wallet</h2>
      <p class="text-gray-600 mb-6">
        Connect your Keplr wallet to view and clear your stuck IBC transfers
      </p>
      <WalletConnect />
    </div>

    <!-- Connected View -->
    <div v-else>
      <!-- User Statistics -->
      <UserStatistics v-if="hasAuthenticated" class="mb-6" />

      <!-- Clearing Wizard -->
      <ClearingWizard />

      <!-- Platform Statistics -->
      <div class="mt-8">
        <h2 class="text-lg font-semibold mb-4">Platform Statistics</h2>
        <PlatformStats />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Info, Wallet } from 'lucide-vue-next'
import { useWalletStore } from '@/stores/wallet'
import { clearingService } from '@/services/clearing'
import WalletConnect from '@/components/WalletConnect.vue'
import ClearingWizard from '@/components/clearing/ClearingWizard.vue'
import UserStatistics from '@/components/clearing/UserStatistics.vue'
import PlatformStats from '@/components/clearing/PlatformStats.vue'

const walletStore = useWalletStore()
const hasAuthenticated = ref(false)

// Check if user has authenticated session
onMounted(async () => {
  const sessionToken = localStorage.getItem('clearing_session_token')
  if (sessionToken && walletStore.isConnected) {
    hasAuthenticated.value = true
  }
})

// Watch for wallet connection to prompt authentication
watch(() => walletStore.isConnected, async (connected) => {
  if (connected && !hasAuthenticated.value) {
    // Prompt for signature authentication
    try {
      const message = clearingService.generateAuthMessage(walletStore.address!)
      // In production, would prompt wallet for signature
      // For now, mark as authenticated
      hasAuthenticated.value = true
    } catch (error) {
      console.error('Authentication failed:', error)
    }
  }
})

import { watch } from 'vue'
</script>