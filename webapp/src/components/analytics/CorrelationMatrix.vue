<template>
  <div class="bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Correlation Matrix</h3>
    <div class="bg-gray-50 rounded-lg p-8 min-h-[400px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
        </svg>
        <h4 class="text-lg font-medium text-gray-900 mb-2">Correlation Matrix</h4>
        <p class="text-sm text-gray-600">Metric correlation analysis coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will show relationships between different metrics</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  metrics?: string[]
  correlations?: number[][]
}

const props = defineProps<Props>()
</script>