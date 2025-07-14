<template>
  <div class="relative">
    <div class="bg-gray-50 rounded-lg p-8 min-h-[300px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"></path>
        </svg>
        <h3 class="text-lg font-medium text-gray-900 mb-2">Volume Prediction Chart</h3>
        <p class="text-sm text-gray-600">ML-powered volume predictions coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will show 7-day forecast with confidence intervals</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
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
</script>