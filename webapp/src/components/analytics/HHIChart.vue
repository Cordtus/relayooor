<template>
  <div class="relative">
    <div class="h-[250px]">
      <canvas ref="chartCanvas"></canvas>
    </div>
    
    <!-- HHI Interpretation -->
    <div class="mt-4 p-3 rounded-lg" :class="interpretationClass">
      <div class="flex items-start gap-3">
        <component :is="interpretationIcon" class="h-5 w-5 mt-0.5" :class="interpretationIconClass" />
        <div class="flex-1">
          <p class="text-sm font-medium" :class="interpretationTextClass">
            {{ interpretation.title }}
          </p>
          <p class="text-xs mt-1" :class="interpretationSubtextClass">
            {{ interpretation.description }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import Chart from 'chart.js/auto'
import { AlertTriangle, CheckCircle, Info } from 'lucide-vue-next'

interface Props {
  data?: {
    current: number
    trend: number
    data: Array<{
      day: number
      value: number
    }>
  }
}

const props = defineProps<Props>()
const chartCanvas = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// HHI Interpretation based on value
const interpretation = computed(() => {
  const hhi = props.data?.current || 0
  
  if (hhi < 1500) {
    return {
      title: 'Competitive Market',
      description: 'The relayer market is highly competitive with no dominant players. This is healthy for the ecosystem.',
      level: 'good'
    }
  } else if (hhi < 2500) {
    return {
      title: 'Moderate Concentration',
      description: 'The market shows moderate concentration. Monitor for increasing dominance by top relayers.',
      level: 'warning'
    }
  } else {
    return {
      title: 'High Concentration',
      description: 'The market is highly concentrated. Consider measures to encourage more relayer participation.',
      level: 'danger'
    }
  }
})

const interpretationClass = computed(() => {
  switch (interpretation.value.level) {
    case 'good': return 'bg-green-50 border border-green-200'
    case 'warning': return 'bg-yellow-50 border border-yellow-200'
    case 'danger': return 'bg-red-50 border border-red-200'
    default: return 'bg-gray-50 border border-gray-200'
  }
})

const interpretationIcon = computed(() => {
  switch (interpretation.value.level) {
    case 'good': return CheckCircle
    case 'warning': return Info
    case 'danger': return AlertTriangle
    default: return Info
  }
})

const interpretationIconClass = computed(() => {
  switch (interpretation.value.level) {
    case 'good': return 'text-green-600'
    case 'warning': return 'text-yellow-600'
    case 'danger': return 'text-red-600'
    default: return 'text-gray-600'
  }
})

const interpretationTextClass = computed(() => {
  switch (interpretation.value.level) {
    case 'good': return 'text-green-900'
    case 'warning': return 'text-yellow-900'
    case 'danger': return 'text-red-900'
    default: return 'text-gray-900'
  }
})

const interpretationSubtextClass = computed(() => {
  switch (interpretation.value.level) {
    case 'good': return 'text-green-700'
    case 'warning': return 'text-yellow-700'
    case 'danger': return 'text-red-700'
    default: return 'text-gray-700'
  }
})

function createChart() {
  if (!chartCanvas.value) return
  
  const chartData = props.data?.data || generateDefaultData()
  const labels = generateLabels(chartData.length)
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: 'HHI Index',
          data: chartData.map(d => d.value),
          borderColor: getHHIColor,
          backgroundColor: (context) => {
            const value = context.parsed.y
            if (value < 1500) return 'rgba(34, 197, 94, 0.1)'
            if (value < 2500) return 'rgba(250, 204, 21, 0.1)'
            return 'rgba(239, 68, 68, 0.1)'
          },
          borderWidth: 2,
          tension: 0.3,
          pointRadius: 4,
          pointHoverRadius: 6,
          pointBackgroundColor: getHHIColor,
          pointBorderColor: '#fff',
          pointBorderWidth: 2,
          fill: true
        }
      ]
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
              const value = context.parsed.y
              const label = getMarketStatus(value)
              return [
                `HHI: ${value.toFixed(0)}`,
                `Status: ${label}`
              ]
            }
          }
        },
        // Note: annotation plugin would need to be registered separately
        // For now, we'll use visual cues in the chart itself
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
          min: 0,
          max: Math.max(4000, Math.max(...chartData.map(d => d.value)) * 1.2),
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 11
            },
            callback: function(value) {
              return value.toLocaleString()
            }
          }
        }
      },
      interaction: {
        mode: 'nearest',
        axis: 'x',
        intersect: false
      }
    }
  })
}

function getHHIColor(context: any): string {
  const value = context.parsed?.y || context.raw || 0
  if (value < 1500) return 'rgb(34, 197, 94)' // green
  if (value < 2500) return 'rgb(250, 204, 21)' // yellow
  return 'rgb(239, 68, 68)' // red
}

function getMarketStatus(hhi: number): string {
  if (hhi < 1500) return 'Competitive'
  if (hhi < 2500) return 'Moderate'
  return 'Concentrated'
}

function generateDefaultData() {
  const startHHI = 2300
  const trend = -20 // Decreasing concentration
  
  return Array.from({ length: 30 }, (_, i) => {
    const dayTrend = (trend / 30) * i
    const randomVariation = (Math.random() - 0.5) * 100
    const value = Math.max(1000, startHHI + dayTrend + randomVariation)
    
    return { day: i + 1, value }
  })
}

function generateLabels(count: number) {
  const labels = []
  const today = new Date()
  
  for (let i = count - 1; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    
    // Show fewer labels for readability
    if (i % 5 === 0 || i === count - 1) {
      labels.push(date.toLocaleDateString('en', { month: 'short', day: 'numeric' }))
    } else {
      labels.push('')
    }
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