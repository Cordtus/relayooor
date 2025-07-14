<template>
  <div class="bg-white rounded-lg shadow p-6">
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <h3 class="text-sm font-medium text-gray-500">{{ title }}</h3>
        <p class="mt-1 text-2xl font-semibold" :class="getValueClass()">{{ value }}</p>
        <p class="mt-1 text-sm text-gray-600">{{ description }}</p>
      </div>
      <div v-if="trend || level" class="flex items-center">
        <span v-if="trend" :class="[
          'text-sm font-medium',
          trend.startsWith('+') ? 'text-green-600' : 'text-red-600'
        ]">
          {{ trend }}
        </span>
        <span v-else-if="level" :class="[
          'px-2 py-1 text-xs font-medium rounded-full',
          level === 'low' ? 'bg-green-100 text-green-800' :
          level === 'medium' ? 'bg-yellow-100 text-yellow-800' :
          level === 'high' ? 'bg-red-100 text-red-800' :
          'bg-gray-100 text-gray-800'
        ]">
          {{ level }}
        </span>
      </div>
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
  level?: 'low' | 'medium' | 'high'
}

const props = defineProps<Props>()

function getValueClass(): string {
  if (props.color) {
    return `text-${props.color}-600`
  }
  if (props.level) {
    switch (props.level) {
      case 'high':
        return 'text-red-600'
      case 'medium':
        return 'text-yellow-600'
      case 'low':
        return 'text-green-600'
    }
  }
  return 'text-gray-900'
}
</script>