<template>
  <div class="packet-search bg-white/10 backdrop-blur-md rounded-xl p-6 shadow-lg">
    <h2 class="text-xl font-semibold mb-6 text-white flex items-center gap-2">
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      Packet Search
    </h2>

    <!-- Search Form -->
    <div class="space-y-4">
      <!-- Search Type Selection -->
      <div class="flex gap-2 mb-4">
        <button
          v-for="type in searchTypes"
          :key="type.value"
          @click="searchType = type.value as 'wallet' | 'chain' | 'advanced'"
          :class="[
            'px-4 py-2 rounded-lg transition-all',
            searchType === type.value
              ? 'bg-blue-500 text-white'
              : 'bg-white/10 text-gray-300 hover:bg-white/20'
          ]"
        >
          {{ type.label }}
        </button>
      </div>

      <!-- Wallet Address Search -->
      <div v-if="searchType === 'wallet'" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">
            Wallet Address
          </label>
          <input
            v-model="walletAddress"
            type="text"
            placeholder="cosmos1... or osmo1..."
            class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                   text-white placeholder-gray-400 focus:outline-none focus:border-blue-400"
            @keyup.enter="performSearch"
          />
        </div>
        
        <div class="flex gap-2">
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input v-model="searchAsSender" type="checkbox" class="rounded" />
            As Sender
          </label>
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input v-model="searchAsReceiver" type="checkbox" class="rounded" />
            As Receiver
          </label>
        </div>
      </div>

      <!-- Chain Search -->
      <div v-if="searchType === 'chain'" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">
            Source Chain
          </label>
          <select
            v-model="sourceChain"
            class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                   text-white focus:outline-none focus:border-blue-400"
          >
            <option value="">All Chains</option>
            <option v-for="chain in availableChains" :key="chain.id" :value="chain.id">
              {{ chain.name }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">
            Destination Chain
          </label>
          <select
            v-model="destinationChain"
            class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                   text-white focus:outline-none focus:border-blue-400"
          >
            <option value="">All Chains</option>
            <option v-for="chain in availableChains" :key="chain.id" :value="chain.id">
              {{ chain.name }}
            </option>
          </select>
        </div>
      </div>

      <!-- Advanced Filters -->
      <div v-if="searchType === 'advanced'" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">
              Min Amount
            </label>
            <input
              v-model="minAmount"
              type="number"
              placeholder="0"
              class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                     text-white placeholder-gray-400 focus:outline-none focus:border-blue-400"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">
              Token
            </label>
            <select
              v-model="selectedToken"
              class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                     text-white focus:outline-none focus:border-blue-400"
            >
              <option value="">All Tokens</option>
              <option value="uatom">ATOM</option>
              <option value="uosmo">OSMO</option>
              <option value="uusdc">USDC</option>
              <option value="untrn">NTRN</option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">
              Min Stuck Duration
            </label>
            <select
              v-model="minStuckDuration"
              class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                     text-white focus:outline-none focus:border-blue-400"
            >
              <option value="0">Any Duration</option>
              <option value="300">5 minutes</option>
              <option value="900">15 minutes</option>
              <option value="3600">1 hour</option>
              <option value="86400">24 hours</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">
              Status
            </label>
            <select
              v-model="packetStatus"
              class="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg 
                     text-white focus:outline-none focus:border-blue-400"
            >
              <option value="stuck">Stuck</option>
              <option value="expired">Expired</option>
              <option value="expiring">Expiring Soon</option>
              <option value="all">All</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Search Actions -->
      <div class="flex gap-2 pt-4">
        <button
          @click="performSearch"
          :disabled="isSearching"
          class="flex-1 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 
                 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          <span v-if="!isSearching" class="flex items-center justify-center gap-2">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            Search
          </span>
          <span v-else>Searching...</span>
        </button>
        
        <button
          @click="resetSearch"
          class="px-4 py-2 bg-white/10 text-gray-300 rounded-lg hover:bg-white/20 transition-all"
        >
          Reset
        </button>
      </div>
    </div>

    <!-- Search Results -->
    <div v-if="searchResults !== null" class="mt-8">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-white">
          Results ({{ searchResults.length }})
        </h3>
        
        <button
          v-if="searchResults.length > 0"
          @click="exportResults"
          class="px-3 py-1 bg-white/10 text-gray-300 rounded-lg hover:bg-white/20 
                 transition-all text-sm flex items-center gap-1"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Export CSV
        </button>
      </div>

      <!-- No Results -->
      <div v-if="searchResults.length === 0" 
           class="text-center py-8 text-gray-400">
        No packets found matching your criteria
      </div>

      <!-- Results Table -->
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-gray-400 border-b border-white/10">
              <th class="pb-2 pr-4">Chain</th>
              <th class="pb-2 pr-4">Channel</th>
              <th class="pb-2 pr-4">Amount</th>
              <th class="pb-2 pr-4">Sender</th>
              <th class="pb-2 pr-4">Receiver</th>
              <th class="pb-2 pr-4">Age</th>
              <th class="pb-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="packet in searchResults" 
                :key="`${packet.chain_id}-${packet.sequence}`"
                class="border-b border-white/5 hover:bg-white/5 transition-colors">
              <td class="py-3 pr-4 text-white">
                {{ getChainName(packet.chain_id) }}
              </td>
              <td class="py-3 pr-4 text-gray-300">
                {{ packet.src_channel }}
              </td>
              <td class="py-3 pr-4 text-white">
                {{ formatAmount(packet.amount, packet.denom) }}
              </td>
              <td class="py-3 pr-4 text-gray-300 font-mono text-xs">
                {{ truncateAddress(packet.sender) }}
              </td>
              <td class="py-3 pr-4 text-gray-300 font-mono text-xs">
                {{ truncateAddress(packet.receiver) }}
              </td>
              <td class="py-3 pr-4 text-gray-300">
                {{ formatDuration(packet.age_seconds) }}
              </td>
              <td class="py-3">
                <button
                  @click="$emit('view-packet', packet)"
                  class="text-blue-400 hover:text-blue-300 transition-colors"
                >
                  View
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="searchResults.length > 0 && hasMore" class="mt-4 text-center">
        <button
          @click="loadMore"
          :disabled="isLoadingMore"
          class="px-4 py-2 bg-white/10 text-gray-300 rounded-lg hover:bg-white/20 
                 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          {{ isLoadingMore ? 'Loading...' : 'Load More' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '@/services/api'

interface PacketSearchResult {
  chain_id: string
  sequence: number
  src_channel: string
  dst_channel: string
  sender: string
  receiver: string
  amount: string
  denom: string
  age_seconds: number
  relay_attempts: number
  last_attempt_by: string
}

const emit = defineEmits<{
  'view-packet': [packet: PacketSearchResult]
}>()

// Search state
const searchType = ref<'wallet' | 'chain' | 'advanced'>('wallet')
const isSearching = ref(false)
const isLoadingMore = ref(false)
const searchResults = ref<PacketSearchResult[] | null>(null)
const currentOffset = ref(0)
const hasMore = ref(false)

// Search parameters
const walletAddress = ref('')
const searchAsSender = ref(true)
const searchAsReceiver = ref(true)
const sourceChain = ref('')
const destinationChain = ref('')
const minAmount = ref('')
const selectedToken = ref('')
const minStuckDuration = ref('0')
const packetStatus = ref('stuck')

const searchTypes = [
  { value: 'wallet', label: 'By Wallet' },
  { value: 'chain', label: 'By Chain' },
  { value: 'advanced', label: 'Advanced' }
]

const availableChains = [
  { id: 'cosmoshub-4', name: 'Cosmos Hub' },
  { id: 'osmosis-1', name: 'Osmosis' },
  { id: 'neutron-1', name: 'Neutron' },
  { id: 'noble-1', name: 'Noble' },
  { id: 'stride-1', name: 'Stride' },
  { id: 'axelar-dojo-1', name: 'Axelar' }
]

async function performSearch() {
  isSearching.value = true
  searchResults.value = null
  currentOffset.value = 0
  
  try {
    let results: PacketSearchResult[] = []
    
    if (searchType.value === 'wallet' && walletAddress.value) {
      // Search by wallet address
      if (searchAsSender.value) {
        const senderResults = await api.searchPackets({
          sender: walletAddress.value,
          min_age_seconds: parseInt(minStuckDuration.value),
          limit: 50
        })
        results = results.concat(senderResults.packets || [])
      }
      
      if (searchAsReceiver.value) {
        const receiverResults = await api.searchPackets({
          receiver: walletAddress.value,
          min_age_seconds: parseInt(minStuckDuration.value),
          limit: 50
        })
        results = results.concat(receiverResults.packets || [])
      }
      
      // Remove duplicates
      const uniqueResults = new Map()
      results.forEach(packet => {
        const key = `${packet.chain_id}-${packet.sequence}`
        if (!uniqueResults.has(key)) {
          uniqueResults.set(key, packet)
        }
      })
      results = Array.from(uniqueResults.values())
      
    } else if (searchType.value === 'chain') {
      // Search by chain
      const params: any = {
        min_age_seconds: parseInt(minStuckDuration.value),
        limit: 50
      }
      
      if (sourceChain.value) {
        params.chain_id = sourceChain.value
      }
      
      const response = await api.searchPackets(params)
      results = response.packets || []
      
      // Filter by destination chain if specified
      if (destinationChain.value) {
        results = results.filter(packet => {
          // This would need proper destination chain resolution
          return true // Placeholder
        })
      }
      
    } else if (searchType.value === 'advanced') {
      // Advanced search
      const params: any = {
        min_age_seconds: parseInt(minStuckDuration.value),
        limit: 50
      }
      
      if (selectedToken.value) {
        params.denom = selectedToken.value
      }
      
      const response = await api.searchPackets(params)
      results = response.packets || []
      
      // Client-side filtering for amount
      if (minAmount.value) {
        const minAmountNum = parseFloat(minAmount.value)
        results = results.filter(packet => {
          const amount = parseFloat(packet.amount) / 1000000 // Convert from micro units
          return amount >= minAmountNum
        })
      }
    }
    
    searchResults.value = results
    hasMore.value = results.length === 50
    
  } catch (error) {
    console.error('Search failed:', error)
    searchResults.value = []
  } finally {
    isSearching.value = false
  }
}

async function loadMore() {
  // Implementation for pagination
  isLoadingMore.value = true
  // ... load more results
  isLoadingMore.value = false
}

function resetSearch() {
  walletAddress.value = ''
  searchAsSender.value = true
  searchAsReceiver.value = true
  sourceChain.value = ''
  destinationChain.value = ''
  minAmount.value = ''
  selectedToken.value = ''
  minStuckDuration.value = '0'
  packetStatus.value = 'stuck'
  searchResults.value = null
}

function exportResults() {
  if (!searchResults.value) return
  
  // Create CSV content
  const headers = ['Chain', 'Channel', 'Sequence', 'Amount', 'Token', 'Sender', 'Receiver', 'Age', 'Attempts']
  const rows = searchResults.value.map(packet => [
    packet.chain_id,
    packet.src_channel,
    packet.sequence,
    packet.amount,
    packet.denom,
    packet.sender,
    packet.receiver,
    packet.age_seconds,
    packet.relay_attempts
  ])
  
  const csv = [headers, ...rows].map(row => row.join(',')).join('\n')
  
  // Download CSV
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `packet-search-${new Date().toISOString().split('T')[0]}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

// Helper functions
function getChainName(chainId: string): string {
  return availableChains.find(c => c.id === chainId)?.name || chainId
}

function formatAmount(amount: string, denom: string): string {
  const value = parseFloat(amount) / 1000000
  const symbol = denom.replace(/^u/, '').toUpperCase()
  return `${value.toLocaleString()} ${symbol}`
}

function truncateAddress(address: string): string {
  if (address.length <= 16) return address
  return `${address.slice(0, 8)}...${address.slice(-6)}`
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}
</script>

<style scoped>
.packet-search {
  background: rgba(17, 24, 39, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

input[type="checkbox"] {
  @apply w-4 h-4 bg-white/10 border-white/20 rounded focus:ring-2 focus:ring-blue-500;
}

select option {
  @apply bg-gray-800 text-white;
}
</style>