<template>
  <div class="bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Technology Stack Distribution</h3>
    <div class="bg-gray-50 rounded-lg p-8 min-h-[300px] flex items-center justify-center">
      <div class="text-center">
        <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path>
        </svg>
        <h4 class="text-lg font-medium text-gray-900 mb-2">Tech Stack Donut Chart</h4>
        <p class="text-sm text-gray-600">Relayer technology distribution coming soon</p>
        <p class="text-xs text-gray-500 mt-2">Will show breakdown of relayer software and versions</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  data?: Array<{
    technology: string
    count: number
    percentage: number
  }>
}

const props = defineProps<Props>()
</script>