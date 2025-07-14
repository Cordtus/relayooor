<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900">IBC Relay Dashboard</h1>
    
    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Total Chains</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.chains }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Active Relayers</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.relayers }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">24h Packets</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.packets }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Success Rate</h3>
        <p class="mt-2 text-3xl font-semibold text-green-600">{{ stats.successRate }}%</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Activity -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Recent Activity</h2>
        <div v-if="recentActivity.length > 0" class="space-y-3">
          <div v-for="activity in recentActivity" :key="activity.timestamp" class="flex items-center justify-between py-2 border-b">
            <div>
              <p class="text-sm font-medium">{{ getChainName(activity.from_chain) }} → {{ getChainName(activity.to_chain) }}</p>
              <p class="text-xs text-gray-500">{{ activity.channel }}</p>
            </div>
            <span :class="[
              'text-xs',
              activity.status === 'success' ? 'text-green-600' : 'text-red-600'
            ]">{{ activity.status }}</span>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">No recent activity</div>
      </div>

      <!-- Top Relayers -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Top Relayers</h2>
        <div v-if="topRelayers.length > 0" class="space-y-3">
          <div v-for="relayer in topRelayers" :key="relayer.address" class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium">{{ formatAddress(relayer.address) }}</p>
              <p class="text-xs text-gray-500">Success Rate: {{ (relayer.successRate || 0).toFixed(1) }}%</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold">{{ formatNumber(relayer.totalPackets || 0) }}</p>
              <p class="text-xs text-gray-500">packets</p>
            </div>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">No relayer data available</div>
      </div>
    </div>

    <!-- Top Routes (by Activity) -->
    <div class="mt-6 bg-white rounded-lg shadow p-6">
      <h2 class="text-lg font-medium mb-4">Top Routes (by Activity)</h2>
      <div v-if="channelCongestion" class="space-y-3">
        <div v-for="channel in channelCongestion.channels" :key="`${channel.src_channel}-${channel.dst_channel}`" 
             class="flex items-center justify-between p-3 border rounded-lg"
             :class="channel.stuck_count > 0 ? 'border-orange-300 bg-orange-50' : 'border-gray-200'">
          <div>
            <p class="text-sm font-medium">{{ channel.src_channel }} → {{ channel.dst_channel }}</p>
            <p class="text-xs text-gray-600">
              {{ channel.stuck_count }} stuck packets
              <span v-if="channel.oldest_stuck_age_seconds">
                (oldest: {{ formatDuration(channel.oldest_stuck_age_seconds) }})
              </span>
            </p>
          </div>
          <div v-if="channel.total_value && Object.keys(channel.total_value).length > 0" class="text-right">
            <p v-for="(amount, denom) in channel.total_value" :key="denom" class="text-xs text-gray-600">
              {{ formatAmount(String(amount), String(denom)) }}
            </p>
          </div>
        </div>
      </div>
      <div v-else class="text-sm text-gray-500">Loading channel data...</div>
    </div>

    <!-- Quick Actions -->
    <div class="bg-blue-50 rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Quick Actions</h2>
      <div class="flex flex-wrap gap-4">
        <router-link
          to="/monitoring"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
        >
          View Full Monitoring
        </router-link>
        <router-link
          to="/packet-clearing"
          class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
        >
          Clear Stuck Packets
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { api, metricsService } from '@/services/api'

// Fetch monitoring data
const { data: monitoringData } = useQuery({
  queryKey: ['monitoring-data'],
  queryFn: async () => {
    return metricsService.getMonitoringData()
  },
  refetchInterval: 10000 // Refresh every 10 seconds
})

// Fetch comprehensive metrics
const { data: comprehensiveMetrics } = useQuery({
  queryKey: ['comprehensive-metrics'],
  queryFn: async () => {
    try {
      const response = await api.get('/api/monitoring/metrics')
      return response.data
    } catch (error) {
      // Use metrics service which has mock data fallback
      const raw = await metricsService.getRawMetrics()
      return metricsService.parsePrometheusMetrics(raw)
    }
  },
  refetchInterval: 10000
})

// Fetch channel congestion data
const { data: channelCongestion } = useQuery({
  queryKey: ['channel-congestion'],
  queryFn: async () => {
    try {
      const response = await api.get('/api/channels/congestion')
      return response.data
    } catch (error) {
      // Return channels from monitoring data
      const data = await metricsService.getMonitoringData()
      return data.channels || []
    }
  },
  refetchInterval: 30000 // Refresh every 30 seconds
})

const stats = ref({
  chains: 0,
  relayers: 0,
  packets: '0',
  successRate: 0
})

const topRelayers = ref<any[]>([])
const recentActivity = ref<any[]>([])

// Update stats when data is loaded
onMounted(() => {
  const updateStats = () => {
    if (comprehensiveMetrics.value) {
      // Use comprehensive metrics for all stats
      stats.value.chains = comprehensiveMetrics.value.system?.totalChains || 0
      stats.value.relayers = comprehensiveMetrics.value.relayers?.length || 0
      stats.value.packets = formatPacketCount(comprehensiveMetrics.value.system?.totalPackets || 0)
      
      // Calculate success rate from channel data
      const channels = comprehensiveMetrics.value.channels || []
      if (channels.length > 0) {
        const avgSuccessRate = channels.reduce((acc: number, ch: any) => acc + (ch.successRate || 0), 0) / channels.length
        stats.value.successRate = Math.round(avgSuccessRate * 10) / 10
      }
      
      // Use relayers from comprehensive metrics
      topRelayers.value = comprehensiveMetrics.value.relayers?.slice(0, 5) || []
      recentActivity.value = comprehensiveMetrics.value.recentPackets?.slice(0, 5).map((p: any) => ({
        from_chain: p.chain_id,
        to_chain: p.dst_channel?.includes('channel-141') ? 'cosmoshub-4' : 'osmosis-1',
        channel: p.src_channel,
        status: p.effected ? 'success' : 'pending',
        timestamp: p.timestamp
      })) || []
    }
    if (monitoringData.value) {
      // Fallback to monitoring data if available
      if (!topRelayers.value.length && monitoringData.value.top_relayers) {
        topRelayers.value = monitoringData.value.top_relayers
      }
      if (!recentActivity.value.length && monitoringData.value.recent_activity) {
        recentActivity.value = monitoringData.value.recent_activity
      }
    }
  }
  
  // Update immediately if data is available
  updateStats()
  
  // Watch for changes
  const interval = setInterval(updateStats, 1000)
  
  // Cleanup
  return () => clearInterval(interval)
})

function formatPacketCount(count: number): string {
  if (count >= 1000000) return (count / 1000000).toFixed(1) + 'M'
  if (count >= 1000) return (count / 1000).toFixed(1) + 'K'
  return count.toString()
}

function getChainName(chainId: string): string {
  const names: Record<string, string> = {
    'osmosis-1': 'Osmosis',
    'cosmoshub-4': 'Cosmos Hub',
    'neutron-1': 'Neutron'
  }
  return names[chainId] || chainId
}

function formatAddress(address: string): string {
  if (!address || address.length < 10) return address
  return address.slice(0, 10) + '...' + address.slice(-4)
}

function formatNumber(num: number): string {
  return new Intl.NumberFormat().format(num)
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

function formatAmount(amount: string, denom: string): string {
  const value = parseFloat(amount) / 1000000
  const symbols: Record<string, string> = {
    'uatom': 'ATOM',
    'uosmo': 'OSMO',
    'untrn': 'NTRN'
  }
  return `${value.toFixed(2)} ${symbols[denom] || denom}`
}
</script>