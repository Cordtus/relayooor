<template>
  <div class="relative">
    <div class="bg-gray-50 rounded-lg p-8 min-h-[300px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
        </svg>
        <h3 class="text-lg font-medium text-gray-900 mb-2">Success Rate Trend Chart</h3>
        <p class="text-sm text-gray-600">Historical success rate analysis coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will display trends and projections</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  data?: {
    current: number
    projected: number
    data: Array<{
      day: number
      rate: number
    }>
  }
}

const props = defineProps<Props>()
</script>