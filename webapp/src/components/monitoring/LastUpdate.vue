<template>
  <div class="text-sm text-gray-500">
    Last update: {{ formattedTime }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { format } from 'date-fns'

const props = defineProps<{
  timestamp: Date
}>()

const formattedTime = computed(() => {
  return format(props.timestamp, 'HH:mm:ss')
})
</script>