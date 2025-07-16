<template>
  <div class="flex items-center gap-2">
    <label class="text-sm text-gray-600">Refresh Rate:</label>
    <select
      :value="refreshInterval"
      @change="updateRefreshRate"
      class="text-sm border-gray-300 rounded-md shadow-sm focus:border-primary-500 focus:ring-primary-500"
    >
      <option v-for="interval in REFRESH_INTERVALS" :key="interval.value" :value="interval.value">
        {{ interval.label }}
      </option>
    </select>
    <span v-if="lastUpdate" class="text-xs text-gray-500">
      Last updated: {{ formatTimestamp(lastUpdate) }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSettingsStore, REFRESH_INTERVALS } from '@/stores/settings'

interface Props {
  lastUpdate?: Date
}

defineProps<Props>()

const settingsStore = useSettingsStore()
const refreshInterval = computed(() => settingsStore.settings.refreshInterval)

function updateRefreshRate(event: Event) {
  const value = parseInt((event.target as HTMLSelectElement).value)
  settingsStore.updateSettings({ refreshInterval: value })
}

function formatTimestamp(date: Date): string {
  const now = new Date()
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)
  
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  return `${Math.floor(seconds / 3600)}h ago`
}
</script>