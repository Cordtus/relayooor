<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">Relayer Analytics</h1>
      <div class="flex items-center gap-4">
        <select v-model="timeRange" class="rounded-md border-gray-300 text-sm">
          <option value="1h">Last Hour</option>
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
        </select>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by address..."
          class="rounded-md border-gray-300 text-sm"
        />
      </div>
    </div>

    <!-- Relayer Competition Overview -->
    <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4">
      <div class="bg-white p-4 rounded-lg shadow">
        <h3 class="text-xs font-medium text-gray-500 uppercase">Total Relayers</h3>
        <p class="mt-2 text-2xl font-semibold text-gray-900">{{ totalRelayers }}</p>
        <p class="text-xs text-gray-600 mt-1">{{ activeRelayers }} active</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <h3 class="text-xs font-medium text-gray-500 uppercase">Market Leader</h3>
        <p class="mt-2 text-sm font-semibold text-gray-900 truncate">{{ marketLeader }}</p>
        <p class="text-xs text-gray-600 mt-1">{{ marketLeaderShare }}% share</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <h3 class="text-xs font-medium text-gray-500 uppercase">Avg Success</h3>
        <p class="mt-2 text-2xl font-semibold text-green-600">{{ avgSuccessRate }}%</p>
        <p class="text-xs text-gray-600 mt-1">across all</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <h3 class="text-xs font-medium text-gray-500 uppercase">Competition</h3>
        <p class="mt-2 text-2xl font-semibold text-orange-600">{{ competitionIndex }}</p>
        <p class="text-xs text-gray-600 mt-1">HHI index</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <h3 class="text-xs font-medium text-gray-500 uppercase">Frontrun Rate</h3>
        <p class="mt-2 text-2xl font-semibold text-red-600">{{ frontrunRate }}%</p>
        <p class="text-xs text-gray-600 mt-1">of packets</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow">
        <h3 class="text-xs font-medium text-gray-500 uppercase">New Entrants</h3>
        <p class="mt-2 text-2xl font-semibold text-purple-600">{{ newRelayers }}</p>
        <p class="text-xs text-gray-600 mt-1">this week</p>
      </div>
    </div>

    <!-- Performance Leaderboard -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h2 class="text-lg font-medium">Performance Leaderboard</h2>
        <div class="flex gap-2">
          <button
            v-for="metric in ['packets', 'success', 'efficiency', 'revenue']"
            :key="metric"
            @click="leaderboardMetric = metric"
            :class="[
              'px-3 py-1 text-xs font-medium rounded-md',
              leaderboardMetric === metric
                ? 'bg-primary-100 text-primary-700'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            ]"
          >
            {{ metric }}
          </button>
        </div>
      </div>
      <div class="p-6">
        <div class="space-y-4">
          <div
            v-for="(relayer, index) in topRelayers"
            :key="relayer.address"
            class="flex items-center justify-between p-4 rounded-lg hover:bg-gray-50 transition-colors"
            :class="index < 3 ? 'bg-gradient-to-r from-yellow-50 to-orange-50' : ''"
          >
            <div class="flex items-center space-x-4">
              <div class="flex-shrink-0">
                <div
                  class="w-10 h-10 rounded-full flex items-center justify-center font-bold text-white"
                  :class="[
                    index === 0 ? 'bg-yellow-500' :
                    index === 1 ? 'bg-gray-400' :
                    index === 2 ? 'bg-orange-600' :
                    'bg-gray-300'
                  ]"
                >
                  {{ index + 1 }}
                </div>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900">
                  {{ formatAddress(relayer.address) }}
                </p>
                <p class="text-xs text-gray-500">
                  {{ relayer.software }} {{ relayer.version }} | {{ relayer.memo }}
                </p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-lg font-semibold text-gray-900">
                {{ getMetricValue(relayer, leaderboardMetric) }}
              </p>
              <p class="text-xs text-gray-500">
                {{ getMetricLabel(leaderboardMetric) }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Detailed Analytics -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Market Share Distribution -->
      <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-lg font-medium mb-4">Market Share Distribution</h3>
        <MarketSharePieChart :relayers="relayers" />
      </div>

      <!-- Software Version Distribution -->
      <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-lg font-medium mb-4">Software Distribution</h3>
        <SoftwareDistributionChart :relayers="relayers" />
      </div>

      <!-- Performance Matrix -->
      <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-lg font-medium mb-4">Performance Matrix</h3>
        <RelayerPerformanceMatrix :relayers="relayers" />
      </div>

      <!-- Frontrun Analysis -->
      <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-lg font-medium mb-4">Frontrun Competition</h3>
        <FrontrunAnalysis :events="frontrunEvents" :relayers="relayers" />
      </div>
    </div>

    <!-- Relayer Details Table -->
    <div class="bg-white shadow rounded-lg">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-lg font-medium">All Relayers</h2>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Relayer
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Total Packets
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Effected
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Success Rate
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Frontrun Count
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Market Share
              </th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                Software
              </th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="relayer in filteredRelayers" :key="relayer.address" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div>
                  <div class="text-sm font-medium text-gray-900">
                    {{ formatAddress(relayer.address) }}
                  </div>
                  <div class="text-xs text-gray-500 truncate max-w-xs">
                    {{ relayer.memo || 'No memo' }}
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                {{ relayer.totalPackets.toLocaleString() }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                {{ relayer.effectedPackets.toLocaleString() }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <span :class="[
                  'text-sm font-medium',
                  relayer.successRate >= 90 ? 'text-green-600' :
                  relayer.successRate >= 75 ? 'text-yellow-600' : 'text-red-600'
                ]">
                  {{ relayer.successRate.toFixed(1) }}%
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                {{ relayer.frontrunCount }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                {{ ((relayer.totalPackets / totalPackets) * 100).toFixed(1) }}%
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">
                  {{ relayer.software }} {{ relayer.version }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span :class="[
                  'px-2 inline-flex text-xs leading-5 font-semibold rounded-full',
                  isActive(relayer) ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                ]">
                  {{ isActive(relayer) ? 'Active' : 'Inactive' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { metricsService } from '@/services/api'
import MarketSharePieChart from '@/components/relayers/MarketSharePieChart.vue'
import SoftwareDistributionChart from '@/components/relayers/SoftwareDistributionChart.vue'
import RelayerPerformanceMatrix from '@/components/relayers/RelayerPerformanceMatrix.vue'
import FrontrunAnalysis from '@/components/relayers/FrontrunAnalysis.vue'

// State
const timeRange = ref('24h')
const searchQuery = ref('')
const leaderboardMetric = ref('packets')

// Fetch relayer data
const { data: metrics } = useQuery({
  queryKey: ['relayers', timeRange],
  queryFn: async () => {
    const raw = await metricsService.getRawMetrics()
    return metricsService.parsePrometheusMetrics(raw)
  },
  refetchInterval: 30000
})

// Computed data
const relayers = computed(() => metrics.value?.relayers || [])
const frontrunEvents = computed(() => metrics.value?.frontrunEvents || [])

const totalRelayers = computed(() => relayers.value.length)
const activeRelayers = computed(() => relayers.value.filter(r => r.totalPackets > 0).length)

const totalPackets = computed(() => 
  relayers.value.reduce((sum, r) => sum + r.totalPackets, 0)
)

const marketLeader = computed(() => {
  const sorted = [...relayers.value].sort((a, b) => b.totalPackets - a.totalPackets)
  return sorted[0]?.address.slice(0, 10) + '...' || 'N/A'
})

const marketLeaderShare = computed(() => {
  const sorted = [...relayers.value].sort((a, b) => b.totalPackets - a.totalPackets)
  if (!sorted[0] || totalPackets.value === 0) return 0
  return ((sorted[0].totalPackets / totalPackets.value) * 100).toFixed(1)
})

const avgSuccessRate = computed(() => {
  if (relayers.value.length === 0) return 0
  const total = relayers.value.reduce((sum, r) => sum + r.successRate, 0)
  return Math.round(total / relayers.value.length)
})

const competitionIndex = computed(() => {
  // Herfindahl-Hirschman Index (HHI)
  if (totalPackets.value === 0) return 0
  const shares = relayers.value.map(r => (r.totalPackets / totalPackets.value) * 100)
  const hhi = shares.reduce((sum, share) => sum + share * share, 0)
  return Math.round(hhi)
})

const frontrunRate = computed(() => {
  const totalFrontrun = relayers.value.reduce((sum, r) => sum + r.uneffectedPackets, 0)
  if (totalPackets.value === 0) return 0
  return ((totalFrontrun / totalPackets.value) * 100).toFixed(1)
})

const newRelayers = computed(() => {
  // Mock data - would need historical data to calculate
  return 3
})

const topRelayers = computed(() => {
  const sorted = [...relayers.value].sort((a, b) => {
    switch (leaderboardMetric.value) {
      case 'success':
        return b.successRate - a.successRate
      case 'efficiency':
        return (b.effectedPackets / b.totalPackets) - (a.effectedPackets / a.totalPackets)
      case 'revenue':
        // Mock revenue calculation
        return (b.effectedPackets * 0.001) - (a.effectedPackets * 0.001)
      default:
        return b.totalPackets - a.totalPackets
    }
  })
  return sorted.slice(0, 10)
})

const filteredRelayers = computed(() => {
  if (!searchQuery.value) return relayers.value
  
  const query = searchQuery.value.toLowerCase()
  return relayers.value.filter(r => 
    r.address.toLowerCase().includes(query) ||
    r.memo.toLowerCase().includes(query)
  )
})

// Helper functions
function formatAddress(address: string): string {
  return `${address.slice(0, 10)}...${address.slice(-6)}`
}

function getMetricValue(relayer: any, metric: string): string {
  switch (metric) {
    case 'success':
      return `${relayer.successRate.toFixed(1)}%`
    case 'efficiency':
      const eff = (relayer.effectedPackets / relayer.totalPackets) * 100
      return `${eff.toFixed(1)}%`
    case 'revenue':
      // Mock revenue calculation
      return `$${(relayer.effectedPackets * 0.001).toFixed(2)}`
    default:
      return relayer.totalPackets.toLocaleString()
  }
}

function getMetricLabel(metric: string): string {
  switch (metric) {
    case 'success':
      return 'success rate'
    case 'efficiency':
      return 'efficiency'
    case 'revenue':
      return 'est. revenue'
    default:
      return 'packets relayed'
  }
}

function isActive(relayer: any): boolean {
  // Consider active if relayed packets in recent time
  return relayer.totalPackets > 0
}
</script>