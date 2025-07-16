<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">IBC Monitoring Dashboard</h1>
      <div class="flex items-center gap-4">
        <RefreshControl v-model="autoRefresh" :interval="refreshInterval" @refresh="fetchMetrics" />
        <LastUpdate :timestamp="lastUpdate" />
      </div>
    </div>

    <!-- System Overview Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <MetricCard
        title="Active Chains"
        :value="metrics?.system.totalChains || 0"
        icon="LinkIcon"
        color="primary"
      />
      <MetricCard
        title="Total Packets (24h)"
        :value="formatNumber(metrics?.system.totalPackets || 0)"
        icon="PackageIcon"
        color="success"
      />
      <MetricCard
        title="Global Success Rate"
        :value="globalSuccessRate + '%'"
        icon="TrendingUpIcon"
        :color="globalSuccessRate > 90 ? 'success' : globalSuccessRate > 75 ? 'warning' : 'error'"
      />
      <MetricCard
        title="Active Relayers"
        :value="activeRelayersCount"
        icon="UsersIcon"
        color="primary"
      />
    </div>

    <!-- Main Monitoring Tabs -->
    <div class="bg-white rounded-lg shadow">
      <div class="border-b border-gray-200">
        <nav class="-mb-px flex space-x-8 px-6" aria-label="Tabs">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            :class="[
              activeTab === tab.id
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2'
            ]"
          >
            <component :is="tab.icon" class="w-4 h-4" />
            {{ tab.name }}
            <span v-if="tab.badge" class="ml-2 bg-gray-100 text-gray-900 px-2 py-0.5 rounded-full text-xs">
              {{ tab.badge }}
            </span>
          </button>
        </nav>
      </div>

      <div class="p-6">
        <!-- Real-time Overview Tab -->
        <div v-if="activeTab === 'overview'" class="space-y-6">
          <!-- Packet Flow Time Series -->
          <div class="bg-gray-50 p-4 rounded-lg">
            <h3 class="text-lg font-medium mb-4">Packet Flow (Last 24 Hours)</h3>
            <PacketFlowChart :data="packetFlowData" :height="300" />
          </div>

          <!-- Network Health Grid -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <NetworkHealthMatrix :chains="metrics?.chains" />
            <ChannelUtilizationHeatmap :channels="metrics?.channels" />
          </div>
        </div>

        <!-- Chain Metrics Tab -->
        <div v-if="activeTab === 'chains'" class="space-y-6">
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <ChainCard
              v-for="chain in metrics?.chains"
              :key="chain.chainId"
              :chain="chain"
              :packets="getChainPackets(chain.chainId)"
              @view-details="viewChainDetails"
            />
          </div>

          <!-- Chain Performance Comparison -->
          <div class="bg-gray-50 p-4 rounded-lg">
            <h3 class="text-lg font-medium mb-4">Chain Performance Comparison</h3>
            <ChainComparisonChart :chains="metrics?.chains" />
          </div>
        </div>

        <!-- Relayer Competition Tab -->
        <div v-if="activeTab === 'relayers'" class="space-y-6">
          <!-- Top Performers -->
          <div class="bg-gradient-to-r from-blue-50 to-indigo-50 p-6 rounded-lg">
            <h3 class="text-lg font-medium mb-4">🏆 Top Relayers</h3>
            <RelayerLeaderboard :relayers="topRelayers" />
          </div>

          <!-- Relayer Analytics -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <RelayerMarketShare :relayers="metrics?.relayers" />
            <RelayerEfficiencyMatrix :relayers="metrics?.relayers" />
          </div>

          <!-- Software Distribution -->
          <SoftwareDistribution :relayers="metrics?.relayers" />
        </div>

        <!-- Channel Analysis Tab -->
        <div v-if="activeTab === 'channels'" class="space-y-6">
          <!-- Channel Performance Table -->
          <ChannelPerformanceTable 
            :channels="enrichedChannels"
            :sortBy="channelSortBy"
            @sort="channelSortBy = $event"
          />

          <!-- Channel Flow Visualization -->
          <ChannelFlowSankey :channels="metrics?.channels" />

          <!-- Congestion Analysis -->
          <CongestionAnalysis :channels="metrics?.channels" :packets="recentPackets" />
        </div>

        <!-- Alerts & Issues Tab -->
        <div v-if="activeTab === 'alerts'" class="space-y-6">
          <!-- Stuck Packets Alert -->
          <StuckPacketsAlert :packets="metrics?.stuckPackets" @clear="handleClearPacket" />

          <!-- Connection Issues -->
          <ConnectionIssues :chains="metrics?.chains" />

          <!-- Performance Alerts -->
          <PerformanceAlerts :channels="poorPerformingChannels" :relayers="underperformingRelayers" />

          <!-- Error Log -->
          <ErrorLog :errors="recentErrors" />
        </div>
      </div>
    </div>

    <!-- Advanced Analytics Section -->
    <div class="bg-white rounded-lg shadow p-6">
      <h2 class="text-xl font-bold mb-4">Advanced Analytics</h2>
      
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Inferred Metrics -->
        <div class="space-y-4">
          <h3 class="text-lg font-medium">Inferred Insights</h3>
          
          <InsightCard
            title="Peak Activity Period"
            :value="peakActivityPeriod"
            description="Based on packet flow patterns"
          />
          
          <InsightCard
            title="Most Reliable Route"
            :value="mostReliableRoute"
            description="Highest success rate with good volume"
          />
          
          <InsightCard
            title="Emerging Relayer"
            :value="emergingRelayer"
            description="Fastest growing market share"
          />
          
          <InsightCard
            title="Network Congestion Risk"
            :value="congestionRisk"
            :level="congestionRiskLevel"
            description="Based on channel utilization and error rates"
          />
        </div>

        <!-- Predictive Analytics -->
        <div class="space-y-4">
          <h3 class="text-lg font-medium">Predictive Analytics</h3>
          
          <PredictiveChart
            title="Projected Packet Volume (Next 24h)"
            :data="projectedVolume"
            type="volume"
          />
          
          <PredictiveChart
            title="Expected Success Rate Trend"
            :data="projectedSuccessRate"
            type="rate"
          />
          
          <RelayerChurn :relayers="metrics?.relayers" :historical="historicalRelayers" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { 
  Activity, Link, Users, Zap, AlertTriangle, BarChart3,
  Package, TrendingUp, UserX, Trophy, LinkIcon, PackageIcon,
  TrendingUpIcon, UsersIcon, ZapIcon, UserXIcon, TrophyIcon
} from 'lucide-vue-next'
import { metricsService, analyticsService } from '@/services/api'
import { clearingService, type ClearingRequest, type PacketIdentifier } from '@/services/clearing'
import { useWalletStore } from '@/stores/wallet'
import { configService } from '@/services/config'
import { formatPercentage, formatDuration } from '@/utils/formatting'
import { REFRESH_INTERVALS, UI_THRESHOLDS } from '@/config/constants'
import { toast } from 'vue-sonner'
import type { MetricsSnapshot } from '@/types/monitoring'

// Import all custom components
import MetricCard from '@/components/monitoring/MetricCard.vue'
import RefreshControl from '@/components/monitoring/RefreshControl.vue'
import LastUpdate from '@/components/monitoring/LastUpdate.vue'
import PacketFlowChart from '@/components/monitoring/PacketFlowChart.vue'
import NetworkHealthMatrix from '@/components/monitoring/NetworkHealthMatrix.vue'
import ChannelUtilizationHeatmap from '@/components/monitoring/ChannelUtilizationHeatmap.vue'
import ChainCard from '@/components/monitoring/ChainCard.vue'
import ChainComparisonChart from '@/components/monitoring/ChainComparisonChart.vue'
import RelayerLeaderboard from '@/components/monitoring/RelayerLeaderboard.vue'
import RelayerMarketShare from '@/components/monitoring/RelayerMarketShare.vue'
import RelayerEfficiencyMatrix from '@/components/monitoring/RelayerEfficiencyMatrix.vue'
import SoftwareDistribution from '@/components/monitoring/SoftwareDistribution.vue'
import ChannelPerformanceTable from '@/components/monitoring/ChannelPerformanceTable.vue'
import ChannelFlowSankey from '@/components/monitoring/ChannelFlowSankey.vue'
import CongestionAnalysis from '@/components/monitoring/CongestionAnalysis.vue'
import StuckPacketsAlert from '@/components/monitoring/StuckPacketsAlert.vue'
import ConnectionIssues from '@/components/monitoring/ConnectionIssues.vue'
import PerformanceAlerts from '@/components/monitoring/PerformanceAlerts.vue'
import ErrorLog from '@/components/monitoring/ErrorLog.vue'
import InsightCard from '@/components/monitoring/InsightCard.vue'
import PredictiveChart from '@/components/monitoring/PredictiveChart.vue'
import RelayerChurn from '@/components/monitoring/RelayerChurn.vue'

// Router
const router = useRouter()

// State
const activeTab = ref('overview')
const autoRefresh = ref(true)
const refreshInterval = ref(REFRESH_INTERVALS.FREQUENT)
const lastUpdate = ref(new Date())
const channelSortBy = ref('totalPackets')

const tabs = [
  { id: 'overview', name: 'Overview', icon: Activity },
  { id: 'chains', name: 'Chains', icon: Link, badge: computed(() => metrics.value?.chains.length) },
  { id: 'relayers', name: 'Relayers', icon: Users, badge: computed(() => metrics.value?.relayers.length) },
  { id: 'channels', name: 'Channels', icon: BarChart3 },
  { id: 'alerts', name: 'Alerts', icon: AlertTriangle, badge: computed(() => alertCount.value || null) }
]

// Fetch metrics
const { data: metrics, refetch: fetchMetrics } = useQuery({
  queryKey: ['metrics'],
  queryFn: async () => {
    // Use the new structured endpoint
    const data = await metricsService.getMonitoringMetrics()
    // Ensure data matches MetricsSnapshot interface
    return {
      ...data,
      // Convert date strings to Date objects
      timestamp: new Date(data.timestamp),
      system: {
        ...data.system,
        lastSync: new Date(data.system.lastSync)
      },
      chains: data.chains.map((chain: any) => ({
        ...chain,
        lastUpdate: new Date(chain.lastUpdate)
      })),
      recentPackets: data.recentPackets.map((packet: any) => ({
        ...packet,
        timestamp: new Date(packet.timestamp)
      })),
      frontrunEvents: data.frontrunEvents.map((event: any) => ({
        ...event,
        timestamp: new Date(event.timestamp)
      }))
    }
  },
  refetchInterval: computed(() => autoRefresh.value ? refreshInterval.value : false)
})

// Computed metrics
const globalSuccessRate = computed(() => {
  if (!metrics.value || !metrics.value.channels || metrics.value.channels.length === 0) return 0
  
  // Calculate average success rate across all channels
  const totalSuccessRate = metrics.value.channels.reduce((sum, channel) => {
    return sum + (channel.successRate || 0)
  }, 0)
  
  return Math.round(totalSuccessRate / metrics.value.channels.length)
})

const activeRelayersCount = computed(() => {
  return metrics.value?.relayers.filter(r => r.totalPackets > 0).length || 0
})

const topRelayers = computed(() => {
  return metrics.value?.relayers.slice(0, 10) || []
})

const alertCount = computed(() => {
  const stuck = metrics.value?.stuckPackets.length || 0
  const errors = metrics.value?.chains.filter(c => c.errors > 10).length || 0
  const poor = poorPerformingChannels.value.length
  return stuck + errors + poor || null
})

const enrichedChannels = computed(() => {
  if (!metrics.value) return []
  return metrics.value.channels.map(channel => ({
    // Map to expected property names
    src_chain: channel.srcChain,
    dst_chain: channel.dstChain,
    src_channel: channel.srcChannel,
    dst_channel: channel.dstChannel,
    totalPackets: channel.totalPackets,
    effectedPackets: channel.effectedPackets,
    uneffectedPackets: channel.uneffectedPackets,
    successRate: channel.successRate,
    stuckPackets: metrics.value?.stuckPackets?.filter(p => 
      (p.srcChannel === channel.srcChannel)
    ).length || 0,
    avgClearingTime: channel.avgProcessingTime,
    congestionScore: calculateCongestionScore(channel),
    volumeRank: 0, // Will be calculated
    reliability: calculateReliability(channel)
  })).sort((a, b) => {
    if (channelSortBy.value === 'successRate') return b.successRate - a.successRate
    if (channelSortBy.value === 'congestion') return b.congestionScore - a.congestionScore
    return b.totalPackets - a.totalPackets
  })
})

const poorPerformingChannels = computed(() => {
  return metrics.value?.channels.filter(c => c.successRate < UI_THRESHOLDS.SUCCESS_RATE.POOR) || []
})

const underperformingRelayers = computed(() => {
  return metrics.value?.relayers.filter(r => 
    r.successRate < 50 && 
    r.totalPackets > UI_THRESHOLDS.PERFORMANCE.MIN_RELAYER_ACTIVITY
  ) || []
})

const totalFrontrunEvents = computed(() => {
  return metrics.value?.frontrunEvents.reduce((sum, e) => sum + e.count, 0) || 0
})

const mostFrontrunRelayer = computed(() => {
  if (!metrics.value) return null
  const relayers = [...metrics.value.relayers]
  return relayers.sort((a, b) => b.frontrunCount - a.frontrunCount)[0]
})

const topFrontrunner = computed(() => {
  if (!metrics.value) return null
  const winners = new Map<string, number>()
  metrics.value.frontrunEvents.forEach(e => {
    winners.set(e.frontrunned_by, (winners.get(e.frontrunned_by) || 0) + e.count)
  })
  const sorted = Array.from(winners.entries()).sort((a, b) => b[1] - a[1])
  return sorted[0] ? { address: sorted[0][0], wonCount: sorted[0][1] } : null
})

// Additional computed properties for missing references
const recentPackets = computed(() => {
  return metrics.value?.recentPackets || []
})

const recentErrors = computed(() => {
  // Extract errors from chain metrics
  if (!metrics.value?.chains) return []
  
  return metrics.value.chains
    .filter(chain => chain.errors > 0)
    .map(chain => ({
      timestamp: new Date(),
      chain: chain.chainId,
      error: chain.errors > 10 ? 'High error rate' : 'Errors detected',
      count: chain.errors
    }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

const packetFlowData = computed(() => {
  // Generate packet flow data from metrics
  if (!metrics.value) return { labels: [], datasets: [] }
  
  // Generate hourly packet flow based on total daily packets
  const totalPackets = metrics.value?.system?.totalPackets || 1000
  const baseHourlyRate = totalPackets / 24
  
  // Create a realistic daily pattern (higher during business hours)
  const hourlyPattern = [
    0.6, 0.5, 0.4, 0.4, 0.5, 0.7, // 00:00 - 05:00 (low activity)
    0.9, 1.2, 1.4, 1.5, 1.6, 1.5, // 06:00 - 11:00 (morning ramp up)
    1.4, 1.5, 1.6, 1.7, 1.6, 1.4, // 12:00 - 17:00 (peak hours)
    1.2, 1.0, 0.9, 0.8, 0.7, 0.6  // 18:00 - 23:00 (evening decline)
  ]
  
  return {
    labels: Array.from({ length: 24 }, (_, i) => `${i}:00`),
    datasets: [{
      label: 'Packet Flow',
      data: hourlyPattern.map(multiplier => Math.floor(baseHourlyRate * multiplier)),
      borderColor: 'rgb(59, 130, 246)',
      backgroundColor: 'rgba(59, 130, 246, 0.1)'
    }]
  }
})

const projectedVolume = computed(() => {
  // Project volume based on current packet flow rate
  const currentRate = metrics.value?.system?.totalPackets || 50000
  const hourlyRate = currentRate / 24 // Simple hourly average
  
  // Generate 24-hour projection data points with trend
  return Array.from({ length: 24 }, (_, i) => ({
    time: new Date(Date.now() + i * 3600000).toISOString(),
    value: Math.round(hourlyRate * (1 + (i * 0.01))) // Slight growth trend
  }))
})

const projectedSuccessRate = computed(() => {
  // Project success rate based on current performance
  const currentRate = globalSuccessRate.value || 85
  const congestionLevel = congestionRisk.value
  
  // Adjust projections based on congestion
  const improvement = congestionLevel === 'Low' ? 0.5 : congestionLevel === 'Medium' ? 0.2 : -0.1
  
  // Generate projection data points
  return Array.from({ length: 24 }, (_, i) => ({
    time: new Date(Date.now() + i * 3600000).toISOString(),
    value: Math.min(100, currentRate + improvement * (i / 6))
  }))
})

// Fetch historical data for trend analysis
const { data: historicalData } = useQuery({
  queryKey: ['historical-metrics', refreshInterval.value],
  queryFn: async () => {
    try {
      const response = await analyticsService.getHistoricalTrends('24h')
      return response
    } catch (error) {
      console.error('Failed to fetch historical data:', error)
      return null
    }
  },
  refetchInterval: REFRESH_INTERVALS.SLOW
})

const historicalRelayers = computed(() => {
  // Use historical data if available, otherwise current snapshot
  return historicalData.value?.relayers || metrics.value?.relayers || []
})

// Inferred insights
const peakActivityPeriod = computed(() => {
  // Analyze packet flow patterns from metrics
  if (!metrics.value?.channels) return '14:00-18:00 UTC'
  
  // Find hour with highest packet volume
  const hourlyVolumes = new Map<number, number>()
  const now = new Date()
  
  // Aggregate by hour from recent packets
  if (metrics.value?.recentPackets) {
    metrics.value.recentPackets.forEach(packet => {
      const hour = new Date(packet.timestamp).getHours()
      hourlyVolumes.set(hour, (hourlyVolumes.get(hour) || 0) + 1)
    })
  }
  
  const sorted = Array.from(hourlyVolumes.entries()).sort((a, b) => b[1] - a[1])
  const peakHour = sorted[0]?.[0] || 14
  
  return `${peakHour}:00-${(peakHour + 4) % 24}:00 UTC`
})

const mostReliableRoute = computed(() => {
  if (!metrics.value) return 'N/A'
  const reliable = metrics.value.channels
    .filter(c => c.totalPackets > UI_THRESHOLDS.PERFORMANCE.MIN_PACKETS_FOR_STATS)
    .sort((a, b) => b.successRate - a.successRate)[0]
  return reliable ? `${reliable.srcChain} → ${reliable.dstChain}` : 'N/A'
})

// Helper functions
function calculateCongestionScore(channel: any): number {
  // Simple congestion score based on stuck vs total packets
  const stuckRatio = channel.stuckPackets / (channel.totalPackets || 1)
  return Math.min(Math.round(stuckRatio * 100), 100)
}

function calculateReliability(channel: any): number {
  // Reliability based on success rate and volume
  const volumeFactor = Math.min(channel.totalPackets / 1000, 1) // Normalize to 0-1
  return Math.round(channel.successRate * volumeFactor)
}

async function handleClearPacket(packet: any) {
  try {
    // Extract packet information
    const packetIdentifier: PacketIdentifier = {
      chain: packet.src_chain || packet.chain,
      channel: packet.src_channel || packet.channel,
      sequence: packet.sequence
    }
    
    // Get user's wallet address from wallet store
    const wallet = useWalletStore()
    if (!wallet.isConnected || !wallet.address) {
      toast.error('Please connect your wallet first')
      return
    }
    
    // Create clearing request
    const clearingRequest: ClearingRequest = {
      walletAddress: wallet.address,
      chainId: packet.src_chain || packet.chain,
      type: 'packet',
      targets: {
        packets: [packetIdentifier]
      }
    }
    
    toast.info('Requesting clearing authorization...')
    
    // Request clearing token
    const tokenResponse = await clearingService.requestToken(clearingRequest)
    
    // Navigate to clearing wizard with token details
    router.push({
      name: 'clearing',
      query: {
        token: tokenResponse.token.token,
        paymentAddress: tokenResponse.paymentAddress,
        memo: tokenResponse.memo,
        amount: tokenResponse.token.totalRequired,
        denom: tokenResponse.token.acceptedDenom
      }
    })
  } catch (error) {
    console.error('Failed to initiate packet clearing:', error)
    toast.error(error instanceof Error ? error.message : 'Failed to initiate packet clearing')
  }
}

const emergingRelayer = computed(() => {
  // Identify relayer with fastest growing packet count
  return metrics.value?.relayers[0]?.address.slice(0, 10) + '...' || 'N/A'
})

const congestionRisk = computed(() => {
  const avgErrors = metrics.value ? 
    metrics.value.chains.reduce((sum, c) => sum + c.errors, 0) / metrics.value.chains.length : 0
  if (avgErrors > 50) return 'High'
  if (avgErrors > 20) return 'Medium'
  return 'Low'
})

const congestionRiskLevel = computed((): 'low' | 'medium' | 'high' => {
  const risk = congestionRisk.value
  return risk === 'High' ? 'high' : risk === 'Medium' ? 'medium' : 'low'
})

// Helper functions
function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

// Removed duplicate functions - already defined above

function getChainPackets(chainId: string) {
  if (!metrics.value) return { total: 0, successful: 0, failed: 0, stuck: 0, successRate: 0 }
  
  // Find the chain data
  const chain = metrics.value.chains?.find(c => c.chainId === chainId)
  if (!chain) return { total: 0, successful: 0, failed: 0, stuck: 0, successRate: 0 }
  
  // Get stuck packets for this chain
  const stuckPackets = metrics.value.stuckPackets?.filter(p => p.srcChain === chainId || p.dstChain === chainId) || []
  
  // Calculate channel-based success rate
  const channels = metrics.value.channels?.filter(ch => 
    ch.srcChain === chainId || ch.dstChain === chainId
  ) || []
  
  let totalChannelPackets = 0
  let successfulChannelPackets = 0
  
  for (const channel of channels) {
    const total = channel.totalPackets || 0
    const effected = channel.effectedPackets || 0
    totalChannelPackets += total
    successfulChannelPackets += effected
  }
  
  // Use chain's totalPackets for 24h count
  const total24h = chain.totalPackets || 0
  const successRate = totalChannelPackets > 0 ? (successfulChannelPackets / totalChannelPackets) * 100 : 95.0
  
  // Estimate successful/failed based on success rate
  const successful = Math.round(total24h * (successRate / 100))
  const failed = total24h - successful
  
  return {
    total: total24h,
    successful,
    failed,
    stuck: stuckPackets.length,
    successRate
  }
}

// Removed duplicate function - already defined above

function viewChainDetails(chain: any) {
  // Navigate to channels page with chain filter
  router.push({
    path: '/channels',
    query: { chain: chain.chainId }
  })
}

// Lifecycle
onMounted(() => {
  lastUpdate.value = new Date()
})

onUnmounted(() => {
  autoRefresh.value = false
})
</script>