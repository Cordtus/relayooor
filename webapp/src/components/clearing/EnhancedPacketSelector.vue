<template>
  <div class="enhanced-packet-selector">
    <!-- Summary stats -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="bg-blue-50 rounded-lg p-4">
        <p class="text-sm text-blue-600 font-medium">Total Transfers</p>
        <p class="text-2xl font-bold text-blue-900">{{ allTransfers.length }}</p>
      </div>
      <div class="bg-orange-50 rounded-lg p-4">
        <p class="text-sm text-orange-600 font-medium">Stuck Packets</p>
        <p class="text-2xl font-bold text-orange-900">{{ stuckPackets.length }}</p>
      </div>
      <div class="bg-purple-50 rounded-lg p-4">
        <p class="text-sm text-purple-600 font-medium">Selected</p>
        <p class="text-2xl font-bold text-purple-900">{{ selected.length }}</p>
      </div>
      <div class="bg-green-50 rounded-lg p-4">
        <p class="text-sm text-green-600 font-medium">Total Value</p>
        <p class="text-lg font-bold text-green-900">{{ formatTotalValue() }}</p>
      </div>
    </div>

    <!-- Filters and Controls -->
    <div class="bg-white border border-gray-200 rounded-lg p-4 mb-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <!-- View Toggle -->
          <div class="flex items-center bg-gray-100 rounded-md p-1">
            <button
              @click="showAllTransfers = false"
              :class="[
                'px-3 py-1 text-sm font-medium rounded transition-colors',
                !showAllTransfers 
                  ? 'bg-white text-gray-900 shadow-sm' 
                  : 'text-gray-600 hover:text-gray-900'
              ]"
            >
              Stuck Only
            </button>
            <button
              @click="showAllTransfers = true"
              :class="[
                'px-3 py-1 text-sm font-medium rounded transition-colors',
                showAllTransfers 
                  ? 'bg-white text-gray-900 shadow-sm' 
                  : 'text-gray-600 hover:text-gray-900'
              ]"
            >
              All Transfers
            </button>
          </div>

          <!-- Quick Actions -->
          <div class="flex items-center gap-2 border-l pl-4">
            <button
              @click="selectAllVisible"
              class="text-sm text-blue-600 hover:text-blue-700 font-medium"
            >
              Select All
            </button>
            <button
              @click="selectNone"
              class="text-sm text-gray-600 hover:text-gray-700 font-medium"
            >
              Clear Selection
            </button>
            <button
              v-if="showAllTransfers"
              @click="selectStuckOnly"
              class="text-sm text-orange-600 hover:text-orange-700 font-medium"
            >
              Select Stuck Only
            </button>
          </div>
        </div>

        <!-- Sort and Filter -->
        <div class="flex items-center gap-4">
          <!-- Search -->
          <div class="relative">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search by address, hash..."
              class="pl-8 pr-3 py-1.5 text-sm border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
            <Search class="absolute left-2 top-2 h-4 w-4 text-gray-400" />
          </div>

          <!-- Sort -->
          <select 
            v-model="sortBy"
            class="text-sm border border-gray-300 rounded-md px-3 py-1.5"
          >
            <option value="age">Sort by Age</option>
            <option value="value">Sort by Value</option>
            <option value="attempts">Sort by Attempts</option>
            <option value="status">Sort by Status</option>
          </select>
        </div>
      </div>

      <!-- Selected Summary -->
      <div v-if="selected.length > 0" class="mt-3 pt-3 border-t">
        <div class="flex items-center justify-between text-sm">
          <span class="text-gray-600">
            {{ selected.length }} transfer{{ selected.length !== 1 ? 's' : '' }} selected
          </span>
          <span class="font-medium text-gray-900">
            Estimated fee: {{ estimatedFee }}
          </span>
        </div>
      </div>
    </div>

    <!-- Transfer list -->
    <div class="space-y-2 max-h-[600px] overflow-y-auto">
      <TransferCard
        v-for="transfer in filteredAndSorted"
        :key="transfer.id"
        :transfer="transfer"
        :selected="isSelected(transfer)"
        :show-clear-button="transfer.status === 'stuck'"
        @toggle="toggleTransfer(transfer)"
        @clear-single="clearSingle(transfer)"
      />
    </div>

    <!-- Empty state -->
    <div v-if="filteredAndSorted.length === 0" class="text-center py-12 bg-gray-50 rounded-lg">
      <PackageIcon class="w-12 h-12 text-gray-400 mx-auto mb-3" />
      <p class="text-gray-600 font-medium">
        {{ searchQuery ? 'No transfers match your search' : 'No transfers found' }}
      </p>
      <p class="text-sm text-gray-500 mt-1">
        {{ showAllTransfers ? 'Try adjusting your filters' : 'All your transfers are flowing smoothly!' }}
      </p>
    </div>

    <!-- Bulk Actions Bar -->
    <div 
      v-if="selected.length > 0"
      class="fixed bottom-4 left-1/2 transform -translate-x-1/2 bg-white border border-gray-200 rounded-lg shadow-lg p-4 flex items-center gap-4"
    >
      <span class="text-sm text-gray-600">
        {{ selected.length }} selected
      </span>
      <button
        @click="clearSelected"
        class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        Clear Selected Packets
      </button>
      <button
        @click="selectNone"
        class="px-4 py-2 text-gray-700 text-sm font-medium hover:text-gray-900"
      >
        Cancel
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Package as PackageIcon, Search } from 'lucide-vue-next'
import TransferCard from './TransferCard.vue'
import { useWalletStore } from '@/stores/wallet'
import { clearingService } from '@/services/clearing'
import { formatAmount } from '@/config/chains'
import type { Transfer, StuckPacket } from '@/types/clearing'

const props = defineProps<{
  allTransfers: Transfer[]
  stuckPackets: StuckPacket[]
  selected: (Transfer | StuckPacket)[]
}>()

const emit = defineEmits<{
  'update:selected': [items: (Transfer | StuckPacket)[]]
  'clear-single': [item: Transfer | StuckPacket]
  'clear-selected': []
}>()

const walletStore = useWalletStore()

// State
const showAllTransfers = ref(false)
const searchQuery = ref('')
const sortBy = ref<'age' | 'value' | 'attempts' | 'status'>('age')

// Computed
const visibleTransfers = computed(() => {
  if (showAllTransfers.value) {
    return props.allTransfers
  }
  return props.stuckPackets
})

const filteredAndSorted = computed(() => {
  let transfers = [...visibleTransfers.value]
  
  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    transfers = transfers.filter(t => 
      t.sender?.toLowerCase().includes(query) ||
      t.receiver?.toLowerCase().includes(query) ||
      t.txHash?.toLowerCase().includes(query) ||
      t.sequence?.toString().includes(query)
    )
  }
  
  // Apply sorting
  switch (sortBy.value) {
    case 'age':
      transfers.sort((a, b) => (b.timestamp || 0) - (a.timestamp || 0))
      break
    case 'value':
      transfers.sort((a, b) => {
        const aValue = BigInt(a.amount || '0')
        const bValue = BigInt(b.amount || '0')
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0
      })
      break
    case 'attempts':
      transfers.sort((a, b) => (b.attempts || 0) - (a.attempts || 0))
      break
    case 'status':
      transfers.sort((a, b) => {
        const statusOrder = { stuck: 0, pending: 1, completed: 2 }
        return (statusOrder[a.status] || 99) - (statusOrder[b.status] || 99)
      })
      break
  }
  
  return transfers
})

const estimatedFee = computed(() => {
  if (props.selected.length === 0) return '0'
  
  const baseFee = 0.1 // Base fee per operation
  const perPacketFee = 0.01 // Fee per packet
  const total = baseFee + (props.selected.length * perPacketFee)
  
  // Get the most common denom from selected
  const denoms = props.selected.map(p => p.denom).filter(Boolean)
  const denom = denoms[0] || 'ATOM'
  
  return `${total.toFixed(2)} ${denom}`
})

// Methods
const isSelected = (transfer: Transfer | StuckPacket): boolean => {
  return props.selected.some(s => s.id === transfer.id)
}

const toggleTransfer = (transfer: Transfer | StuckPacket) => {
  const newSelection = isSelected(transfer)
    ? props.selected.filter(s => s.id !== transfer.id)
    : [...props.selected, transfer]
  
  emit('update:selected', newSelection)
}

const selectAllVisible = () => {
  emit('update:selected', [...filteredAndSorted.value])
}

const selectNone = () => {
  emit('update:selected', [])
}

const selectStuckOnly = () => {
  const stuckOnly = filteredAndSorted.value.filter(t => t.status === 'stuck')
  emit('update:selected', stuckOnly)
}

const clearSingle = (transfer: Transfer | StuckPacket) => {
  emit('clear-single', transfer)
}

const clearSelected = () => {
  emit('clear-selected')
}

const formatTotalValue = (): string => {
  if (props.selected.length === 0) return '0'
  
  // Group by denom
  const byDenom = props.selected.reduce((acc, item) => {
    const denom = item.denom || 'unknown'
    if (!acc[denom]) {
      acc[denom] = BigInt(0)
    }
    acc[denom] += BigInt(item.amount || '0')
    return acc
  }, {} as Record<string, bigint>)
  
  // Format each denom
  const formatted = Object.entries(byDenom).map(([denom, amount]) => {
    return formatAmount(amount.toString(), denom)
  })
  
  // Join all amounts
  return formatted.join(' + ')
}
</script>

<style scoped>
.enhanced-packet-selector {
  @apply relative;
}

/* Custom scrollbar for transfer list */
.max-h-\[600px\]::-webkit-scrollbar {
  width: 8px;
}

.max-h-\[600px\]::-webkit-scrollbar-track {
  @apply bg-gray-100 rounded;
}

.max-h-\[600px\]::-webkit-scrollbar-thumb {
  @apply bg-gray-400 rounded hover:bg-gray-500;
}
</style>