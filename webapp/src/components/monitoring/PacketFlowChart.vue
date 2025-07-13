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
}

const props = withDefaults(defineProps<Props>(), {
  height: 300
})

const chartCanvas = ref<HTMLCanvasElement>()
let chartInstance: Chart | null = null

const createChart = () => {
  if (!chartCanvas.value || !props.data) return

  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return

  // Destroy existing chart
  if (chartInstance) {
    chartInstance.destroy()
  }

  // Generate time series data
  const labels = Array.from({ length: 24 }, (_, i) => `${i}:00`)
  const effectedData = Array.from({ length: 24 }, () => Math.floor(Math.random() * 1000) + 500)
  const uneffectedData = Array.from({ length: 24 }, () => Math.floor(Math.random() * 200) + 50)

  chartInstance = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: 'Effected Packets',
          data: effectedData,
          borderColor: 'rgb(34, 197, 94)',
          backgroundColor: 'rgba(34, 197, 94, 0.1)',
          fill: true,
          tension: 0.4
        },
        {
          label: 'Uneffected Packets',
          data: uneffectedData,
          borderColor: 'rgb(239, 68, 68)',
          backgroundColor: 'rgba(239, 68, 68, 0.1)',
          fill: true,
          tension: 0.4
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
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
            text: 'Hour (UTC)'
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
}

onMounted(() => {
  createChart()
})

watch(() => props.data, () => {
  createChart()
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.destroy()
  }
})
</script>