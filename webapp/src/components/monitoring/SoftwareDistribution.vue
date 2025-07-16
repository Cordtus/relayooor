<template>
  <div class="bg-white rounded-lg shadow p-4">
    <h3 class="text-lg font-medium mb-4">Software Distribution</h3>
    <div v-if="distribution.length > 0">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- Software Types Chart -->
        <div>
          <h4 class="text-sm font-medium text-gray-600 mb-2">By Software Type</h4>
          <div class="h-48">
            <canvas ref="softwareChart"></canvas>
          </div>
        </div>
        
        <!-- Version Distribution -->
        <div>
          <h4 class="text-sm font-medium text-gray-600 mb-2">Version Details</h4>
          <div class="space-y-2 max-h-48 overflow-y-auto">
            <div v-for="sw in distribution" :key="sw.name" class="border-b pb-2">
              <div class="flex justify-between items-center">
                <span class="font-medium">{{ sw.name }}</span>
                <span class="text-sm text-gray-500">{{ sw.count }} relayers</span>
              </div>
              <div class="text-xs text-gray-400">
                <span v-for="(version, idx) in sw.versions.slice(0, 3)" :key="version">
                  {{ version }}<span v-if="idx < 2">, </span>
                </span>
                <span v-if="sw.versions.length > 3">...</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Summary Stats -->
      <div class="mt-4 pt-4 border-t grid grid-cols-3 gap-4 text-center">
        <div>
          <p class="text-2xl font-bold text-blue-600">{{ uniqueSoftwareCount }}</p>
          <p class="text-xs text-gray-500">Software Types</p>
        </div>
        <div>
          <p class="text-2xl font-bold text-green-600">{{ totalVersions }}</p>
          <p class="text-xs text-gray-500">Unique Versions</p>
        </div>
        <div>
          <p class="text-2xl font-bold text-purple-600">{{ dominantSoftware }}</p>
          <p class="text-xs text-gray-500">Most Popular</p>
        </div>
      </div>
    </div>
    <div v-else class="text-center py-8 text-gray-500">
      <p>Software distribution data not available</p>
      <p class="text-sm mt-2">Relayers have not reported their software versions</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Relayer {
  address: string
  software?: string
  version?: string
  memo?: string
}

const props = defineProps<{
  relayers?: Relayer[]
}>()

const softwareChart = ref<HTMLCanvasElement>()
let chart: Chart | null = null

// Group relayers by software and version
const distribution = computed(() => {
  if (!props.relayers || props.relayers.length === 0) return []
  
  const softwareMap = new Map<string, { count: number, versions: Set<string> }>()
  
  props.relayers.forEach(relayer => {
    // Try to extract software from memo if not explicitly set
    const software = relayer.software || extractSoftwareFromMemo(relayer.memo || '')
    const version = relayer.version || 'unknown'
    
    if (!softwareMap.has(software)) {
      softwareMap.set(software, { count: 0, versions: new Set() })
    }
    
    const entry = softwareMap.get(software)!
    entry.count++
    entry.versions.add(version)
  })
  
  // Convert to array and sort by count
  return Array.from(softwareMap.entries())
    .map(([name, data]) => ({
      name,
      count: data.count,
      versions: Array.from(data.versions).sort()
    }))
    .sort((a, b) => b.count - a.count)
})

const uniqueSoftwareCount = computed(() => distribution.value.length)

const totalVersions = computed(() => {
  return distribution.value.reduce((sum, sw) => sum + sw.versions.length, 0)
})

const dominantSoftware = computed(() => {
  if (distribution.value.length === 0) return 'N/A'
  return distribution.value[0].name
})

// Extract software from memo patterns
function extractSoftwareFromMemo(memo: string): string {
  const memoLower = memo.toLowerCase()
  
  // Common patterns
  if (memoLower.includes('hermes')) return 'Hermes'
  if (memoLower.includes('rly') || memoLower.includes('relayer')) return 'Go Relayer'
  if (memoLower.includes('ts-relayer')) return 'TS Relayer'
  if (memoLower.includes('neutron')) return 'Neutron'
  if (memoLower.includes('ibc-go')) return 'IBC-Go'
  
  // If memo looks like a known relayer operator
  if (memoLower.includes('cephalopod')) return 'Hermes'
  if (memoLower.includes('strangelove')) return 'Go Relayer'
  if (memoLower.includes('notional')) return 'Hermes'
  if (memoLower.includes('polkachu')) return 'Hermes'
  
  return 'Unknown'
}

function createChart() {
  if (!softwareChart.value || distribution.value.length === 0) return
  
  const ctx = softwareChart.value.getContext('2d')
  if (!ctx) return
  
  const colors = [
    '#3B82F6', // blue
    '#10B981', // green
    '#F59E0B', // amber
    '#EF4444', // red
    '#8B5CF6', // violet
    '#EC4899', // pink
    '#6366F1', // indigo
  ]
  
  chart = new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: distribution.value.map(d => d.name),
      datasets: [{
        data: distribution.value.map(d => d.count),
        backgroundColor: distribution.value.map((_, i) => colors[i % colors.length]),
        borderWidth: 0
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'bottom',
          labels: {
            padding: 10,
            font: { size: 11 }
          }
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              const total = context.dataset.data.reduce((a, b) => a + b, 0)
              const percentage = ((context.parsed / total) * 100).toFixed(1)
              return `${context.label}: ${context.parsed} (${percentage}%)`
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