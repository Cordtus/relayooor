<template>
  <div class="bg-white rounded-lg shadow p-4">
    <h3 class="text-lg font-medium mb-4">Market Share Distribution</h3>
    <div v-if="marketData.length > 0">
      <div class="h-64">
        <canvas ref="chartCanvas"></canvas>
      </div>
      <div class="mt-4 grid grid-cols-2 gap-2 text-sm">
        <div v-for="(item, index) in topRelayers" :key="item.address" class="flex items-center gap-2">
          <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: colors[index % colors.length] }"></div>
          <span class="text-gray-600">{{ formatAddress(item.address) }}</span>
          <span class="font-medium">{{ item.percentage.toFixed(1) }}%</span>
        </div>
      </div>
    </div>
    <div v-else class="text-center py-8 text-gray-500">
      No relayer data available
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Relayer {
  address: string
  totalPackets: number
  effectedPackets: number
}

const props = defineProps<{
  relayers?: Relayer[]
}>()

const chartCanvas = ref<HTMLCanvasElement>()
let chart: Chart | null = null

const colors = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
  '#EC4899', '#6366F1', '#14B8A6', '#F97316', '#84CC16'
]

const marketData = computed(() => {
  if (!props.relayers || props.relayers.length === 0) return []
  
  const totalPackets = props.relayers.reduce((sum, r) => sum + r.totalPackets, 0)
  if (totalPackets === 0) return []
  
  return props.relayers
    .map(r => ({
      address: r.address,
      packets: r.totalPackets,
      percentage: (r.totalPackets / totalPackets) * 100
    }))
    .sort((a, b) => b.packets - a.packets)
})

const topRelayers = computed(() => {
  return marketData.value.slice(0, 8)
})

const otherRelayers = computed(() => {
  const others = marketData.value.slice(8)
  if (others.length === 0) return null
  
  return {
    address: 'Others',
    packets: others.reduce((sum, r) => sum + r.packets, 0),
    percentage: others.reduce((sum, r) => sum + r.percentage, 0)
  }
})

function formatAddress(address: string): string {
  if (address === 'Others') return address
  return address.slice(0, 8) + '...' + address.slice(-4)
}

function createChart() {
  if (!chartCanvas.value || marketData.value.length === 0) return
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  const data = [...topRelayers.value]
  if (otherRelayers.value) {
    data.push(otherRelayers.value)
  }
  
  chart = new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: data.map(d => formatAddress(d.address)),
      datasets: [{
        data: data.map(d => d.packets),
        backgroundColor: data.map((_, i) => colors[i % colors.length]),
        borderWidth: 0
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
            label: function(context) {
              const percentage = ((context.parsed / context.dataset.data.reduce((a, b) => a + b, 0)) * 100).toFixed(1)
              return `${context.label}: ${context.parsed.toLocaleString()} packets (${percentage}%)`
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

watch(() => props.relayers, () => {
  if (chart) {
    chart.destroy()
  }
  createChart()
})
</script>