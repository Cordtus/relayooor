<template>
  <div class="bg-white shadow rounded-lg overflow-hidden">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200">
      <h3 class="text-lg leading-6 font-medium text-gray-900">
        Channel Performance Metrics
      </h3>
      <p class="mt-1 text-sm text-gray-500">
        Real-time performance data for all active IBC channels
      </p>
    </div>
    
    <!-- Table -->
    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th 
              v-for="column in columns" 
              :key="column.key"
              @click="column.sortable !== false && handleSort(column.key)"
              :class="[
                'px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider',
                column.sortable !== false ? 'cursor-pointer hover:bg-gray-100' : ''
              ]"
            >
              <div class="flex items-center gap-1">
                {{ column.label }}
                <ChevronUp 
                  v-if="sortBy === column.key && sortOrder === 'asc'" 
                  class="h-3 w-3"
                />
                <ChevronDown 
                  v-if="sortBy === column.key && sortOrder === 'desc'" 
                  class="h-3 w-3"
                />
              </div>
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr 
            v-for="channel in sortedChannels" 
            :key="`${channel.src_chain}-${channel.src_channel}`"
            class="hover:bg-gray-50"
          >
            <!-- Channel Info -->
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div>
                  <div class="text-sm font-medium text-gray-900">
                    {{ formatChainName(channel.src_chain) }} → {{ formatChainName(channel.dst_chain) }}
                  </div>
                  <div class="text-xs text-gray-500">
                    {{ channel.src_channel }} ↔ {{ channel.dst_channel }}
                  </div>
                </div>
              </div>
            </td>
            
            <!-- Total Packets -->
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm text-gray-900">{{ formatNumber(channel.totalPackets) }}</div>
            </td>
            
            <!-- Success Rate -->
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div 
                  :class="[
                    'text-sm font-medium',
                    channel.successRate >= 95 ? 'text-green-600' :
                    channel.successRate >= 80 ? 'text-yellow-600' :
                    'text-red-600'
                  ]"
                >
                  {{ channel.successRate.toFixed(1) }}%
                </div>
                <div class="ml-2 w-16 bg-gray-200 rounded-full h-2">
                  <div 
                    :class="[
                      'h-2 rounded-full',
                      channel.successRate >= 95 ? 'bg-green-500' :
                      channel.successRate >= 80 ? 'bg-yellow-500' :
                      'bg-red-500'
                    ]"
                    :style="`width: ${channel.successRate}%`"
                  />
                </div>
              </div>
            </td>
            
            <!-- Stuck Packets -->
            <td class="px-6 py-4 whitespace-nowrap">
              <span 
                :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  channel.stuckPackets > 0 ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'
                ]"
              >
                {{ channel.stuckPackets || 0 }}
              </span>
            </td>
            
            <!-- Avg Clearing Time -->
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDuration(channel.avgClearingTime || 0) }}
            </td>
            
            <!-- Congestion Score -->
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <span 
                  :class="[
                    'text-sm font-medium',
                    calculateCongestionScore(channel) <= 30 ? 'text-green-600' :
                    calculateCongestionScore(channel) <= 70 ? 'text-yellow-600' :
                    'text-red-600'
                  ]"
                >
                  {{ calculateCongestionScore(channel) }}%
                </span>
                <AlertTriangle 
                  v-if="calculateCongestionScore(channel) > 70" 
                  class="ml-1 h-4 w-4 text-red-500"
                />
              </div>
            </td>
            
            <!-- Status -->
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center gap-2">
                <span 
                  :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    getStatusClass(channel)
                  ]"
                  :title="getStatusReason(channel)"
                >
                  {{ getStatus(channel) }}
                </span>
                <span v-if="getStatus(channel) !== 'Healthy'" class="text-xs text-gray-500">
                  {{ getStatusReason(channel) }}
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- Empty State -->
    <div v-if="!channels || channels.length === 0" class="text-center py-8">
      <Package class="mx-auto h-12 w-12 text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900">No channels data</h3>
      <p class="mt-1 text-sm text-gray-500">Channel performance data will appear here.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ChevronUp, ChevronDown, AlertTriangle, Package } from 'lucide-vue-next'

interface Channel {
  src_chain: string
  dst_chain: string
  src_channel: string
  dst_channel: string
  totalPackets: number
  effectedPackets: number
  uneffectedPackets: number
  successRate: number
  stuckPackets?: number
  avgClearingTime?: number
  congestionScore?: number
}

const props = defineProps<{
  channels?: Channel[]
  sortBy?: string
}>()

const emit = defineEmits<{
  sort: [column: string]
}>()

// Local sort state
const sortBy = ref(props.sortBy || 'totalPackets')
const sortOrder = ref<'asc' | 'desc'>('desc')

const columns = [
  { key: 'channel', label: 'Channel', sortable: false },
  { key: 'totalPackets', label: 'Total Packets' },
  { key: 'successRate', label: 'Success Rate' },
  { key: 'stuckPackets', label: 'Stuck' },
  { key: 'avgClearingTime', label: 'Avg Clear Time' },
  { key: 'congestionScore', label: 'Congestion' },
  { key: 'status', label: 'Status', sortable: false }
]

function calculateCongestionScore(channel: any): number {
  // Calculate congestion based on success rate and uneffected packets ratio
  if (!channel.totalPackets || channel.totalPackets === 0) return 0
  
  // Lower success rate = higher congestion
  const successRateFactor = Math.max(0, 100 - (channel.successRate || 0))
  
  // Higher uneffected ratio = higher congestion
  const uneffectedRatio = (channel.uneffectedPackets || 0) / channel.totalPackets * 100
  
  // Combined score (weighted average)
  return Math.round((successRateFactor * 0.7) + (uneffectedRatio * 0.3))
}

const sortedChannels = computed(() => {
  if (!props.channels) return []
  
  const sorted = [...props.channels].sort((a, b) => {
    let aVal: any, bVal: any
    
    switch (sortBy.value) {
      case 'totalPackets':
        aVal = a.totalPackets
        bVal = b.totalPackets
        break
      case 'successRate':
        aVal = a.successRate
        bVal = b.successRate
        break
      case 'stuckPackets':
        aVal = a.stuckPackets || 0
        bVal = b.stuckPackets || 0
        break
      case 'avgClearingTime':
        aVal = a.avgClearingTime || 0
        bVal = b.avgClearingTime || 0
        break
      case 'congestionScore':
        aVal = calculateCongestionScore(a)
        bVal = calculateCongestionScore(b)
        break
      default:
        return 0
    }
    
    if (sortOrder.value === 'asc') {
      return aVal - bVal
    } else {
      return bVal - aVal
    }
  })
  
  return sorted
})

function handleSort(column: string) {
  if (sortBy.value === column) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortOrder.value = 'desc'
  }
  emit('sort', column)
}

function formatChainName(chainId: string): string {
  const names: Record<string, string> = {
    'cosmoshub-4': 'Cosmos Hub',
    'osmosis-1': 'Osmosis',
    'neutron-1': 'Neutron',
    'stargaze-1': 'Stargaze',
    'juno-1': 'Juno'
  }
  return names[chainId] || chainId
}

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function formatDuration(seconds: number): string {
  if (seconds === 0) return 'N/A'
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  return `${Math.floor(seconds / 3600)}h`
}

function getStatus(channel: Channel): string {
  if (channel.successRate >= 95 && (!channel.stuckPackets || channel.stuckPackets === 0)) {
    return 'Healthy'
  } else if (channel.successRate >= 80 || (channel.stuckPackets && channel.stuckPackets < 5)) {
    return 'Warning'
  } else {
    return 'Critical'
  }
}

function getStatusClass(channel: Channel): string {
  const status = getStatus(channel)
  switch (status) {
    case 'Healthy':
      return 'bg-green-100 text-green-800'
    case 'Warning':
      return 'bg-yellow-100 text-yellow-800'
    case 'Critical':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

function getStatusReason(channel: Channel): string {
  const reasons = []
  
  if (channel.successRate < 95) {
    reasons.push(`${channel.successRate.toFixed(1)}% success rate`)
  }
  
  if (channel.stuckPackets && channel.stuckPackets > 0) {
    reasons.push(`${channel.stuckPackets} stuck packets`)
  }
  
  const congestion = calculateCongestionScore(channel)
  if (congestion > 50) {
    reasons.push(`${congestion}% congestion`)
  }
  
  return reasons.join(', ') || 'Operating normally'
}
</script>