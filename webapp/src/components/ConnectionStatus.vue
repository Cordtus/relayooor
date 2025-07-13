<template>
  <div class="flex items-center space-x-2">
    <div class="flex items-center">
      <div class="h-2 w-2 rounded-full" :class="statusColor"></div>
      <span class="ml-2 text-sm text-gray-600">{{ statusText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { metricsService } from '@/services/api'

const { data: metrics } = useQuery({
  queryKey: ['connection-status'],
  queryFn: async () => {
    try {
      await metricsService.getRawMetrics()
      return { connected: true }
    } catch {
      return { connected: false }
    }
  },
  refetchInterval: 10000
})

const isConnected = computed(() => metrics.value?.connected ?? false)
const statusColor = computed(() => isConnected.value ? 'bg-green-400' : 'bg-red-400')
const statusText = computed(() => isConnected.value ? 'Connected' : 'Disconnected')
</script>