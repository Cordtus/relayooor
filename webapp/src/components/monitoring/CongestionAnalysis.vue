<template>
  <div class="bg-white rounded-lg shadow p-4">
    <h3 class="text-lg font-medium mb-4">Congestion Analysis</h3>
    
    <div v-if="congestionData.length > 0">
      <!-- Congestion Timeline -->
      <div class="mb-6">
        <h4 class="text-sm font-medium text-gray-600 mb-2">Congestion Over Time</h4>
        <div class="h-48">
          <canvas ref="congestionChart"></canvas>
        </div>
      </div>
      
      <!-- Congested Channels -->
      <div class="mb-4">
        <h4 class="text-sm font-medium text-gray-600 mb-2">Most Congested Channels</h4>
        <div class="space-y-2">
          <div v-for="channel in congestedChannels" :key="`${channel.srcChannel}-${channel.dstChannel}`"
               class="border rounded p-3">
            <div class="flex justify-between items-start">
              <div>
                <p class="font-medium text-sm">
                  {{ channel.srcChain }} → {{ channel.dstChain || 'Unknown' }}
                </p>
                <p class="text-xs text-gray-500">
                  {{ channel.srcChannel }} → {{ channel.dstChannel }}
                </p>
              </div>
              <div class="text-right">
                <p class="text-sm font-medium" :class="getCongestionColorClass(channel.congestionScore)">
                  {{ channel.congestionScore }}% congested
                </p>
                <p class="text-xs text-gray-500">
                  {{ channel.stuckPackets }} stuck / {{ channel.totalPackets }} total
                </p>
              </div>
            </div>
            
            <!-- Congestion Factors -->
            <div class="mt-2 flex flex-wrap gap-2">
              <span v-for="factor in getCongestionFactors(channel)" :key="factor"
                    class="text-xs px-2 py-1 bg-gray-100 rounded">
                {{ factor }}
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Congestion Insights -->
      <div class="bg-blue-50 rounded p-3">
        <h4 class="text-sm font-medium text-blue-900 mb-2">Insights</h4>
        <ul class="text-sm text-blue-800 space-y-1">
          <li v-for="insight in congestionInsights" :key="insight" class="flex items-start">
            <span class="mr-2">•</span>
            <span>{{ insight }}</span>
          </li>
        </ul>
      </div>
    </div>
    
    <div v-else class="text-center py-8 text-gray-500">
      <p>No congestion detected</p>
      <p class="text-sm mt-2">All channels operating normally</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Chart from 'chart.js/auto'
import type { ChannelMetrics } from '@/types/monitoring'

interface Props {
  channels?: ChannelMetrics[]
  packets?: any[]
}

const props = defineProps<Props>()

const congestionChart = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// Calculate congestion data from recent packets
const congestionData = computed(() => {
  if (!props.packets || props.packets.length === 0) return []
  
  // Group packets by hour to show congestion trends
  const hourlyData = new Map<string, { total: number, stuck: number }>()
  const now = new Date()
  
  for (let i = 0; i < 24; i++) {
    const hour = new Date(now.getTime() - i * 3600000)
    const key = hour.toISOString().slice(0, 13) // YYYY-MM-DDTHH
    hourlyData.set(key, { total: 0, stuck: 0 })
  }
  
  // Count packets per hour
  props.packets.forEach(packet => {
    const hour = new Date(packet.timestamp).toISOString().slice(0, 13)
    if (hourlyData.has(hour)) {
      const data = hourlyData.get(hour)!
      data.total++
      if (!packet.effected) {
        data.stuck++
      }
    }
  })
  
  // Convert to array and calculate congestion percentage
  return Array.from(hourlyData.entries())
    .map(([hour, data]) => ({
      hour,
      congestion: data.total > 0 ? (data.stuck / data.total) * 100 : 0,
      total: data.total,
      stuck: data.stuck
    }))
    .reverse() // Show oldest to newest
})

// Find most congested channels
const congestedChannels = computed(() => {
  if (!props.channels) return []
  
  return props.channels
    .map(channel => {
      // Calculate stuck packets for this channel
      const stuckPackets = props.packets?.filter(p => 
        p.src_channel === channel.srcChannel && !p.effected
      ).length || 0
      
      const congestionScore = channel.totalPackets > 0 
        ? Math.round((stuckPackets / channel.totalPackets) * 100)
        : 0
      
      return {
        ...channel,
        stuckPackets,
        congestionScore
      }
    })
    .filter(ch => ch.congestionScore > 5) // Only show channels with >5% congestion
    .sort((a, b) => b.congestionScore - a.congestionScore)
    .slice(0, 5)
})

// Generate congestion insights
const congestionInsights = computed(() => {
  const insights: string[] = []
  
  if (congestedChannels.value.length === 0) {
    insights.push('No significant congestion detected across channels')
    return insights
  }
  
  // Identify patterns
  const avgCongestion = congestedChannels.value.reduce((sum, ch) => sum + ch.congestionScore, 0) / congestedChannels.value.length
  insights.push(`Average congestion: ${avgCongestion.toFixed(1)}% across affected channels`)
  
  // Check for chain-specific issues
  const chainCongestion = new Map<string, number>()
  congestedChannels.value.forEach(ch => {
    chainCongestion.set(ch.srcChain, (chainCongestion.get(ch.srcChain) || 0) + 1)
  })
  
  const congestedChains = Array.from(chainCongestion.entries())
    .filter(([_, count]) => count > 1)
    .map(([chain]) => chain)
  
  if (congestedChains.length > 0) {
    insights.push(`Chains with multiple congested channels: ${congestedChains.join(', ')}`)
  }
  
  // Time-based patterns
  const recentCongestion = congestionData.value.slice(-6) // Last 6 hours
  const avgRecent = recentCongestion.reduce((sum, d) => sum + d.congestion, 0) / 6
  if (avgRecent > 10) {
    insights.push(`Elevated congestion in the last 6 hours (${avgRecent.toFixed(1)}% average)`)
  }
  
  return insights
})

function getCongestionColorClass(score: number): string {
  if (score > 50) return 'text-red-600'
  if (score > 20) return 'text-orange-600'
  if (score > 10) return 'text-yellow-600'
  return 'text-green-600'
}

function getCongestionFactors(channel: any): string[] {
  const factors: string[] = []
  
  if (channel.successRate < 90) {
    factors.push(`Low success rate (${channel.successRate.toFixed(1)}%)`)
  }
  
  if (channel.avgProcessingTime > 300) {
    factors.push('Slow processing')
  }
  
  if (channel.totalPackets > 10000) {
    factors.push('High volume')
  }
  
  if (channel.stuckPackets > 10) {
    factors.push(`${channel.stuckPackets} stuck packets`)
  }
  
  return factors
}

function createChart() {
  if (!congestionChart.value || congestionData.value.length === 0) return
  
  const ctx = congestionChart.value.getContext('2d')
  if (!ctx) return
  
  const labels = congestionData.value.map(d => {
    const hour = new Date(d.hour + ':00:00Z')
    return hour.getHours() + ':00'
  })
  
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label: 'Congestion %',
        data: congestionData.value.map(d => d.congestion),
        borderColor: 'rgb(239, 68, 68)',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        tension: 0.3,
        fill: true
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            afterLabel: function(context) {
              const dataPoint = congestionData.value[context.dataIndex]
              return `${dataPoint.stuck} stuck / ${dataPoint.total} total packets`
            }
          }
        }
      },
      scales: {
        y: {
          beginAtZero: true,
          max: 100,
          ticks: {
            callback: function(value) {
              return value + '%'
            }
          }
        }
      }
    }
  })
}

onMounted(() => {
  createChart()
})

watch(() => [props.channels, props.packets], () => {
  if (chart) {
    chart.destroy()
  }
  createChart()
})
</script>