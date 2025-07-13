<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-2xl font-bold text-gray-900">Advanced Analytics</h1>
      <div class="flex items-center gap-4">
        <select v-model="timeRange" class="rounded-md border-gray-300 text-sm">
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
          <option value="90d">Last 90 Days</option>
        </select>
        <button
          @click="exportData"
          class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
        >
          <Download class="h-4 w-4 mr-2" />
          Export
        </button>
      </div>
    </div>

    <!-- Key Insights -->
    <div class="bg-gradient-to-r from-purple-50 to-indigo-50 rounded-lg p-6">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">Key Insights & Predictions</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <InsightCard
          title="Peak Activity Time"
          :value="insights.peakTime"
          trend="+12%"
          description="Most active period"
          icon="Clock"
        />
        <InsightCard
          title="Optimal Route"
          :value="insights.optimalRoute"
          trend="+8%"
          description="Best success/volume ratio"
          icon="Route"
        />
        <InsightCard
          title="Rising Relayer"
          :value="insights.risingRelayer"
          trend="+45%"
          description="Fastest growing market share"
          icon="TrendingUp"
        />
        <InsightCard
          title="Congestion Risk"
          :value="insights.congestionRisk"
          :trend="insights.congestionTrend"
          description="Network health indicator"
          icon="AlertTriangle"
          :color="insights.congestionLevel"
        />
      </div>
    </div>

    <!-- Predictive Analytics -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-lg font-medium mb-4">Volume Prediction (Next 7 Days)</h3>
        <VolumePredictionChart :data="volumePrediction" />
        <div class="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <p class="text-gray-500">Expected Total</p>
            <p class="text-xl font-semibold">{{ formatNumber(volumePrediction.expectedTotal) }}</p>
          </div>
          <div>
            <p class="text-gray-500">Confidence</p>
            <p class="text-xl font-semibold">{{ volumePrediction.confidence }}%</p>
          </div>
        </div>
      </div>

      <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-lg font-medium mb-4">Success Rate Trend</h3>
        <SuccessRateTrendChart :data="successRateTrend" />
        <div class="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <p class="text-gray-500">Current Rate</p>
            <p class="text-xl font-semibold text-green-600">{{ successRateTrend.current }}%</p>
          </div>
          <div>
            <p class="text-gray-500">Projected (7d)</p>
            <p class="text-xl font-semibold">{{ successRateTrend.projected }}%</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Network Analysis -->
    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Network Flow Analysis</h2>
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Flow Sankey Diagram -->
        <div class="lg:col-span-2">
          <NetworkFlowSankey :flows="networkFlows" />
        </div>
        
        <!-- Flow Statistics -->
        <div class="space-y-4">
          <div class="bg-gray-50 rounded-lg p-4">
            <h4 class="text-sm font-medium text-gray-700 mb-2">Top Routes</h4>
            <div class="space-y-2">
              <div v-for="route in topRoutes" :key="route.id" class="flex justify-between text-sm">
                <span class="text-gray-600">{{ route.name }}</span>
                <span class="font-medium">{{ formatNumber(route.volume) }}</span>
              </div>
            </div>
          </div>
          
          <div class="bg-gray-50 rounded-lg p-4">
            <h4 class="text-sm font-medium text-gray-700 mb-2">Bottlenecks</h4>
            <div class="space-y-2">
              <div v-for="bottleneck in bottlenecks" :key="bottleneck.channel" 
                class="flex items-center justify-between text-sm">
                <span class="text-gray-600">{{ bottleneck.channel }}</span>
                <span class="text-red-600 font-medium">{{ bottleneck.congestion }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Competition Analysis -->
    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Competition Dynamics</h2>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Market Concentration -->
        <div>
          <h3 class="text-sm font-medium text-gray-700 mb-3">Market Concentration (HHI)</h3>
          <HHIChart :data="marketConcentration" />
          <p class="mt-2 text-sm text-gray-600">
            Current HHI: <span class="font-semibold">{{ marketConcentration.current }}</span>
            <span :class="[
              'ml-2',
              marketConcentration.trend > 0 ? 'text-red-600' : 'text-green-600'
            ]">
              {{ marketConcentration.trend > 0 ? '↑' : '↓' }} {{ Math.abs(marketConcentration.trend) }}%
            </span>
          </p>
        </div>
        
        <!-- Relayer Churn -->
        <div>
          <h3 class="text-sm font-medium text-gray-700 mb-3">Relayer Churn Rate</h3>
          <ChurnRateChart :data="relayerChurn" />
          <div class="mt-2 grid grid-cols-2 gap-4 text-sm">
            <div>
              <p class="text-gray-500">New Entrants</p>
              <p class="font-semibold text-green-600">+{{ relayerChurn.newEntrants }}</p>
            </div>
            <div>
              <p class="text-gray-500">Exits</p>
              <p class="font-semibold text-red-600">-{{ relayerChurn.exits }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Anomaly Detection -->
    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Anomaly Detection</h2>
      <div v-if="anomalies.length > 0" class="space-y-3">
        <div
          v-for="anomaly in anomalies"
          :key="anomaly.id"
          :class="[
            'p-4 rounded-lg border-l-4',
            anomaly.severity === 'high' ? 'bg-red-50 border-red-400' :
            anomaly.severity === 'medium' ? 'bg-yellow-50 border-yellow-400' :
            'bg-blue-50 border-blue-400'
          ]"
        >
          <div class="flex items-start">
            <AlertCircle :class="[
              'h-5 w-5 mt-0.5 mr-3',
              anomaly.severity === 'high' ? 'text-red-400' :
              anomaly.severity === 'medium' ? 'text-yellow-400' :
              'text-blue-400'
            ]" />
            <div class="flex-1">
              <h4 class="text-sm font-medium text-gray-900">{{ anomaly.title }}</h4>
              <p class="text-sm text-gray-600 mt-1">{{ anomaly.description }}</p>
              <p class="text-xs text-gray-500 mt-2">
                Detected {{ formatDistanceToNow(anomaly.timestamp) }} ago
              </p>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="text-center py-8 text-gray-500">
        <CheckCircle class="h-12 w-12 mx-auto mb-3 text-green-500" />
        <p>No anomalies detected</p>
      </div>
    </div>

    <!-- Performance Optimization Recommendations -->
    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Optimization Recommendations</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <RecommendationCard
          v-for="rec in recommendations"
          :key="rec.id"
          :recommendation="rec"
          @implement="implementRecommendation"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Download, AlertCircle, CheckCircle } from 'lucide-vue-next'
import { formatDistanceToNow } from 'date-fns'
import InsightCard from '@/components/analytics/InsightCard.vue'
import VolumePredictionChart from '@/components/analytics/VolumePredictionChart.vue'
import SuccessRateTrendChart from '@/components/analytics/SuccessRateTrendChart.vue'
import NetworkFlowSankey from '@/components/analytics/NetworkFlowSankey.vue'
import HHIChart from '@/components/analytics/HHIChart.vue'
import ChurnRateChart from '@/components/analytics/ChurnRateChart.vue'
import RecommendationCard from '@/components/analytics/RecommendationCard.vue'

// State
const timeRange = ref('7d')

// Mock data - in production, calculate from real metrics
const insights = ref({
  peakTime: '14:00-18:00 UTC',
  optimalRoute: 'Osmosis → Cosmos',
  risingRelayer: 'osmo1xyz...abc',
  congestionRisk: 'Medium',
  congestionTrend: -5,
  congestionLevel: 'warning'
})

const volumePrediction = ref({
  expectedTotal: 2450000,
  confidence: 87,
  data: generatePredictionData()
})

const successRateTrend = ref({
  current: 87.3,
  projected: 89.1,
  data: generateTrendData()
})

const networkFlows = ref(generateNetworkFlows())

const topRoutes = computed(() => [
  { id: 1, name: 'Osmosis → Cosmos Hub', volume: 1234567 },
  { id: 2, name: 'Cosmos Hub → Osmosis', volume: 987654 },
  { id: 3, name: 'Neutron → Osmosis', volume: 456789 }
])

const bottlenecks = computed(() => [
  { channel: 'channel-141', congestion: 78 },
  { channel: 'channel-0', congestion: 65 },
  { channel: 'channel-874', congestion: 45 }
])

const marketConcentration = ref({
  current: 2345,
  trend: -2.3,
  data: generateHHIData()
})

const relayerChurn = ref({
  newEntrants: 5,
  exits: 2,
  data: generateChurnData()
})

const anomalies = ref([
  {
    id: 1,
    severity: 'medium',
    title: 'Unusual frontrun pattern detected',
    description: 'Relayer osmo1abc... has been frontrun 15 times in the last hour, significantly above average',
    timestamp: new Date(Date.now() - 30 * 60 * 1000)
  },
  {
    id: 2,
    severity: 'low',
    title: 'New relayer with high volume',
    description: 'Previously unknown relayer cosmos1xyz... suddenly processing high packet volume',
    timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000)
  }
])

const recommendations = ref([
  {
    id: 1,
    type: 'channel',
    priority: 'high',
    title: 'Switch to Less Congested Route',
    description: 'Channel-141 is experiencing high congestion. Consider routing through channel-2494',
    impact: '+15% success rate',
    effort: 'Low'
  },
  {
    id: 2,
    type: 'timing',
    priority: 'medium',
    title: 'Optimize Relay Timing',
    description: 'Schedule high-volume relays during 02:00-06:00 UTC for lower competition',
    impact: '+8% efficiency',
    effort: 'Medium'
  },
  {
    id: 3,
    type: 'config',
    priority: 'low',
    title: 'Update Gas Settings',
    description: 'Current gas prices are 20% above optimal. Adjust to recommended levels',
    impact: '-20% costs',
    effort: 'Low'
  }
])

// Helper functions
function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function exportData() {
  // Implement data export functionality
  console.log('Exporting analytics data...')
}

function implementRecommendation(rec: any) {
  console.log('Implementing recommendation:', rec)
}

// Data generation functions (mock)
function generatePredictionData() {
  return Array.from({ length: 7 }, (_, i) => ({
    day: i + 1,
    predicted: Math.floor(Math.random() * 50000) + 300000,
    confidence: Math.random() * 20 + 80
  }))
}

function generateTrendData() {
  return Array.from({ length: 30 }, (_, i) => ({
    day: i + 1,
    rate: Math.random() * 10 + 85
  }))
}

function generateNetworkFlows() {
  return {
    nodes: [
      { id: 'osmosis', name: 'Osmosis' },
      { id: 'cosmos', name: 'Cosmos Hub' },
      { id: 'neutron', name: 'Neutron' }
    ],
    links: [
      { source: 'osmosis', target: 'cosmos', value: 1000 },
      { source: 'cosmos', target: 'osmosis', value: 800 },
      { source: 'neutron', target: 'osmosis', value: 400 }
    ]
  }
}

function generateHHIData() {
  return Array.from({ length: 30 }, (_, i) => ({
    day: i + 1,
    value: Math.random() * 500 + 2000
  }))
}

function generateChurnData() {
  return Array.from({ length: 12 }, (_, i) => ({
    week: i + 1,
    entries: Math.floor(Math.random() * 10),
    exits: Math.floor(Math.random() * 5)
  }))
}
</script>