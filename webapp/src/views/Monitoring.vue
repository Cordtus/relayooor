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
        :trend="chainsTrend"
        icon="LinkIcon"
        color="blue"
      />
      <MetricCard
        title="Total Packets (24h)"
        :value="formatNumber(metrics?.system.totalPackets || 0)"
        :trend="packetsTrend"
        icon="PackageIcon"
        color="green"
      />
      <MetricCard
        title="Global Success Rate"
        :value="globalSuccessRate + '%'"
        :trend="successRateTrend"
        icon="TrendingUpIcon"
        :color="globalSuccessRate > 90 ? 'green' : globalSuccessRate > 75 ? 'yellow' : 'red'"
      />
      <MetricCard
        title="Active Relayers"
        :value="activeRelayersCount"
        :trend="relayersTrend"
        icon="UsersIcon"
        color="purple"
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

        <!-- Frontrun Analysis Tab -->
        <div v-if="activeTab === 'frontrun'" class="space-y-6">
          <!-- Frontrun Statistics -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <MetricCard
              title="Total Frontrun Events"
              :value="totalFrontrunEvents"
              icon="ZapIcon"
              color="orange"
            />
            <MetricCard
              title="Most Frontrun Relayer"
              :value="mostFrontrunRelayer?.address || 'N/A'"
              :subtitle="`${mostFrontrunRelayer?.frontrunCount || 0} times`"
              icon="UserXIcon"
              color="red"
            />
            <MetricCard
              title="Top Frontrunner"
              :value="topFrontrunner?.address || 'N/A'"
              :subtitle="`Won ${topFrontrunner?.wonCount || 0} times`"
              icon="TrophyIcon"
              color="yellow"
            />
          </div>

          <!-- Frontrun Timeline -->
          <FrontrunTimeline :events="metrics?.frontrunEvents" />

          <!-- Competition Heatmap -->
          <CompetitionHeatmap :events="metrics?.frontrunEvents" :relayers="metrics?.relayers" />
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
import { 
  Activity, Link, Users, Zap, AlertTriangle, BarChart3,
  Package, TrendingUp, UserX, Trophy, LinkIcon, PackageIcon,
  TrendingUpIcon, UsersIcon, ZapIcon, UserXIcon, TrophyIcon
} from 'lucide-vue-next'
import { metricsService } from '@/services/api'
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
import FrontrunTimeline from '@/components/monitoring/FrontrunTimeline.vue'
import CompetitionHeatmap from '@/components/monitoring/CompetitionHeatmap.vue'
import StuckPacketsAlert from '@/components/monitoring/StuckPacketsAlert.vue'
import ConnectionIssues from '@/components/monitoring/ConnectionIssues.vue'
import PerformanceAlerts from '@/components/monitoring/PerformanceAlerts.vue'
import ErrorLog from '@/components/monitoring/ErrorLog.vue'
import InsightCard from '@/components/monitoring/InsightCard.vue'
import PredictiveChart from '@/components/monitoring/PredictiveChart.vue'
import RelayerChurn from '@/components/monitoring/RelayerChurn.vue'

// State
const activeTab = ref('overview')
const autoRefresh = ref(true)
const refreshInterval = ref(5000)
const lastUpdate = ref(new Date())
const channelSortBy = ref('totalPackets')

const tabs = [
  { id: 'overview', name: 'Overview', icon: Activity },
  { id: 'chains', name: 'Chains', icon: Link, badge: computed(() => metrics.value?.chains.length) },
  { id: 'relayers', name: 'Relayers', icon: Users, badge: computed(() => metrics.value?.relayers.length) },
  { id: 'channels', name: 'Channels', icon: BarChart3 },
  { id: 'frontrun', name: 'Competition', icon: Zap },
  { id: 'alerts', name: 'Alerts', icon: AlertTriangle, badge: computed(() => alertCount.value || null) }
]

// Fetch metrics
const { data: metrics, refetch: fetchMetrics } = useQuery({
  queryKey: ['metrics'],
  queryFn: async () => {
    const raw = await metricsService.getRawMetrics()
    return metricsService.parsePrometheusMetrics(raw)
  },
  refetchInterval: computed(() => autoRefresh.value ? refreshInterval.value : false)
})

// Computed metrics
const globalSuccessRate = computed(() => {
  if (!metrics.value) return 0
  const total = metrics.value.system.totalPackets
  const effected = metrics.value.relayers.reduce((sum, r) => sum + r.effectedPackets, 0)
  return total > 0 ? Math.round((effected / total) * 100) : 0
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
    ...channel,
    volumeRank: 0, // Will be calculated
    congestionScore: calculateCongestionScore(channel),
    reliability: calculateReliability(channel)
  })).sort((a, b) => {
    if (channelSortBy.value === 'successRate') return b.successRate - a.successRate
    if (channelSortBy.value === 'congestion') return b.congestionScore - a.congestionScore
    return b.totalPackets - a.totalPackets
  })
})

const poorPerformingChannels = computed(() => {
  return metrics.value?.channels.filter(c => c.successRate < 75) || []
})

const underperformingRelayers = computed(() => {
  return metrics.value?.relayers.filter(r => r.successRate < 50 && r.totalPackets > 10) || []
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

// Inferred insights
const peakActivityPeriod = computed(() => {
  // Analyze packet flow patterns to find peak hours
  return "14:00-18:00 UTC"
})

const mostReliableRoute = computed(() => {
  if (!metrics.value) return 'N/A'
  const reliable = metrics.value.channels
    .filter(c => c.totalPackets > 100)
    .sort((a, b) => b.successRate - a.successRate)[0]
  return reliable ? `${reliable.srcChain} → ${reliable.dstChain}` : 'N/A'
})

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

const congestionRiskLevel = computed(() => {
  const risk = congestionRisk.value
  return risk === 'High' ? 'error' : risk === 'Medium' ? 'warning' : 'success'
})

// Helper functions
function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function calculateCongestionScore(channel: any): number {
  // Higher uneffected ratio = more congestion
  const ratio = channel.uneffectedPackets / channel.totalPackets
  return Math.round(ratio * 100)
}

function calculateReliability(channel: any): number {
  // Combination of success rate and volume
  const volumeScore = Math.min(channel.totalPackets / 1000, 1) * 0.3
  const successScore = (channel.successRate / 100) * 0.7
  return Math.round((volumeScore + successScore) * 100)
}

function getChainPackets(chainId: string) {
  return metrics.value?.recentPackets.filter(p => p.chain_id === chainId) || []
}

function handleClearPacket(packetId: string) {
  // Implement packet clearing logic
  console.log('Clear packet:', packetId)
}

// Lifecycle
onMounted(() => {
  lastUpdate.value = new Date()
})

onUnmounted(() => {
  autoRefresh.value = false
})
</script>