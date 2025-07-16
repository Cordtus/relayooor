<template>
  <div class="empty-state">
    <component v-if="icon" :is="icon" class="empty-state-icon" />
    <h3 class="empty-state-title">{{ title }}</h3>
    <p v-if="description" class="empty-state-description">{{ description }}</p>
    <div v-if="$slots.action" class="mt-6">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

interface Props {
  icon?: Component
  title: string
  description?: string
}

defineProps<Props>()
</script>