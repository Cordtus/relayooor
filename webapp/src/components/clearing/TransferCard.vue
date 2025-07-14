<template>
  <div
    @click="$emit('toggle')"
    :class="[
      'border rounded-lg p-4 cursor-pointer transition-all duration-200',
      selected 
        ? 'border-blue-500 bg-blue-50 shadow-sm' 
        : 'border-gray-200 hover:border-gray-300 bg-white'
    ]"
  >
    <div class="flex items-start justify-between">
      <!-- Left side - Checkbox and details -->
      <div class="flex items-start gap-3">
        <input
          type="checkbox"
          :checked="selected"
          @click.stop
          @change="$emit('toggle')"
          class="mt-1 h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-2 focus:ring-blue-500"
        >
        
        <div class="flex-1">
          <!-- Header -->
          <div class="flex items-center gap-3 mb-2">
            <span class="font-mono text-sm font-medium text-gray-900">
              #{{ transfer.sequence }}
            </span>
            
            <StatusBadge :status="transfer.status" />
            
            <span class="text-xs text-gray-500">
              {{ getChannelPair(transfer.srcChain, transfer.srcChannel) }}
            </span>
          </div>
          
          <!-- Addresses -->
          <div class="space-y-1 text-sm">
            <div class="flex items-center gap-2">
              <span class="text-gray-500 w-12">From:</span>
              <span class="font-mono text-gray-700">{{ formatAddress(transfer.sender) }}</span>
              <CopyButton :text="transfer.sender" />
            </div>
            <div class="flex items-center gap-2">
              <span class="text-gray-500 w-12">To:</span>
              <span class="font-mono text-gray-700">{{ formatAddress(transfer.receiver) }}</span>
              <CopyButton :text="transfer.receiver" />
            </div>
          </div>
          
          <!-- Transaction Hash -->
          <div v-if="transfer.txHash" class="mt-2 flex items-center gap-2">
            <span class="text-xs text-gray-500">Tx:</span>
            <a
              :href="getExplorerUrl(transfer.srcChain, transfer.txHash)"
              target="_blank"
              rel="noopener noreferrer"
              @click.stop
              class="font-mono text-xs text-blue-600 hover:text-blue-700 hover:underline"
            >
              {{ formatHash(transfer.txHash) }}
              <ExternalLink class="inline-block w-3 h-3 ml-1" />
            </a>
          </div>
        </div>
      </div>

      <!-- Right side - Amount and actions -->
      <div class="text-right">
        <!-- Amount -->
        <div class="mb-2">
          <p class="font-semibold text-gray-900">
            {{ formatAmount(transfer.amount, transfer.denom, transfer.srcChain) }}
          </p>
          <p v-if="transfer.usdValue" class="text-xs text-gray-500">
            ≈ ${{ transfer.usdValue.toFixed(2) }}
          </p>
        </div>
        
        <!-- Time info -->
        <div class="space-y-1 text-xs text-gray-500">
          <p>{{ formatTimeAgo(transfer.timestamp) }}</p>
          <p v-if="transfer.status === 'stuck'" class="text-orange-600 font-medium">
            Stuck for {{ formatDuration(Date.now() - transfer.timestamp) }}
          </p>
          <p v-if="transfer.attempts > 1" class="text-red-600">
            {{ transfer.attempts }} failed attempts
          </p>
        </div>
        
        <!-- Clear button -->
        <button
          v-if="showClearButton && transfer.status === 'stuck'"
          @click.stop="$emit('clear-single')"
          class="mt-2 px-3 py-1 text-xs font-medium text-white bg-orange-600 hover:bg-orange-700 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-orange-500"
        >
          Clear Now
        </button>
      </div>
    </div>
    
    <!-- Additional info on selection -->
    <div v-if="selected && transfer.memo" class="mt-3 pt-3 border-t border-gray-200">
      <p class="text-xs text-gray-600">
        <span class="font-medium">Memo:</span> {{ transfer.memo }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ExternalLink } from 'lucide-vue-next'
import StatusBadge from './StatusBadge.vue'
import CopyButton from '@/components/CopyButton.vue'
import { getChannelPair, formatAmount } from '@/config/chains'
import { formatDistanceToNow } from 'date-fns'
import type { Transfer, StuckPacket } from '@/types/clearing'

const props = defineProps<{
  transfer: Transfer | StuckPacket
  selected: boolean
  showClearButton?: boolean
}>()

defineEmits<{
  toggle: []
  'clear-single': []
}>()

// Format functions
const formatAddress = (address: string): string => {
  if (!address) return 'Unknown'
  if (address.length <= 20) return address
  return `${address.slice(0, 10)}...${address.slice(-8)}`
}

const formatHash = (hash: string): string => {
  if (!hash) return ''
  return `${hash.slice(0, 8)}...${hash.slice(-6)}`
}

const formatTimeAgo = (timestamp: number): string => {
  if (!timestamp) return 'Unknown time'
  return formatDistanceToNow(new Date(timestamp), { addSuffix: true })
}

const formatDuration = (ms: number): string => {
  const hours = Math.floor(ms / (1000 * 60 * 60))
  const minutes = Math.floor((ms % (1000 * 60 * 60)) / (1000 * 60))
  
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

const getExplorerUrl = (chain: string, txHash: string): string => {
  const explorers: Record<string, string> = {
    'cosmoshub-4': 'https://www.mintscan.io/cosmos/txs/',
    'osmosis-1': 'https://www.mintscan.io/osmosis/txs/',
    'neutron-1': 'https://www.mintscan.io/neutron/txs/',
    'terra2-1': 'https://www.mintscan.io/terra/txs/',
    'juno-1': 'https://www.mintscan.io/juno/txs/'
  }
  
  const baseUrl = explorers[chain] || 'https://www.mintscan.io/cosmos/txs/'
  return `${baseUrl}${txHash}`
}
</script>