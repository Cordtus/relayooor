<template>
  <div class="bg-white rounded-lg shadow p-4">
    <h3 class="text-lg font-medium mb-4">Channel Flow Visualization</h3>
    <div v-if="flowData.nodes.length > 0">
      <div class="h-96">
        <canvas ref="sankeyChart"></canvas>
      </div>
      
      <!-- Flow Summary -->
      <div class="mt-4 pt-4 border-t">
        <h4 class="text-sm font-medium text-gray-600 mb-2">Top Flows</h4>
        <div class="space-y-1">
          <div v-for="flow in topFlows" :key="`${flow.source}-${flow.target}`" 
               class="flex justify-between items-center text-sm">
            <span>{{ flow.source }} → {{ flow.target }}</span>
            <span class="font-medium">{{ formatNumber(flow.value) }} packets</span>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="text-center py-8 text-gray-500">
      <p>No channel flow data available</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Chart from 'chart.js/auto'
import type { ChannelMetrics } from '@/types/monitoring'

interface Props {
  channels?: ChannelMetrics[]
}

const props = defineProps<Props>()

const sankeyChart = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// Transform channel data into sankey format
const flowData = computed(() => {
  if (!props.channels || props.channels.length === 0) {
    return { nodes: [], links: [] }
  }
  
  const nodeSet = new Set<string>()
  const links: Array<{ source: string; target: string; value: number }> = []
  
  // Aggregate flows between chains
  const flowMap = new Map<string, number>()
  
  props.channels.forEach(channel => {
    const source = getChainShortName(channel.srcChain)
    const target = channel.dstChain && channel.dstChain !== 'unknown' 
      ? getChainShortName(channel.dstChain) 
      : 'Unknown'
    
    if (target === 'Unknown') return
    
    nodeSet.add(source)
    nodeSet.add(target)
    
    const key = `${source}-${target}`
    flowMap.set(key, (flowMap.get(key) || 0) + channel.totalPackets)
  })
  
  // Convert to links array
  flowMap.forEach((value, key) => {
    const [source, target] = key.split('-')
    links.push({ source, target, value })
  })
  
  // Sort links by value
  links.sort((a, b) => b.value - a.value)
  
  return {
    nodes: Array.from(nodeSet).map(id => ({ id })),
    links
  }
})

const topFlows = computed(() => {
  return flowData.value.links.slice(0, 5)
})

function getChainShortName(chainId: string): string {
  const names: Record<string, string> = {
    'cosmoshub-4': 'Cosmos',
    'osmosis-1': 'Osmosis',
    'neutron-1': 'Neutron',
    'noble-1': 'Noble',
    'axelar-dojo-1': 'Axelar',
    'stride-1': 'Stride'
  }
  return names[chainId] || chainId
}

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

function createChart() {
  if (!sankeyChart.value || flowData.value.nodes.length === 0) return
  
  const ctx = sankeyChart.value.getContext('2d')
  if (!ctx) return
  
  // For now, create a simple flow visualization using a horizontal bar chart
  // (Chart.js doesn't have native sankey support)
  const labels = flowData.value.links.map(l => `${l.source} → ${l.target}`)
  const data = flowData.value.links.map(l => l.value)
  
  chart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [{
        label: 'Packet Flow',
        data: data,
        backgroundColor: data.map((_, i) => {
          const colors = [
            'rgba(59, 130, 246, 0.8)',
            'rgba(16, 185, 129, 0.8)',
            'rgba(245, 158, 11, 0.8)',
            'rgba(239, 68, 68, 0.8)',
            'rgba(139, 92, 246, 0.8)'
          ]
          return colors[i % colors.length]
        }),
        borderColor: data.map((_, i) => {
          const colors = [
            'rgb(59, 130, 246)',
            'rgb(16, 185, 129)',
            'rgb(245, 158, 11)',
            'rgb(239, 68, 68)',
            'rgb(139, 92, 246)'
          ]
          return colors[i % colors.length]
        }),
        borderWidth: 1
      }]
    },
    options: {
      indexAxis: 'y',
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              return `${formatNumber(context.parsed.x)} packets`
            }
          }
        }
      },
      scales: {
        x: {
          beginAtZero: true,
          grid: {
            display: true,
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            callback: function(value) {
              return formatNumber(value as number)
            }
          }
        },
        y: {
          grid: {
            display: false
          }
        }
      }
    }
  })
}

onMounted(() => {
  createChart()
})

watch(() => props.channels, () => {
  if (chart) {
    chart.destroy()
  }
  createChart()
})
</script>