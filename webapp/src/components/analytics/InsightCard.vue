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
    <div v-if="props.icon" class="mt-4">
      <div class="h-24 flex items-center justify-center">
        <div class="w-full h-full bg-gradient-to-br from-gray-50 to-gray-100 rounded-lg flex items-center justify-center">
          <component 
            :is="getIconComponent()" 
            class="h-12 w-12"
            :class="getIconColor()"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Clock, Route, TrendingUp, AlertTriangle } from 'lucide-vue-next'

interface Props {
  title: string
  value: string | number
  trend?: string
  description?: string
  icon?: string
  color?: string
}

const props = defineProps<Props>()

const getIconComponent = () => {
  switch (props.icon) {
    case 'Clock': return Clock
    case 'Route': return Route
    case 'TrendingUp': return TrendingUp
    case 'AlertTriangle': return AlertTriangle
    default: return Clock
  }
}

const getIconColor = () => {
  switch (props.color) {
    case 'error': return 'text-red-500'
    case 'warning': return 'text-yellow-500'
    case 'success': return 'text-green-500'
    default: return 'text-blue-500'
  }
}
</script>