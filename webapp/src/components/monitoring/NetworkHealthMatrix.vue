<template>
  <div class="bg-white p-4 rounded-lg">
    <h3 class="text-lg font-medium mb-4">Network Health Status</h3>
    <div class="grid grid-cols-3 gap-4">
      <div v-for="chain in chains" :key="chain.chainId" 
        class="border rounded-lg p-4 text-center cursor-pointer hover:shadow-md transition-shadow"
        :class="getHealthClass(chain)">
        <h4 class="font-medium text-gray-900">{{ chain.chainName }}</h4>
        <p class="text-sm text-gray-600 mt-1">{{ chain.chainId }}</p>
        <p v-if="chain.chainId === 'neutron-1'" class="text-xs text-blue-600 mt-1" title="Neutron uses Slinky oracle vote extensions which require special handling">
          ⚠️ Limited monitoring
        </p>
        
        <!-- Health Indicators -->
        <div class="mt-3 space-y-1">
          <div class="flex justify-between text-xs">
            <span>Packets:</span>
            <span class="font-medium">{{ chain.totalPackets.toLocaleString() }}</span>
          </div>
          <div class="flex justify-between text-xs">
            <span>Errors:</span>
            <span class="font-medium" :class="chain.errors > 10 ? 'text-red-600' : ''">
              {{ chain.errors }}
            </span>
          </div>
          <div class="flex justify-between text-xs">
            <span>Reconnects:</span>
            <span class="font-medium">{{ chain.reconnects }}</span>
          </div>
        </div>
        
        <!-- Status Indicator -->
        <div class="mt-3 flex items-center justify-center">
          <div class="flex items-center space-x-2">
            <div :class="[
              'h-2 w-2 rounded-full',
              chain.status === 'connected' ? 'bg-green-400' : 
              chain.status === 'error' ? 'bg-red-400' : 'bg-yellow-400'
            ]"></div>
            <span class="text-xs capitalize">{{ chain.status }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ChainMetrics } from '@/types/monitoring'

interface Props {
  chains?: ChainMetrics[]
}

const props = defineProps<Props>()

function getHealthClass(chain: ChainMetrics): string {
  // Special case for Neutron - show as info/warning due to Slinky compatibility
  if (chain.chainId === 'neutron-1') return 'border-blue-300 bg-blue-50'
  
  if (chain.errors > 50) return 'border-red-300 bg-red-50'
  if (chain.errors > 20 || chain.reconnects > 10) return 'border-yellow-300 bg-yellow-50'
  return 'border-green-300 bg-green-50'
}
</script>