<template>
  <div class="flex items-center gap-3">
    <label class="text-sm text-content-secondary">Refresh Rate:</label>
    <Dropdown 
      v-model="refreshInterval" 
      :options="[...REFRESH_INTERVALS]" 
      optionLabel="label" 
      optionValue="value"
      class="w-32"
      @change="updateRefreshRate"
    />
    <span v-if="lastUpdate" class="text-xs text-content-tertiary">
      Last updated: {{ formatTimestamp(lastUpdate) }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSettingsStore, REFRESH_INTERVALS } from '@/stores/settings'
import Dropdown from 'primevue/dropdown'

interface Props {
  lastUpdate?: Date
}

defineProps<Props>()

const settingsStore = useSettingsStore()
const refreshInterval = computed({
  get: () => settingsStore.settings.refreshInterval,
  set: (value) => settingsStore.updateSettings({ refreshInterval: value })
})

function updateRefreshRate(event: any) {
  settingsStore.updateSettings({ refreshInterval: event.value })
}

function formatTimestamp(date: Date): string {
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).format(date)
}
</script>