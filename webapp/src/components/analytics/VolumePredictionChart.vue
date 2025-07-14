<template>
  <div class="relative">
    <div class="h-[300px]">
      <canvas ref="chartCanvas"></canvas>
    </div>
    <div class="mt-4 flex items-center justify-between text-sm">
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <div class="w-3 h-3 bg-blue-500 rounded"></div>
          <span class="text-gray-600">Predicted Volume</span>
        </div>
        <div class="flex items-center gap-2">
          <div class="w-3 h-3 bg-blue-200 rounded"></div>
          <span class="text-gray-600">Confidence Interval</span>
        </div>
      </div>
      <span class="text-xs text-gray-500">Based on historical patterns</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Props {
  data?: {
    expectedTotal: number
    confidence: number
    data: Array<{
      day: number
      predicted: number
      confidence: number
    }>
  }
}

const props = defineProps<Props>()
const chartCanvas = ref<HTMLCanvasElement>()
let chart: Chart | null = null

function createChart() {
  if (!chartCanvas.value) return
  
  const chartData = props.data?.data || generateDefaultData()
  const dates = generateDates()
  
  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return
  
  // Calculate confidence intervals
  const upperBound = chartData.map(d => d.predicted * (1 + (100 - d.confidence) / 200))
  const lowerBound = chartData.map(d => d.predicted * (1 - (100 - d.confidence) / 200))
  
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: dates,
      datasets: [
        {
          label: 'Predicted Volume',
          data: chartData.map(d => d.predicted),
          borderColor: 'rgb(59, 130, 246)',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          borderWidth: 3,
          tension: 0.4,
          pointRadius: 4,
          pointHoverRadius: 6,
          pointBackgroundColor: 'rgb(59, 130, 246)',
          pointBorderColor: '#fff',
          pointBorderWidth: 2
        },
        {
          label: 'Upper Bound',
          data: upperBound,
          borderColor: 'rgba(59, 130, 246, 0.3)',
          backgroundColor: 'rgba(59, 130, 246, 0.05)',
          borderWidth: 1,
          borderDash: [5, 5],
          pointRadius: 0,
          fill: '+1'
        },
        {
          label: 'Lower Bound',
          data: lowerBound,
          borderColor: 'rgba(59, 130, 246, 0.3)',
          backgroundColor: 'rgba(59, 130, 246, 0.05)',
          borderWidth: 1,
          borderDash: [5, 5],
          pointRadius: 0,
          fill: false
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
          intersect: false,
          callbacks: {
            label: function(context) {
              const label = context.dataset.label || ''
              const value = context.parsed.y
              
              if (label === 'Predicted Volume') {
                const confidence = chartData[context.dataIndex].confidence
                return [
                  `Volume: ${formatNumber(value)}`,
                  `Confidence: ${confidence}%`
                ]
              }
              
              return `${label}: ${formatNumber(value)}`
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
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 11
            },
            callback: function(value) {
              return formatNumber(value as number)
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
  const baseVolume = 350000
  const trend = 1.02 // 2% daily growth
  
  return Array.from({ length: 7 }, (_, i) => {
    const dayMultiplier = Math.pow(trend, i)
    const randomVariation = 0.9 + Math.random() * 0.2
    const predicted = Math.round(baseVolume * dayMultiplier * randomVariation)
    const confidence = Math.max(70, 95 - (i * 3)) // Confidence decreases with time
    
    return { day: i + 1, predicted, confidence }
  })
}

function generateDates() {
  const dates = []
  const today = new Date()
  
  for (let i = 0; i < 7; i++) {
    const date = new Date(today)
    date.setDate(date.getDate() + i + 1)
    dates.push(date.toLocaleDateString('en', { month: 'short', day: 'numeric' }))
  }
  
  return dates
}

function formatNumber(value: number): string {
  if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'
  if (value >= 1000) return (value / 1000).toFixed(0) + 'K'
  return value.toString()
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