<template>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-sm font-medium text-gray-900">{{ title }}</h4>
      <span class="text-xs text-gray-500">Prediction based on historical trends</span>
    </div>
    
    <div class="h-48">
      <canvas ref="chartCanvas"></canvas>
    </div>
    
    <div class="mt-3 grid grid-cols-2 gap-4 text-xs">
      <div>
        <p class="text-gray-500">Current</p>
        <p class="font-medium text-gray-900">{{ formatValue(currentValue) }}</p>
      </div>
      <div>
        <p class="text-gray-500">Predicted</p>
        <p class="font-medium" :class="getTrendClass()">{{ formatValue(predictedValue) }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Props {
  title: string
  data?: any[]
  type: 'volume' | 'rate'
}

const props = defineProps<Props>()
const chartCanvas = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// Calculate current and predicted values
const currentValue = computed(() => {
  if (!props.data || props.data.length === 0) return 0
  const recent = props.data.slice(-24) // Last 24 data points
  return recent.reduce((sum, d) => sum + d.value, 0) / recent.length
})

const predictedValue = computed(() => {
  if (!props.data || props.data.length === 0) return 0
  
  // Simple linear prediction based on trend
  const recent = props.data.slice(-48) // Last 48 data points
  const firstHalf = recent.slice(0, 24)
  const secondHalf = recent.slice(24)
  
  const firstAvg = firstHalf.reduce((sum, d) => sum + d.value, 0) / firstHalf.length
  const secondAvg = secondHalf.reduce((sum, d) => sum + d.value, 0) / secondHalf.length
  
  const trend = secondAvg - firstAvg
  return Math.max(0, secondAvg + trend) // Project the trend forward
})

function formatValue(value: number): string {
  if (props.type === 'rate') {
    return `${value.toFixed(1)}%`
  }
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toFixed(0)
}

function getTrendClass(): string {
  const diff = predictedValue.value - currentValue.value
  if (Math.abs(diff) < currentValue.value * 0.05) return 'text-gray-900' // Less than 5% change
  if (props.type === 'rate') {
    return diff > 0 ? 'text-green-600' : 'text-red-600'
  }
  return diff > 0 ? 'text-green-600' : 'text-yellow-600'
}

function createChart() {
  if (!chartCanvas.value || !props.data || props.data.length === 0) return
  
  // Generate prediction data
  const historicalData = props.data
  const futureData = generatePrediction(historicalData)
  
  const labels = [
    ...historicalData.map(d => d.time),
    ...futureData.map(d => d.time)
  ]
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: 'Historical',
          data: [...historicalData.map(d => d.value), ...Array(futureData.length).fill(null)],
          borderColor: 'rgb(59, 130, 246)',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          tension: 0.4,
          spanGaps: false
        },
        {
          label: 'Predicted',
          data: [...Array(historicalData.length - 1).fill(null), historicalData[historicalData.length - 1].value, ...futureData.map(d => d.value)],
          borderColor: 'rgb(168, 85, 247)',
          backgroundColor: 'rgba(168, 85, 247, 0.1)',
          borderDash: [5, 5],
          tension: 0.4,
          spanGaps: false
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
          mode: 'index',
          intersect: false
        }
      },
      scales: {
        x: {
          display: false
        },
        y: {
          display: true,
          grid: {
            display: true,
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 10
            },
            callback: function(value) {
              if (props.type === 'rate') return value + '%'
              const val = value as number
              if (val >= 1000) return (val / 1000) + 'K'
              return val
            }
          }
        }
      }
    }
  })
}


function generatePrediction(historicalData: any[]) {
  const predictions = []
  const now = Date.now()
  
  // Simple linear regression for prediction
  const recent = historicalData.slice(-24)
  const avgGrowth = recent.reduce((sum, d, i) => {
    if (i === 0) return sum
    return sum + (d.value - recent[i - 1].value)
  }, 0) / (recent.length - 1)
  
  let lastValue = historicalData[historicalData.length - 1].value
  
  for (let i = 1; i <= 24; i++) {
    const time = new Date(now + i * 3600000).toISOString()
    // Use predictable variation based on time pattern
    const timeVariation = Math.sin(i * Math.PI / 12) * 0.05 * lastValue
    const value = lastValue + avgGrowth + timeVariation
    
    predictions.push({
      time,
      value: Math.max(0, value)
    })
    
    lastValue = value
  }
  
  return predictions
}

onMounted(() => {
  createChart()
})

// Update chart when data changes
watch(() => props.data, (newData) => {
  if (!chart || !newData) return
  
  // Update existing chart data instead of recreating
  const historicalData = newData
  const futureData = generatePrediction(historicalData)
  
  // Update chart data without destroying
  chart.data.labels = [
    ...historicalData.map(d => d.time),
    ...futureData.map(d => d.time)
  ]
  
  chart.data.datasets[0].data = [
    ...historicalData.map(d => d.value), 
    ...Array(futureData.length).fill(null)
  ]
  
  chart.data.datasets[1].data = [
    ...Array(historicalData.length - 1).fill(null), 
    historicalData[historicalData.length - 1].value, 
    ...futureData.map(d => d.value)
  ]
  
  // Update with animation disabled for smooth transition
  chart.update('none')
})
</script>