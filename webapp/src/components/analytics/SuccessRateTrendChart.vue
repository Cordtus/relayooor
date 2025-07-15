<template>
  <div class="relative">
    <div class="h-[300px]">
      <canvas ref="chartCanvas"></canvas>
    </div>
    <div class="mt-4 grid grid-cols-3 gap-4 text-sm">
      <div class="text-center">
        <p class="text-xs text-gray-500">30-Day Average</p>
        <p class="text-lg font-semibold text-gray-900">{{ averageRate.toFixed(1) }}%</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-gray-500">Trend</p>
        <p class="text-lg font-semibold" :class="trendClass">
          {{ trendDirection }} {{ Math.abs(trendValue).toFixed(1) }}%
        </p>
      </div>
      <div class="text-center">
        <p class="text-xs text-gray-500">Volatility</p>
        <p class="text-lg font-semibold text-gray-900">{{ volatility }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Props {
  data?: {
    current: number
    projected: number
    data: Array<{
      day: number
      rate: number
    }>
  }
}

const props = defineProps<Props>()
const chartCanvas = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// Computed metrics
const averageRate = computed(() => {
  const rates = props.data?.data || []
  if (rates.length === 0) return 87.5
  return rates.reduce((sum, d) => sum + d.rate, 0) / rates.length
})

const trendValue = computed(() => {
  const rates = props.data?.data || []
  if (rates.length < 2) return 0
  
  // Simple linear regression for trend
  const n = rates.length
  const sumX = rates.reduce((sum, d) => sum + d.day, 0)
  const sumY = rates.reduce((sum, d) => sum + d.rate, 0)
  const sumXY = rates.reduce((sum, d) => sum + (d.day * d.rate), 0)
  const sumX2 = rates.reduce((sum, d) => sum + (d.day * d.day), 0)
  
  const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX)
  return slope * 30 // Trend over 30 days
})

const trendDirection = computed(() => trendValue.value >= 0 ? '↑' : '↓')
const trendClass = computed(() => trendValue.value >= 0 ? 'text-green-600' : 'text-red-600')

const volatility = computed(() => {
  const rates = props.data?.data || []
  if (rates.length < 2) return 'Low'
  
  const avg = averageRate.value
  const variance = rates.reduce((sum, d) => sum + Math.pow(d.rate - avg, 2), 0) / rates.length
  const stdDev = Math.sqrt(variance)
  
  if (stdDev < 2) return 'Low'
  if (stdDev < 5) return 'Medium'
  return 'High'
})

function createChart() {
  if (!chartCanvas.value) return
  
  const chartData = props.data?.data || generateDefaultData()
  const labels = generateLabels(chartData.length)
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  // Generate moving average
  const movingAvg = calculateMovingAverage(chartData.map(d => d.rate), 7)
  
  // Generate projection
  const projectionStart = chartData.length
  const projectedData = generateProjection(chartData, 7)
  
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: [...labels, ...generateProjectionLabels(7)],
      datasets: [
        {
          label: 'Success Rate',
          data: [...chartData.map(d => d.rate), ...Array(7).fill(null)],
          borderColor: 'rgb(34, 197, 94)',
          backgroundColor: 'rgba(34, 197, 94, 0.1)',
          borderWidth: 2,
          tension: 0.1,
          pointRadius: 3,
          pointHoverRadius: 5
        },
        {
          label: '7-Day Moving Average',
          data: [...movingAvg, ...Array(7).fill(null)],
          borderColor: 'rgb(59, 130, 246)',
          borderWidth: 2,
          borderDash: [5, 5],
          tension: 0.4,
          pointRadius: 0
        },
        {
          label: 'Projected',
          data: [...Array(projectionStart - 1).fill(null), chartData[chartData.length - 1].rate, ...projectedData],
          borderColor: 'rgb(168, 85, 247)',
          backgroundColor: 'rgba(168, 85, 247, 0.05)',
          borderWidth: 2,
          borderDash: [8, 4],
          tension: 0.4,
          pointRadius: 3,
          pointStyle: 'rect'
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
            boxWidth: 30,
            padding: 10
          }
        },
        tooltip: {
          mode: 'index',
          intersect: false,
          callbacks: {
            label: function(context) {
              const label = context.dataset.label || ''
              const value = context.parsed.y
              return `${label}: ${value.toFixed(1)}%`
            }
          }
        },
        // Note: annotation plugin would need to be registered separately
      },
      scales: {
        x: {
          grid: {
            display: false
          },
          ticks: {
            font: {
              size: 11
            },
            maxRotation: 45,
            minRotation: 45
          }
        },
        y: {
          min: Math.max(0, Math.min(...chartData.map(d => d.rate)) - 5),
          max: Math.min(100, Math.max(...chartData.map(d => d.rate)) + 5),
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 11
            },
            callback: function(value) {
              return value + '%'
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

function generateDefaultData() {
  const baseRate = 85
  const days = 30
  
  return Array.from({ length: days }, (_, i) => {
    // Add some realistic variation
    const trend = i * 0.1 // Slight upward trend
    // Use day-based variation instead of random
    const dayHash = (i * 31) % 13 // Pseudo-random based on day
    const dailyVariation = ((dayHash / 13) - 0.5) * 4
    const weeklyPattern = Math.sin((i % 7) / 7 * Math.PI) * 2
    
    const rate = Math.max(70, Math.min(95, baseRate + trend + dailyVariation + weeklyPattern))
    
    return { day: i + 1, rate }
  })
}

function generateLabels(count: number) {
  const labels = []
  const today = new Date()
  
  for (let i = count - 1; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    labels.push(date.toLocaleDateString('en', { month: 'short', day: 'numeric' }))
  }
  
  return labels
}

function generateProjectionLabels(count: number) {
  const labels = []
  const today = new Date()
  
  for (let i = 1; i <= count; i++) {
    const date = new Date(today)
    date.setDate(date.getDate() + i)
    labels.push(date.toLocaleDateString('en', { month: 'short', day: 'numeric' }))
  }
  
  return labels
}

function calculateMovingAverage(data: number[], window: number) {
  const result = []
  
  for (let i = 0; i < data.length; i++) {
    if (i < window - 1) {
      result.push(null)
    } else {
      const sum = data.slice(i - window + 1, i + 1).reduce((a, b) => a + b, 0)
      result.push(sum / window)
    }
  }
  
  return result
}

function generateProjection(historicalData: Array<{day: number, rate: number}>, days: number) {
  const recent = historicalData.slice(-14) // Use last 2 weeks for projection
  const avgRate = recent.reduce((sum, d) => sum + d.rate, 0) / recent.length
  
  // Calculate trend
  const firstWeekAvg = recent.slice(0, 7).reduce((sum, d) => sum + d.rate, 0) / 7
  const secondWeekAvg = recent.slice(7).reduce((sum, d) => sum + d.rate, 0) / 7
  const weeklyTrend = secondWeekAvg - firstWeekAvg
  const dailyTrend = weeklyTrend / 7
  
  const projections = []
  let currentRate = historicalData[historicalData.length - 1].rate
  
  for (let i = 0; i < days; i++) {
    // Add trend with some dampening
    currentRate += dailyTrend * Math.pow(0.9, i)
    
    // Add some random variation that increases with time
    // Uncertainty increases with projection distance
    const uncertaintyBase = ((i * 37) % 11) / 11 // Deterministic pseudo-random
    const uncertainty = (uncertaintyBase - 0.5) * 2 * (i + 1) * 0.3
    
    // Constrain to reasonable bounds
    const projectedRate = Math.max(70, Math.min(95, currentRate + uncertainty))
    projections.push(projectedRate)
  }
  
  return projections
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