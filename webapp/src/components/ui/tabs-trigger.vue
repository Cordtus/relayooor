<template>
  <TabsTrigger
    :class="cn(
      'inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm',
      props.class
    )"
    v-bind="props"
  >
    <slot />
  </TabsTrigger>
</template>

<script setup lang="ts">
import { TabsTrigger } from 'radix-vue'
import type { TabsTriggerProps } from 'radix-vue'
import { cn } from '@/lib/utils'

interface Props extends TabsTriggerProps {
  class?: string
}

const props = defineProps<Props>()
</script>