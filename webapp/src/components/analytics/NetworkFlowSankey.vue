<template>
  <div class="bg-gray-50 rounded-lg p-8 min-h-[400px] flex items-center justify-center">
    <div class="text-center">
      <svg class="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
      </svg>
      <h3 class="text-lg font-medium text-gray-900 mb-2">Network Flow Sankey Diagram</h3>
      <p class="text-sm text-gray-600">Interactive flow visualization coming soon</p>
      <p class="text-xs text-gray-500 mt-2">This will display packet flows between chains</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  flows?: {
    nodes: Array<{ id: string; name: string }>
    links: Array<{ source: string; target: string; value: number }>
  }
}

const props = defineProps<Props>()
</script>