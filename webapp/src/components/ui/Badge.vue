<template>
  <span :class="badgeClasses">
    <span v-if="dot" :class="dotClasses" />
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  variant?: 'success' | 'warning' | 'error' | 'info' | 'default'
  size?: 'sm' | 'md'
  dot?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'sm',
  dot: false
})

const badgeClasses = computed(() => {
  const baseClasses = 'inline-flex items-center font-medium rounded-full'
  
  const variantClasses = {
    success: 'bg-status-success-light text-status-success-dark',
    warning: 'bg-status-warning-light text-status-warning-dark',
    error: 'bg-status-error-light text-status-error-dark',
    info: 'bg-status-info-light text-status-info-dark',
    default: 'bg-gray-100 text-gray-800'
  }
  
  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-3 py-1 text-sm'
  }
  
  return cn(
    baseClasses,
    variantClasses[props.variant],
    sizeClasses[props.size]
  )
})

const dotClasses = computed(() => {
  const baseClasses = 'w-2 h-2 rounded-full mr-1.5'
  
  const variantClasses = {
    success: 'bg-status-success',
    warning: 'bg-status-warning',
    error: 'bg-status-error',
    info: 'bg-status-info',
    default: 'bg-gray-500'
  }
  
  return cn(baseClasses, variantClasses[props.variant])
})
</script>