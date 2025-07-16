<template>
  <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:shadow-md transition-shadow">
    <!-- Chain Header -->
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2">
        <div class="w-10 h-10 rounded-full flex items-center justify-center text-white font-semibold"
             :style="{ backgroundColor: getChainColor(chain.chainId) }">
          {{ chain.chainName?.charAt(0) || chain.chainId.charAt(0).toUpperCase() }}
        </div>
        <div>
          <h3 class="font-medium text-gray-900">{{ chain.chainName || chain.chainId }}</h3>
          <p class="text-xs text-gray-500">{{ chain.chainId }}</p>
        </div>
      </div>
      <div class="flex items-center gap-1">
        <div class="w-2 h-2 rounded-full"
             :class="getStatusClass(chain.status)">
        </div>
        <span class="text-xs text-gray-600">{{ chain.status }}</span>
      </div>
    </div>

    <!-- Chain Metrics -->
    <div class="space-y-2">
      <div v-if="chain.height" class="flex justify-between items-center">
        <span class="text-sm text-gray-600">Height</span>
        <span class="text-sm font-medium">{{ formatNumber(chain.height) }}</span>
      </div>
      
      <div v-if="chain.totalTxs" class="flex justify-between items-center">
        <span class="text-sm text-gray-600">Total Txs</span>
        <span class="text-sm font-medium">{{ formatNumber(chain.totalTxs) }}</span>
      </div>
      
      <div class="flex justify-between items-center">
        <span class="text-sm text-gray-600">24h Packets</span>
        <span class="text-sm font-medium">{{ formatNumber(packets?.total || 0) }}</span>
      </div>
      
      <div class="flex justify-between items-center">
        <span class="text-sm text-gray-600">Success Rate</span>
        <span class="text-sm font-medium" 
              :class="getSuccessRateClass(packets?.successRate || 0)">
          {{ (packets?.successRate || 0).toFixed(1) }}%
        </span>
      </div>
      
      <div v-if="packets?.stuck && packets.stuck > 0" 
           class="flex justify-between items-center">
        <span class="text-sm text-gray-600">Stuck</span>
        <span class="text-sm font-medium text-orange-600">{{ packets.stuck }}</span>
      </div>
    </div>

    <!-- Chain Actions -->
    <div class="mt-3 pt-3 border-t border-gray-100">
      <button @click="$emit('view-details', chain)"
              class="text-xs text-blue-600 hover:text-blue-700 font-medium">
        View Details →
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { configService } from '@/services/config'

interface ChainInfo {
  chainId: string
  chainName?: string
  status: string
  height?: number
  totalTxs?: number
  totalPackets?: number
  reconnects?: number
  timeouts?: number
  errors?: number
  lastUpdate?: Date
}

interface PacketStats {
  total: number
  successful: number
  failed: number
  stuck: number
  successRate: number
}

const props = defineProps<{
  chain: ChainInfo
  packets?: PacketStats
}>()

const emit = defineEmits<{
  'view-details': [chain: ChainInfo]
}>()

function getChainColor(chainId: string): string {
  // Generate a consistent color based on chain ID hash
  let hash = 0
  for (let i = 0; i < chainId.length; i++) {
    const char = chainId.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32-bit integer
  }
  
  // Generate a hue based on the hash
  const hue = Math.abs(hash) % 360
  return `hsl(${hue}, 65%, 45%)`
}

function getStatusClass(status: string): string {
  switch (status?.toLowerCase()) {
    case 'connected':
      return 'bg-green-500'
    case 'syncing':
      return 'bg-yellow-500 animate-pulse'
    case 'disconnected':
      return 'bg-red-500'
    default:
      return 'bg-gray-400'
  }
}

function getSuccessRateClass(rate: number): string {
  if (rate >= 95) return 'text-green-600'
  if (rate >= 85) return 'text-yellow-600'
  return 'text-red-600'
}

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}
</script>