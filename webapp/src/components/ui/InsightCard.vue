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
    <div v-if="icon" class="mt-4">
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
import { Clock, Route, TrendingUp, AlertTriangle } from 'lucide-vue-next'

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
    // Handle dynamic color classes properly
    const colorMap: Record<string, string> = {
      'error': 'text-red-600',
      'warning': 'text-yellow-600',
      'success': 'text-green-600',
      'blue': 'text-blue-600',
      'gray': 'text-gray-600'
    }
    return colorMap[props.color] || 'text-gray-900'
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