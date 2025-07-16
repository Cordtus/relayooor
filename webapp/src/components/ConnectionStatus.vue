<template>
  <div class="flex items-center space-x-2">
    <div :class="statusClasses" />
    <span class="text-sm text-content-secondary">
      {{ statusText }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConnectionStore } from '@/stores/connection'

const connectionStore = useConnectionStore()

const statusClasses = computed(() => {
  const base = 'w-2 h-2 rounded-full'
  if (connectionStore.connectionStatus === 'connected') {
    return `${base} bg-status-success`
  } else if (connectionStore.connectionStatus === 'connecting') {
    return `${base} bg-status-warning animate-pulse`
  } else {
    return `${base} bg-status-error`
  }
})

const statusText = computed(() => {
  switch (connectionStore.connectionStatus) {
    case 'connected':
      return 'Connected'
    case 'connecting':
      return 'Connecting...'
    default:
      return 'Disconnected'
  }
})
</script>