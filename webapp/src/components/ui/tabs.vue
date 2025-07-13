<template>
  <TabsRoot v-bind="props">
    <slot />
  </TabsRoot>
</template>

<script setup lang="ts">
import { TabsRoot } from 'radix-vue'
import type { TabsRootProps } from 'radix-vue'

const props = defineProps<TabsRootProps>()
</script>

<script lang="ts">
export { default as TabsList } from './tabs-list.vue'
export { default as TabsTrigger } from './tabs-trigger.vue'
export { default as TabsContent } from './tabs-content.vue'
</script>