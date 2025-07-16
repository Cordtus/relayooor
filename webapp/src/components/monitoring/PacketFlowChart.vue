<template>
  <div class="relative" :style="{ height: height + 'px' }">
    <canvas ref="chartCanvas"></canvas>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

interface Props {
  data?: any
  height?: number
  maxDataPoints?: number
  updateInterval?: number
}

const props = withDefaults(defineProps<Props>(), {
  height: 300,
  maxDataPoints: 60, // 60 minutes of data
  updateInterval: 5000 // Update every 5 seconds
})

const chartCanvas = ref<HTMLCanvasElement>()
let chartInstance: Chart | null = null
let updateTimer: ReturnType<typeof setInterval> | null = null

// Store historical data
const historicalData = ref<{
  labels: string[]
  effected: number[]
  uneffected: number[]
}>({
  labels: [],
  effected: [],
  uneffected: []
})

const initializeData = () => {
  // Initialize with some historical data
  const now = new Date()
  for (let i = props.maxDataPoints - 1; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000) // 1 minute intervals
    historicalData.value.labels.push(time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }))
    // Use time-based pattern instead of random
    const minutes = time.getMinutes()
    const hourPattern = 1 + Math.sin(minutes * Math.PI / 30) * 0.3 // Sine wave pattern
    historicalData.value.effected.push(Math.floor(800 * hourPattern))
    historicalData.value.uneffected.push(Math.floor(150 * hourPattern))
  }
}

const addDataPoint = () => {
  const now = new Date()
  const label = now.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
  
  // Add new data point
  historicalData.value.labels.push(label)
  // Calculate based on current minute for consistent pattern
  const minutes = now.getMinutes()
  const seconds = now.getSeconds()
  const timePattern = 1 + Math.sin((minutes + seconds/60) * Math.PI / 30) * 0.3
  
  // Base rates from props.data if available, otherwise use defaults
  const baseEffected = props.data?.baseEffected || 800
  const baseUneffected = props.data?.baseUneffected || 150
  
  historicalData.value.effected.push(Math.floor(baseEffected * timePattern))
  historicalData.value.uneffected.push(Math.floor(baseUneffected * timePattern))
  
  // Remove oldest data point if we exceed max
  if (historicalData.value.labels.length > props.maxDataPoints) {
    historicalData.value.labels.shift()
    historicalData.value.effected.shift()
    historicalData.value.uneffected.shift()
  }
  
  // Update chart without recreating it
  if (chartInstance) {
    chartInstance.data.labels = [...historicalData.value.labels]
    chartInstance.data.datasets[0].data = [...historicalData.value.effected]
    chartInstance.data.datasets[1].data = [...historicalData.value.uneffected]
    chartInstance.update('none') // Use 'none' animation mode for smooth updates
  }
}

const createChart = () => {
  if (!chartCanvas.value) return

  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return

  // Initialize data if empty
  if (historicalData.value.labels.length === 0) {
    initializeData()
  }

  chartInstance = new Chart(ctx, {
    type: 'line',
    data: {
      labels: [...historicalData.value.labels],
      datasets: [
        {
          label: 'Effected Packets',
          data: [...historicalData.value.effected],
          borderColor: 'rgb(34, 197, 94)',
          backgroundColor: 'rgba(34, 197, 94, 0.1)',
          fill: true,
          tension: 0.4,
          pointRadius: 2,
          pointHoverRadius: 4
        },
        {
          label: 'Uneffected Packets',
          data: [...historicalData.value.uneffected],
          borderColor: 'rgb(239, 68, 68)',
          backgroundColor: 'rgba(239, 68, 68, 0.1)',
          fill: true,
          tension: 0.4,
          pointRadius: 2,
          pointHoverRadius: 4
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 0 // Disable animations for smoother updates
      },
      plugins: {
        legend: {
          position: 'top'
        },
        tooltip: {
          mode: 'index',
          intersect: false
        }
      },
      scales: {
        x: {
          display: true,
          title: {
            display: true,
            text: 'Time (UTC)'
          },
          ticks: {
            maxRotation: 45,
            minRotation: 45,
            autoSkip: true,
            maxTicksLimit: 10
          }
        },
        y: {
          display: true,
          title: {
            display: true,
            text: 'Packets'
          },
          beginAtZero: true
        }
      }
    }
  })
  
  // Start the update timer
  startUpdateTimer()
}

const startUpdateTimer = () => {
  if (updateTimer) {
    clearInterval(updateTimer)
  }
  
  updateTimer = setInterval(() => {
    addDataPoint()
  }, props.updateInterval)
}

const stopUpdateTimer = () => {
  if (updateTimer) {
    clearInterval(updateTimer)
    updateTimer = null
  }
}

onMounted(() => {
  createChart()
})

watch(() => props.data, () => {
  // When new data comes in, integrate it into the chart
  // without recreating the entire chart
  if (props.data && chartInstance) {
    // Parse and merge new data with existing data
    // This is where you'd integrate real data from props.data
    addDataPoint()
  }
})

onUnmounted(() => {
  stopUpdateTimer()
  if (chartInstance) {
    chartInstance.destroy()
  }
})
</script>