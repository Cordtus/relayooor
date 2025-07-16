<template>
  <div class="bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-gray-200">
      <h3 class="text-lg font-medium text-gray-900">Performance Alerts</h3>
    </div>
    
    <div class="p-6">
      <!-- Active Performance Issues -->
      <div v-if="performanceIssues.length > 0" class="space-y-4">
        <div
          v-for="issue in performanceIssues"
          :key="issue.id"
          class="border rounded-lg p-4"
          :class="getAlertClass(issue.severity)"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <component :is="getAlertIcon(issue.type)" class="h-5 w-5" :class="getAlertIconClass(issue.severity)" />
                <h4 class="text-sm font-medium text-gray-900">
                  {{ issue.title }}
                </h4>
                <span :class="[
                  'px-2 py-0.5 text-xs font-medium rounded-full',
                  issue.severity === 'critical' ? 'bg-red-100 text-red-800' :
                  issue.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-blue-100 text-blue-800'
                ]">
                  {{ issue.severity }}
                </span>
              </div>
              
              <p class="mt-1 text-sm text-gray-600">{{ issue.description }}</p>
              
              <div class="mt-2 flex items-center gap-4 text-xs text-gray-500">
                <span>{{ issue.metric }}: {{ issue.currentValue }}{{ issue.unit }}</span>
                <span>Threshold: {{ issue.threshold }}{{ issue.unit }}</span>
                <span v-if="issue.trend" :class="getTrendClass(issue.trend)">
                  <TrendingUp v-if="issue.trend === 'increasing'" class="inline h-3 w-3" />
                  <TrendingDown v-else class="inline h-3 w-3" />
                  {{ issue.trendText }}
                </span>
              </div>
              
              <!-- Affected Items -->
              <div v-if="issue.affectedItems && issue.affectedItems.length > 0" class="mt-2">
                <p class="text-xs text-gray-500">Affected:</p>
                <div class="mt-1 flex flex-wrap gap-1">
                  <span
                    v-for="item in issue.affectedItems.slice(0, 5)"
                    :key="item"
                    class="px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded"
                  >
                    {{ item }}
                  </span>
                  <span v-if="issue.affectedItems.length > 5" class="px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded">
                    +{{ issue.affectedItems.length - 5 }} more
                  </span>
                </div>
              </div>
            </div>
            
            <!-- Actions -->
            <div class="ml-4 flex flex-col gap-2">
              <button
                @click="viewDetails(issue)"
                class="px-3 py-1 text-xs font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-md"
              >
                View Details
              </button>
              <button
                v-if="issue.hasAutoFix"
                @click="autoFix(issue)"
                class="px-3 py-1 text-xs font-medium text-green-700 bg-green-50 hover:bg-green-100 rounded-md"
              >
                Auto Fix
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- No Performance Issues -->
      <div v-else class="text-center py-8">
        <Zap class="mx-auto h-12 w-12 text-green-500" />
        <h3 class="mt-2 text-sm font-medium text-gray-900">Optimal Performance</h3>
        <p class="mt-1 text-sm text-gray-500">
          All channels and relayers are performing within expected parameters.
        </p>
      </div>
      
      <!-- Performance Thresholds -->
      <div class="mt-6 border-t pt-6">
        <h4 class="text-sm font-medium text-gray-900 mb-3">Performance Thresholds</h4>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Success Rate Alert</span>
              <span class="font-medium">&lt; {{ thresholds.successRate }}%</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">High Latency</span>
              <span class="font-medium">&gt; {{ thresholds.latency }}s</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Packet Backlog</span>
              <span class="font-medium">&gt; {{ thresholds.backlog }} packets</span>
            </div>
          </div>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Throughput Drop</span>
              <span class="font-medium">&lt; {{ thresholds.throughput }}% baseline</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Error Rate</span>
              <span class="font-medium">&gt; {{ thresholds.errorRate }}%</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Channel Congestion</span>
              <span class="font-medium">&gt; {{ thresholds.congestion }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { 
  Zap,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  Gauge,
  Activity,
  AlertCircle
} from 'lucide-vue-next'

interface Channel {
  channelId: string
  srcChain: string
  dstChain: string
  successRate: number
  avgProcessingTime: number
  volume24h: number
  effectedPackets: number
  uneffectedPackets: number
  totalPackets: number
}

interface PerformanceIssue {
  id: string
  type: 'low_success_rate' | 'high_latency' | 'throughput_drop' | 'high_error_rate' | 'congestion'
  severity: 'critical' | 'warning' | 'info'
  title: string
  description: string
  metric: string
  currentValue: number
  threshold: number
  unit: string
  trend?: 'increasing' | 'decreasing'
  trendText?: string
  affectedItems?: string[]
  hasAutoFix?: boolean
}

const props = defineProps<{
  channels?: Channel[]
}>()

// Performance thresholds
const thresholds = {
  successRate: 85,
  latency: 30,
  backlog: 100,
  throughput: 70,
  errorRate: 5,
  congestion: 80
}

// Calculate performance issues from channel data
const performanceIssues = computed((): PerformanceIssue[] => {
  if (!props.channels) return []
  
  const issues: PerformanceIssue[] = []
  
  // Check each channel for performance issues
  props.channels.forEach(channel => {
    // Low success rate
    if (channel.successRate < thresholds.successRate) {
      issues.push({
        id: `success-${channel.channelId}`,
        type: 'low_success_rate',
        severity: channel.successRate < 70 ? 'critical' : 'warning',
        title: 'Low Success Rate Detected',
        description: `Channel ${channel.srcChain} → ${channel.dstChain} is experiencing low success rates`,
        metric: 'Success Rate',
        currentValue: channel.successRate,
        threshold: thresholds.successRate,
        unit: '%',
        trend: 'decreasing',
        trendText: 'Declining over 1h',
        affectedItems: [channel.channelId],
        hasAutoFix: true
      })
    }
    
    // High latency
    if (channel.avgProcessingTime > thresholds.latency) {
      issues.push({
        id: `latency-${channel.channelId}`,
        type: 'high_latency',
        severity: channel.avgProcessingTime > 60 ? 'critical' : 'warning',
        title: 'High Processing Latency',
        description: `Packets taking longer than expected to process`,
        metric: 'Avg Latency',
        currentValue: channel.avgProcessingTime,
        threshold: thresholds.latency,
        unit: 's',
        trend: 'increasing',
        trendText: 'Up 25% from baseline',
        affectedItems: [channel.channelId]
      })
    }
    
    // Channel congestion
    const congestionScore = calculateCongestion(channel)
    if (congestionScore > thresholds.congestion) {
      issues.push({
        id: `congestion-${channel.channelId}`,
        type: 'congestion',
        severity: 'warning',
        title: 'Channel Congestion',
        description: `High ratio of uneffected packets indicating congestion`,
        metric: 'Congestion',
        currentValue: congestionScore,
        threshold: thresholds.congestion,
        unit: '%',
        affectedItems: [channel.channelId]
      })
    }
  })
  
  // Add some mock aggregate issues
  if (props.channels.length > 0) {
    // Overall throughput drop
    const avgThroughput = calculateAvgThroughput()
    if (avgThroughput < thresholds.throughput) {
      issues.push({
        id: 'throughput-global',
        type: 'throughput_drop',
        severity: 'warning',
        title: 'Global Throughput Drop',
        description: 'Overall packet throughput is below expected baseline',
        metric: 'Throughput',
        currentValue: avgThroughput,
        threshold: thresholds.throughput,
        unit: '% of baseline',
        trend: 'decreasing',
        trendText: 'Down from peak',
        affectedItems: ['All channels']
      })
    }
  }
  
  return issues.sort((a, b) => {
    // Sort by severity (critical first)
    const severityOrder = { critical: 0, warning: 1, info: 2 }
    return severityOrder[a.severity] - severityOrder[b.severity]
  })
})

function calculateCongestion(channel: Channel): number {
  if (!channel.totalPackets || channel.totalPackets === 0) return 0
  const uneffectedRatio = (channel.uneffectedPackets || 0) / channel.totalPackets
  return Math.round(uneffectedRatio * 100)
}

function calculateAvgThroughput(): number {
  // Mock calculation - in production would compare to historical baseline
  return Math.floor(Math.random() * 30) + 60 // Random between 60-90
}

function getAlertClass(severity: string): string {
  switch (severity) {
    case 'critical':
      return 'border-red-300 bg-red-50'
    case 'warning':
      return 'border-yellow-300 bg-yellow-50'
    case 'info':
      return 'border-blue-300 bg-blue-50'
    default:
      return 'border-gray-300'
  }
}

function getAlertIcon(type: string) {
  switch (type) {
    case 'low_success_rate':
      return AlertTriangle
    case 'high_latency':
      return Activity
    case 'throughput_drop':
      return TrendingDown
    case 'congestion':
      return Gauge
    default:
      return AlertCircle
  }
}

function getAlertIconClass(severity: string): string {
  switch (severity) {
    case 'critical':
      return 'text-red-600'
    case 'warning':
      return 'text-yellow-600'
    case 'info':
      return 'text-blue-600'
    default:
      return 'text-gray-600'
  }
}

function getTrendClass(trend: string): string {
  return trend === 'increasing' ? 'text-red-600' : 'text-yellow-600'
}

function viewDetails(issue: PerformanceIssue) {
  console.log('Viewing details for:', issue)
  // In production, this could open a modal or navigate to a details page
}

function autoFix(issue: PerformanceIssue) {
  console.log('Attempting auto-fix for:', issue)
  // In production, this could trigger automated remediation
}
</script>