<template>
  <div class="bg-white overflow-hidden shadow rounded-lg">
    <div class="p-5">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <div :class="[
            'rounded-md p-3',
            color === 'blue' ? 'bg-blue-500' : '',
            color === 'green' ? 'bg-green-500' : '',
            color === 'red' ? 'bg-red-500' : '',
            color === 'yellow' ? 'bg-yellow-500' : '',
            color === 'purple' ? 'bg-purple-500' : '',
            color === 'orange' ? 'bg-orange-500' : ''
          ]">
            <component :is="icon" class="h-6 w-6 text-white" />
          </div>
        </div>
        <div class="ml-5 w-0 flex-1">
          <dl>
            <dt class="text-sm font-medium text-gray-500 truncate">
              {{ title }}
            </dt>
            <dd class="flex items-baseline">
              <div class="text-2xl font-semibold text-gray-900">
                {{ value }}
              </div>
              <div v-if="trend" class="ml-2 flex items-baseline text-sm font-semibold"
                :class="trend > 0 ? 'text-green-600' : 'text-red-600'">
                <component :is="trend > 0 ? 'TrendingUp' : 'TrendingDown'" class="h-4 w-4 flex-shrink-0" />
                <span class="ml-1">{{ Math.abs(trend) }}%</span>
              </div>
            </dd>
            <dd v-if="subtitle" class="text-sm text-gray-500">{{ subtitle }}</dd>
          </dl>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { TrendingUp, TrendingDown } from 'lucide-vue-next'

interface Props {
  title: string
  value: string | number
  subtitle?: string
  trend?: number
  icon: any
  color?: 'blue' | 'green' | 'red' | 'yellow' | 'purple' | 'orange'
}

const props = withDefaults(defineProps<Props>(), {
  color: 'blue'
})
</script>