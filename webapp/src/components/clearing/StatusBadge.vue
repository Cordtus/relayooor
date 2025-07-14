<template>
  <Badge :variant="badgeVariant" dot>
    {{ statusLabels[status] || status }}
  </Badge>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Badge from '@/components/ui/Badge.vue'

const props = defineProps<{
  status: string
}>()

const badgeVariant = computed(() => {
  const variantMap: Record<string, 'success' | 'warning' | 'error' | 'info' | 'default'> = {
    stuck: 'warning',
    pending: 'warning',
    processing: 'info',
    completed: 'success',
    failed: 'error'
  }
  return variantMap[props.status] || 'default'
})

const statusLabels: Record<string, string> = {
  stuck: 'Stuck',
  pending: 'Pending',
  processing: 'Processing',
  completed: 'Completed',
  failed: 'Failed'
}
</script>