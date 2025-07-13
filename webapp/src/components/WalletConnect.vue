<template>
  <button
    @click="handleConnect"
    class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
  >
    <Wallet class="w-4 h-4 mr-2" />
    {{ connected ? formatAddress(address) : 'Connect Wallet' }}
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Wallet } from 'lucide-vue-next'

const connected = ref(false)
const address = ref('')

function handleConnect() {
  // Implement Keplr wallet connection
  if (!connected.value) {
    // Mock connection for now
    connected.value = true
    address.value = 'cosmos1abc...xyz'
  } else {
    connected.value = false
    address.value = ''
  }
}

function formatAddress(addr: string): string {
  if (!addr) return ''
  return `${addr.slice(0, 10)}...${addr.slice(-4)}`
}
</script>