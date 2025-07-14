<template>
  <span
    :class="[
      'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
      statusClasses[status] || statusClasses.unknown
    ]"
  >
    <span
      :class="[
        'w-1.5 h-1.5 rounded-full mr-1.5',
        dotClasses[status] || dotClasses.unknown
      ]"
    />
    {{ statusLabels[status] || status }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  status: string
}>()

const statusClasses: Record<string, string> = {
  stuck: 'bg-orange-100 text-orange-800',
  pending: 'bg-yellow-100 text-yellow-800',
  processing: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  unknown: 'bg-gray-100 text-gray-800'
}

const dotClasses: Record<string, string> = {
  stuck: 'bg-orange-400',
  pending: 'bg-yellow-400',
  processing: 'bg-blue-400',
  completed: 'bg-green-400',
  failed: 'bg-red-400',
  unknown: 'bg-gray-400'
}

const statusLabels: Record<string, string> = {
  stuck: 'Stuck',
  pending: 'Pending',
  processing: 'Processing',
  completed: 'Completed',
  failed: 'Failed'
}
</script>