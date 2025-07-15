<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">IBC Channels</h1>
      <div class="flex items-center gap-4">
        <select v-model="filterChain" class="rounded-md border-gray-300 text-sm">
          <option value="">All Chains</option>
          <option v-for="chain in chainRegistry?.chains || []" :key="chain.chain_id" :value="chain.chain_id">
            {{ chain.pretty_name }}
          </option>
        </select>
        <select v-model="sortBy" class="rounded-md border-gray-300 text-sm">
          <option value="volume">Sort by Volume</option>
          <option value="success">Sort by Success Rate</option>
          <option value="congestion">Sort by Congestion</option>
        </select>
      </div>
    </div>

    <!-- Channel Overview Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Total Channels</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ totalChannels }}</p>
        <p class="text-sm text-gray-600 mt-1">{{ activeChannels }} active</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Total Volume (24h)</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ formatNumber(totalVolume) }}</p>
        <p class="text-sm text-gray-600 mt-1">packets relayed</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Avg Success Rate</h3>
        <p class="mt-2 text-3xl font-semibold text-green-600">{{ avgSuccessRate }}%</p>
        <p class="text-sm text-gray-600 mt-1">across all channels</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Congested Channels</h3>
        <p class="mt-2 text-3xl font-semibold text-orange-600">{{ congestedChannels }}</p>
        <p class="text-sm text-gray-600 mt-1">need attention</p>
      </div>
    </div>

    <!-- Channel Table -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-lg font-medium">Channel Performance</h2>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Channel
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Route
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                24h Volume
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Success Rate
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Avg Time
              </th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="channel in sortedChannels" :key="channel.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div>
                  <div class="text-sm font-medium text-gray-900">
                    {{ channel.srcChannel }}
                  </div>
                  <div class="text-sm text-gray-500">
                    ↓ {{ channel.dstChannel }}
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div>
                    <div class="text-sm text-gray-900">{{ getChainName(channel.srcChain) }}</div>
                    <div class="text-sm text-gray-500">→ {{ getChainName(channel.dstChain) }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                {{ formatNumber(channel.volume24h) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <div class="flex items-center justify-end">
                  <span :class="[
                    'text-sm font-medium',
                    channel.successRate >= 90 ? 'text-green-600' :
                    channel.successRate >= 75 ? 'text-yellow-600' : 'text-red-600'
                  ]">
                    {{ channel.successRate.toFixed(1) }}%
                  </span>
                  <div class="ml-2 w-16 bg-gray-200 rounded-full h-2">
                    <div 
                      class="h-2 rounded-full"
                      :class="[
                        channel.successRate >= 90 ? 'bg-green-600' :
                        channel.successRate >= 75 ? 'bg-yellow-600' : 'bg-red-600'
                      ]"
                      :style="{ width: channel.successRate + '%' }"
                    ></div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                {{ channel.avgProcessingTime }}s
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span :class="[
                  'px-2 inline-flex text-xs leading-5 font-semibold rounded-full',
                  channel.status === 'active' ? 'bg-green-100 text-green-800' :
                  channel.status === 'congested' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-gray-100 text-gray-800'
                ]">
                  {{ channel.status }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="viewDetails(channel)" class="text-blue-600 hover:text-blue-900 font-medium">
                  Details →
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Channel Flow Visualization -->
    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Channel Flow Visualization</h2>
      <ChannelFlowDiagram :channels="channels" />
    </div>

    <!-- Channel Details Modal -->
    <ChannelDetailsModal 
      v-if="selectedChannel"
      :channel="selectedChannel"
      @close="selectedChannel = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { metricsService, chainRegistryService } from '@/services/api'
import ChannelFlowDiagram from '@/components/channels/ChannelFlowDiagram.vue'
import ChannelDetailsModal from '@/components/channels/ChannelDetailsModal.vue'

// State
const filterChain = ref('')
const sortBy = ref('volume')
const selectedChannel = ref<any>(null)

// Load chain registry on mount
onMounted(async () => {
  await chainRegistryService.getChainRegistry()
})

// Fetch channel data
const { data: metrics } = useQuery({
  queryKey: ['channels'],
  queryFn: async () => {
    const raw = await metricsService.getRawMetrics()
    return metricsService.parsePrometheusMetrics(raw)
  },
  refetchInterval: 30000
})

// Fetch chain registry
const { data: chainRegistry } = useQuery({
  queryKey: ['chain-registry'],
  queryFn: () => chainRegistryService.getChainRegistry(),
  staleTime: 600000 // Cache for 10 minutes
})

// Mock enhanced channel data
const channels = computed(() => {
  if (!metrics.value) return []
  return metrics.value.channels.map(channel => ({
    ...channel,
    id: `${channel.srcChain}-${channel.srcChannel}`,
    volume24h: channel.totalPackets,
    avgProcessingTime: channel.avgProcessingTime || 10, // Use actual or default 10s
    status: channel.successRate > 90 ? 'active' : 
            channel.successRate > 75 ? 'congested' : 'degraded'
  }))
})

// Computed stats
const totalChannels = computed(() => channels.value.length)
const activeChannels = computed(() => channels.value.filter(c => c.status === 'active').length)
const totalVolume = computed(() => channels.value.reduce((sum, c) => sum + c.volume24h, 0))
const avgSuccessRate = computed(() => {
  if (channels.value.length === 0) return 0
  const total = channels.value.reduce((sum, c) => sum + c.successRate, 0)
  return Math.round(total / channels.value.length)
})
const congestedChannels = computed(() => channels.value.filter(c => c.status === 'congested').length)

// Filtered and sorted channels
const sortedChannels = computed(() => {
  let filtered = channels.value
  
  if (filterChain.value) {
    filtered = filtered.filter(c => 
      c.srcChain === filterChain.value || c.dstChain === filterChain.value
    )
  }
  
  return filtered.sort((a, b) => {
    switch (sortBy.value) {
      case 'success':
        return b.successRate - a.successRate
      case 'congestion':
        return a.successRate - b.successRate
      default:
        return b.volume24h - a.volume24h
    }
  })
})

// Helper functions
function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function getChainName(chainId: string): string {
  // Use chain registry service which has caching
  return chainRegistryService.getChainName(chainId)
}

function viewDetails(channel: any) {
  selectedChannel.value = channel
}
</script>