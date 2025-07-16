<template>
  <div class="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow">
    <div class="flex items-start justify-between mb-4">
      <span :class="[
        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
        recommendation.priority === 'high' ? 'bg-red-100 text-red-800' :
        recommendation.priority === 'medium' ? 'bg-yellow-100 text-yellow-800' :
        'bg-green-100 text-green-800'
      ]">
        {{ recommendation.priority }} priority
      </span>
      <span class="text-sm text-gray-500">{{ recommendation.type }}</span>
    </div>
    
    <h3 class="text-lg font-medium text-gray-900 mb-2">{{ recommendation.title }}</h3>
    <p class="text-sm text-gray-600 mb-4">{{ recommendation.description }}</p>
    
    <div class="flex items-center justify-between text-sm">
      <span class="text-green-600 font-medium">{{ recommendation.impact }}</span>
      <span class="text-gray-500">Effort: {{ recommendation.effort }}</span>
    </div>
    
    <button
      @click="$emit('implement', recommendation)"
      class="mt-4 w-full bg-indigo-600 text-white rounded-md py-2 text-sm font-medium hover:bg-indigo-700 transition-colors flex items-center justify-center gap-2"
    >
      <component :is="getActionIcon(recommendation.type)" class="h-4 w-4" />
      {{ getActionText(recommendation.type) }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { Settings, Zap, Clock, TrendingUp } from 'lucide-vue-next'

interface Recommendation {
  id: number
  type: string
  priority: 'high' | 'medium' | 'low'
  title: string
  description: string
  impact: string
  effort: string
}

interface Props {
  recommendation: Recommendation
}

const props = defineProps<Props>()
const emit = defineEmits(['implement'])

function getActionIcon(type: string) {
  switch (type) {
    case 'config':
      return Settings
    case 'action':
      return Zap
    case 'timing':
      return Clock
    default:
      return TrendingUp
  }
}

function getActionText(type: string): string {
  switch (type) {
    case 'config':
      return 'View Configuration'
    case 'action':
      return 'Take Action'
    case 'timing':
      return 'Optimize Schedule'
    case 'channel':
      return 'Analyze Channel'
    default:
      return 'View Details'
  }
}
</script>