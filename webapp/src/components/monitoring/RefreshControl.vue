<template>
  <div class="flex items-center space-x-2">
    <label class="text-sm text-gray-600">Auto-refresh:</label>
    <input
      type="checkbox"
      :checked="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
      class="rounded text-primary-600 focus:ring-primary-500"
    />
    <button
      @click="$emit('refresh')"
      class="p-1 rounded hover:bg-gray-100"
      title="Refresh now"
    >
      <RefreshCw class="w-4 h-4 text-gray-600" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'

defineProps<{
  modelValue: boolean
  interval: number
}>()

defineEmits<{
  'update:modelValue': [value: boolean]
  'refresh': []
}>()
</script>