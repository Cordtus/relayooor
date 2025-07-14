<template>
  <div
    v-show="isActive"
    :id="id"
    role="tabpanel"
    :aria-labelledby="`${id}-trigger`"
    class="mt-2"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { inject, computed } from 'vue'

interface Props {
  value: string
}

const props = defineProps<Props>()

const activeTab = inject<{ value: string }>('activeTab')
const id = computed(() => `tab-${props.value}`)
const isActive = computed(() => activeTab?.value === props.value)
</script>