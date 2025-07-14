<template>
  <div class="bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Success Rate Heatmap</h3>
    <div class="bg-gray-50 rounded-lg p-8 min-h-[400px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"></path>
        </svg>
        <h4 class="text-lg font-medium text-gray-900 mb-2">Success Rate Heatmap</h4>
        <p class="text-sm text-gray-600">Time-based success rate visualization coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will show success rates across different time periods and channels</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  data?: Array<{
    hour: number
    day: string
    successRate: number
    volume: number
  }>
}

const props = defineProps<Props>()
</script>