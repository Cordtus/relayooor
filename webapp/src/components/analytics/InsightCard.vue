<template>
  <div class="bg-white rounded-lg shadow p-6">
    <div class="flex items-start justify-between">
      <div>
        <h3 class="text-sm font-medium text-gray-500">{{ title }}</h3>
        <p class="mt-1 text-2xl font-semibold text-gray-900">{{ value }}</p>
        <p class="mt-1 text-sm text-gray-600">{{ description }}</p>
      </div>
      <div v-if="trend" class="flex items-center">
        <span :class="[
          'text-sm font-medium',
          trend.startsWith('+') ? 'text-green-600' : 'text-red-600'
        ]">
          {{ trend }}
        </span>
      </div>
    </div>
    <div class="mt-4 text-center text-gray-500">
      <p class="text-sm">InsightCard component coming soon</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  value: string | number
  trend?: string
  description?: string
  icon?: string
  color?: string
}

const props = defineProps<Props>()
</script>