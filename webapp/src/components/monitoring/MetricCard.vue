<template>
  <div class="card-base p-5">
    <div class="flex items-center">
      <div class="flex-shrink-0">
        <div :class="iconClasses">
          <component :is="icon" class="h-6 w-6 text-content-inverse" />
        </div>
      </div>
      <div class="ml-5 w-0 flex-1">
        <dl>
          <dt class="text-sm font-medium text-content-secondary truncate">
            {{ title }}
          </dt>
          <dd class="flex items-baseline">
            <div class="text-2xl font-semibold text-content-primary">
              {{ value }}
            </div>
            <div v-if="trend" class="ml-2 flex items-baseline text-sm font-semibold"
              :class="trend > 0 ? 'text-status-success' : 'text-status-error'">
              <component :is="trend > 0 ? 'TrendingUp' : 'TrendingDown'" class="h-4 w-4 flex-shrink-0" />
              <span class="ml-1">{{ Math.abs(trend) }}%</span>
            </div>
          </dd>
          <dd v-if="subtitle" class="text-sm text-content-secondary">{{ subtitle }}</dd>
        </dl>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { TrendingUp, TrendingDown } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

interface Props {
  title: string
  value: string | number
  subtitle?: string
  trend?: number
  icon: any
  color?: 'primary' | 'success' | 'error' | 'warning' | 'info'
}

const props = withDefaults(defineProps<Props>(), {
  color: 'primary'
})

const iconClasses = computed(() => {
  const colorMap = {
    primary: 'bg-primary-600',
    success: 'bg-status-success',
    error: 'bg-status-error',
    warning: 'bg-status-warning',
    info: 'bg-status-info'
  }
  
  return cn('rounded-md p-3', colorMap[props.color])
})
</script>