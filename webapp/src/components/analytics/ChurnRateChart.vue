<template>
  <div class="relative">
    <div class="h-[250px]">
      <canvas ref="chartCanvas"></canvas>
    </div>
    
    <!-- Churn Summary -->
    <div class="mt-4 grid grid-cols-3 gap-3 text-center">
      <div class="bg-gray-50 rounded-lg p-3">
        <p class="text-xs text-gray-500">Avg Weekly Churn</p>
        <p class="text-lg font-semibold text-gray-900">{{ avgChurnRate }}%</p>
      </div>
      <div class="bg-green-50 rounded-lg p-3">
        <p class="text-xs text-gray-500">Net Growth</p>
        <p class="text-lg font-semibold text-green-600">+{{ netGrowth }}</p>
      </div>
      <div class="bg-blue-50 rounded-lg p-3">
        <p class="text-xs text-gray-500">Retention Rate</p>
        <p class="text-lg font-semibold text-blue-600">{{ retentionRate }}%</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Props {
  data?: {
    newEntrants: number
    exits: number
    data: Array<{
      week: number
      entries: number
      exits: number
    }>
  }
}

const props = defineProps<Props>()
const chartCanvas = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// Calculate metrics
const avgChurnRate = computed(() => {
  const data = props.data?.data || []
  if (data.length === 0) return 0
  
  const avgExits = data.reduce((sum, d) => sum + d.exits, 0) / data.length
  const avgTotal = 50 // Assume average 50 active relayers
  return Math.round((avgExits / avgTotal) * 100)
})

const netGrowth = computed(() => {
  const data = props.data?.data || []
  if (data.length === 0) return 0
  
  const totalEntries = data.reduce((sum, d) => sum + d.entries, 0)
  const totalExits = data.reduce((sum, d) => sum + d.exits, 0)
  return totalEntries - totalExits
})

const retentionRate = computed(() => {
  return 100 - avgChurnRate.value
})

function createChart() {
  if (!chartCanvas.value) return
  
  const chartData = props.data?.data || generateDefaultData()
  const labels = generateLabels(chartData.length)
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  // Calculate net change for each week
  const netChange = chartData.map(d => d.entries - d.exits)
  
  chart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels,
      datasets: [
        {
          label: 'New Entrants',
          data: chartData.map(d => d.entries),
          backgroundColor: 'rgba(34, 197, 94, 0.8)',
          borderColor: 'rgb(34, 197, 94)',
          borderWidth: 1,
          borderRadius: 4,
          stack: 'entries'
        },
        {
          label: 'Exits',
          data: chartData.map(d => -d.exits),
          backgroundColor: 'rgba(239, 68, 68, 0.8)',
          borderColor: 'rgb(239, 68, 68)',
          borderWidth: 1,
          borderRadius: 4,
          stack: 'exits'
        },
        {
          label: 'Net Change',
          data: netChange,
          type: 'line',
          borderColor: 'rgb(59, 130, 246)',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          borderWidth: 3,
          tension: 0.4,
          pointRadius: 4,
          pointHoverRadius: 6,
          pointBackgroundColor: 'rgb(59, 130, 246)',
          pointBorderColor: '#fff',
          pointBorderWidth: 2,
          yAxisID: 'y1'
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: true,
          position: 'top',
          labels: {
            font: {
              size: 11
            },
            boxWidth: 20,
            padding: 10
          }
        },
        tooltip: {
          mode: 'index',
          intersect: false,
          callbacks: {
            label: function(context) {
              const label = context.dataset.label || ''
              const value = Math.abs(context.parsed.y)
              
              if (label === 'Exits') {
                return `${label}: ${value}`
              }
              
              return `${label}: ${value}`
            }
          }
        }
      },
      scales: {
        x: {
          grid: {
            display: false
          },
          ticks: {
            font: {
              size: 11
            }
          }
        },
        y: {
          stacked: true,
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 11
            },
            callback: function(value) {
              return Math.abs(value as number)
            }
          }
        },
        y1: {
          type: 'linear',
          display: true,
          position: 'right',
          grid: {
            drawOnChartArea: false
          },
          ticks: {
            font: {
              size: 11
            },
            callback: function(value) {
              const val = value as number
              return val > 0 ? `+${val}` : val
            }
          }
        }
      },
      interaction: {
        mode: 'index',
        intersect: false
      }
    }
  })
}

function generateDefaultData() {
  // Generate 12 weeks of data
  return Array.from({ length: 12 }, (_, i) => {
    // Simulate growth trend with some variation
    const baseEntries = 3 + Math.floor(i / 4) // Gradual increase
    const baseExits = 2
    
    return {
      week: i + 1,
      entries: Math.max(1, baseEntries + Math.floor(Math.random() * 3)),
      exits: Math.max(0, baseExits + Math.floor(Math.random() * 2))
    }
  })
}

function generateLabels(count: number) {
  const labels = []
  const today = new Date()
  
  for (let i = count - 1; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - (i * 7))
    labels.push(`Week ${count - i}`)
  }
  
  return labels
}

onMounted(() => {
  createChart()
})

watch(() => props.data, () => {
  if (chart) {
    chart.destroy()
  }
  createChart()
}, { deep: true })
</script>