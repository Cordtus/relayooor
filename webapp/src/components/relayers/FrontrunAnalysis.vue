<template>
  <div class="bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Frontrun Analysis</h3>
    <div class="bg-gray-50 rounded-lg p-8 min-h-[400px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
        </svg>
        <h4 class="text-lg font-medium text-gray-900 mb-2">Frontrun Detection & Analysis</h4>
        <p class="text-sm text-gray-600">Advanced frontrun analytics coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will analyze frontrun patterns and provide insights</p>
      </div>
    </div>
    
    <div class="mt-6 grid grid-cols-3 gap-4">
      <div class="text-center">
        <p class="text-2xl font-semibold text-gray-900">--</p>
        <p class="text-sm text-gray-500">Total Frontruns</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-semibold text-gray-900">--</p>
        <p class="text-sm text-gray-500">Affected Relayers</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-semibold text-gray-900">--</p>
        <p class="text-sm text-gray-500">Avg. Gas Difference</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  data?: {
    frontruns: Array<{
      timestamp: Date
      victim: string
      frontrunner: string
      gasDifference: number
    }>
    summary: {
      totalFrontruns: number
      affectedRelayers: number
      avgGasDifference: number
    }
  }
}

const props = defineProps<Props>()
</script>