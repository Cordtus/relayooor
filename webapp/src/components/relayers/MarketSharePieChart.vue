<template>
  <div class="bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Market Share Distribution</h3>
    <div class="bg-gray-50 rounded-lg p-8 min-h-[300px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"></path>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z"></path>
        </svg>
        <h4 class="text-lg font-medium text-gray-900 mb-2">Market Share Pie Chart</h4>
        <p class="text-sm text-gray-600">Interactive market share visualization coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will show relayer market share distribution</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  data?: Array<{
    relayer: string
    share: number
    packets: number
  }>
}

const props = defineProps<Props>()
</script>