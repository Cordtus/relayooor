<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-900">IBC Relay Dashboard</h1>
      <RefreshRateSelector :lastUpdate="lastUpdateTime" />
    </div>
    
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
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ formatNumber(stats.packets) }}</p>
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
              <p class="text-sm font-medium">{{ getChainNameSync(activity.from_chain) }} → {{ getChainNameSync(activity.to_chain) }}</p>
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

    <!-- Three Column Layout for Additional Stats -->
    <div class="mt-6 grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Top 5 Token Routes -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Top Token Routes</h2>
        <div v-if="topTokenRoutes.length > 0" class="space-y-3">
          <div v-for="(route, index) in topTokenRoutes" :key="index" 
               class="flex items-center justify-between p-2 border-b">
            <div class="flex-1">
              <p class="text-sm font-medium">{{ route.token }}</p>
              <p class="text-xs text-gray-500">{{ route.srcChain }} → {{ route.dstChain }}</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold">{{ formatNumber(route.packetCount) }}</p>
              <p class="text-xs text-gray-500">packets</p>
            </div>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">No token route data</div>
      </div>

      <!-- Top 5 Relayers by Memo -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Top Relayers by Label</h2>
        <div v-if="topRelayersByMemo.length > 0" class="space-y-3">
          <div v-for="relayer in topRelayersByMemo" :key="relayer.memo || relayer.address" 
               class="flex items-center justify-between p-2 border-b">
            <div class="flex-1">
              <p class="text-sm font-medium">{{ relayer.memo || 'Anonymous' }}</p>
              <p class="text-xs text-gray-500">{{ relayer.addressCount }} address{{ relayer.addressCount > 1 ? 'es' : '' }}</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold">{{ formatNumber(relayer.totalPackets) }}</p>
              <p class="text-xs text-gray-500">{{ relayer.successRate.toFixed(1) }}% success</p>
            </div>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">No relayer data</div>
      </div>

      <!-- Top 5 Chains by Timeouts -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Chains by Packet Timeouts</h2>
        <div v-if="topChainsByTimeouts.length > 0" class="space-y-3">
          <div v-for="chain in topChainsByTimeouts" :key="chain.chainId" 
               class="flex items-center justify-between p-2 border-b">
            <div class="flex-1">
              <p class="text-sm font-medium">{{ chain.chainName || chain.chainId }}</p>
              <p class="text-xs text-gray-500">{{ formatNumber(chain.totalPackets) }} total packets</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold text-orange-600">{{ chain.timeouts }}</p>
              <p class="text-xs text-gray-500">timeouts</p>
            </div>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">No timeout data</div>
      </div>
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
import { ref, onMounted, watchEffect } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { api, metricsService } from '@/services/api'
import { configService } from '@/services/config'
import { useSettingsStore } from '@/stores/settings'
import { formatNumber, formatAddress, formatNumberWithCommas, formatDuration, formatAmount } from '@/utils/formatting'
import { REFRESH_INTERVALS } from '@/config/constants'
import RefreshRateSelector from '@/components/RefreshRateSelector.vue'

const settingsStore = useSettingsStore()
const lastUpdateTime = ref(new Date())

// Fetch monitoring data
const { data: monitoringData } = useQuery({
  queryKey: ['monitoring-data'],
  queryFn: async () => {
    lastUpdateTime.value = new Date()
    return metricsService.getMonitoringData()
  },
  refetchInterval: () => settingsStore.settings.refreshInterval
})

// Fetch comprehensive metrics
const { data: comprehensiveMetrics } = useQuery({
  queryKey: ['comprehensive-metrics'],
  queryFn: async () => {
    lastUpdateTime.value = new Date()
    // Get structured metrics from the monitoring endpoint
    return metricsService.getMonitoringMetrics()
  },
  refetchInterval: () => settingsStore.settings.refreshInterval
})

// Fetch stuck packets data
const { data: stuckPacketsData } = useQuery({
  queryKey: ['stuck-packets'],
  queryFn: async () => {
    try {
      const response = await api.get('/packets/stuck')
      return response.data
    } catch (error) {
      return { packets: [] }
    }
  },
  refetchInterval: () => settingsStore.settings.refreshInterval
})

const stats = ref({
  chains: 0,
  relayers: 0,
  packets: 0,
  successRate: 0
})

const topRelayers = ref<any[]>([])
const recentActivity = ref<any[]>([])
const topTokenRoutes = ref<any[]>([])
const topRelayersByMemo = ref<any[]>([])
const topChainsByTimeouts = ref<any[]>([])

// Update stats when data is loaded
// Update stats function
const updateStats = () => {
  if (comprehensiveMetrics.value) {
    // Use comprehensive metrics for all stats
    stats.value.chains = comprehensiveMetrics.value.system?.totalChains || 0
    stats.value.relayers = comprehensiveMetrics.value.relayers?.length || 0
    stats.value.packets = comprehensiveMetrics.value.system?.totalPackets || 0
    
    // Calculate success rate from channel data
    const channels = comprehensiveMetrics.value.channels || []
    if (channels.length > 0) {
      const avgSuccessRate = channels.reduce((acc: number, ch: any) => acc + (ch.successRate || 0), 0) / channels.length
      stats.value.successRate = Math.round(avgSuccessRate * 10) / 10
    }
    
    // Use relayers from comprehensive metrics
    topRelayers.value = comprehensiveMetrics.value.relayers?.slice(0, 5) || []
    
    // Calculate top token routes from stuck packets and channel data
    if (stuckPacketsData.value?.packets) {
      const tokenRouteMap = new Map()
      stuckPacketsData.value.packets.forEach((packet: any) => {
        const token = extractTokenFromDenom(packet.denom)
        const key = `${token}-${packet.chain_id}-${packet.dst_channel}`
        if (!tokenRouteMap.has(key)) {
          tokenRouteMap.set(key, {
            token,
            srcChain: packet.chain_id,
            dstChain: getDestChainFromChannel(packet.dst_channel),
            packetCount: 0,
            totalValue: 0
          })
        }
        const route = tokenRouteMap.get(key)
        route.packetCount++
        route.totalValue += parseInt(packet.amount) || 0
      })
      topTokenRoutes.value = Array.from(tokenRouteMap.values())
        .sort((a, b) => b.packetCount - a.packetCount)
        .slice(0, 5)
    }
    
    // Group relayers by memo
    if (comprehensiveMetrics.value.relayers) {
      const memoMap = new Map()
      comprehensiveMetrics.value.relayers.forEach((relayer: any) => {
        const memo = relayer.memo || 'Anonymous'
        if (!memoMap.has(memo)) {
          memoMap.set(memo, {
            memo,
            addresses: new Set(),
            totalPackets: 0,
            effectedPackets: 0,
            addressCount: 0
          })
        }
        const group = memoMap.get(memo)
        group.addresses.add(relayer.address)
        group.totalPackets += relayer.totalPackets
        group.effectedPackets += relayer.effectedPackets
      })
      
      topRelayersByMemo.value = Array.from(memoMap.values())
        .map(group => ({
          ...group,
          addressCount: group.addresses.size,
          successRate: group.totalPackets > 0 ? (group.effectedPackets / group.totalPackets) * 100 : 0
        }))
        .sort((a, b) => b.totalPackets - a.totalPackets)
        .slice(0, 5)
    }
    
    // Calculate chains by timeouts
    if (comprehensiveMetrics.value.chains) {
      topChainsByTimeouts.value = comprehensiveMetrics.value.chains
        .map((chain: any) => ({
          ...chain,
          timeouts: chain.timeouts || 0
        }))
        .filter(chain => chain.timeouts > 0)
        .sort((a, b) => b.timeouts - a.timeouts)
        .slice(0, 5)
    }
    
    // For recent activity, check if we have recentPackets, otherwise use monitoring data
    if (comprehensiveMetrics.value.recentPackets && comprehensiveMetrics.value.recentPackets.length > 0) {
      recentActivity.value = comprehensiveMetrics.value.recentPackets.slice(0, 5).map((p: any) => ({
        from_chain: p.chain_id,
        to_chain: p.dst_chain || 'Unknown', // Use actual destination chain from API
        channel: p.src_channel || p.dst_channel || 'unknown',
        status: p.effected ? 'success' : 'pending',
        timestamp: p.timestamp
      }))
    }
  }
  if (monitoringData.value) {
    // Use monitoring data to fill in any missing pieces
    if (!topRelayers.value.length && monitoringData.value.top_relayers) {
      topRelayers.value = monitoringData.value.top_relayers
    }
    if (!recentActivity.value.length && monitoringData.value.recent_activity) {
      recentActivity.value = monitoringData.value.recent_activity
    }
  }
}

// Watch for data changes
watchEffect(() => {
  updateStats()
})

// Update is handled by react-query refetch intervals


// Get chain name from config service
async function getChainName(chainId: string): Promise<string> {
  const chain = await configService.getChain(chainId)
  return chain?.chain_name || chainId
}

// For synchronous usage in template, we'll use a reactive map
const chainNames = ref<Record<string, string>>({})

// Load chain names on mount
onMounted(async () => {
  const chains = await configService.getAllChains()
  chains.forEach(chain => {
    chainNames.value[chain.chain_id] = chain.chain_name
  })
})

// Helper for template usage
function getChainNameSync(chainId: string): string {
  return chainNames.value[chainId] || chainId
}

// Extract token name from IBC denom
function extractTokenFromDenom(denom: string): string {
  // Handle IBC denoms like "transfer/channel-750/uusdc"
  if (denom.includes('/')) {
    const parts = denom.split('/')
    const token = parts[parts.length - 1]
    // Convert common denoms to readable names
    if (token === 'uusdc') return 'USDC'
    if (token === 'uatom') return 'ATOM'
    if (token === 'uosmo') return 'OSMO'
    if (token === 'ustrd') return 'STRD'
    if (token === 'utia') return 'TIA'
    if (token === 'inj') return 'INJ'
    if (token.startsWith('u')) return token.substring(1).toUpperCase()
    return token.toUpperCase()
  }
  // Handle native denoms
  if (denom === 'uusdc') return 'USDC'
  if (denom === 'uatom') return 'ATOM'
  if (denom === 'uosmo') return 'OSMO'
  if (denom.startsWith('u')) return denom.substring(1).toUpperCase()
  return denom.toUpperCase()
}

// Get destination chain from channel (simplified mapping)
function getDestChainFromChannel(dstChannel: string): string {
  // This would ideally come from a channel registry
  // For now, use common mappings
  if (dstChannel === 'channel-0') return 'cosmoshub-4'
  if (dstChannel === 'channel-1') return 'noble-1'
  if (dstChannel === 'channel-141') return 'cosmoshub-4'
  if (dstChannel === 'channel-208') return 'axelar-dojo-1'
  return 'Unknown'
}
</script>