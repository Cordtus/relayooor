<template>
  <div class="bg-white shadow rounded-lg p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Your Clearing Statistics</h3>
    
    <div v-if="loading" class="text-center py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
    </div>
    
    <div v-else-if="error" class="text-center py-8 text-red-600">
      <AlertCircle class="h-8 w-8 mx-auto mb-2" />
      <p>Failed to load statistics</p>
    </div>
    
    <div v-else class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="text-center">
        <p class="text-2xl font-semibold text-gray-900">{{ stats?.totalRequests || 0 }}</p>
        <p class="text-sm text-gray-600">Total Requests</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-semibold text-green-600">{{ stats?.successfulClears || 0 }}</p>
        <p class="text-sm text-gray-600">Successful</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-semibold text-blue-600">{{ stats?.totalPacketsCleared || 0 }}</p>
        <p class="text-sm text-gray-600">Packets Cleared</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-semibold text-gray-900">
          {{ formatSuccessRate(stats?.successRate) }}%
        </p>
        <p class="text-sm text-gray-600">Success Rate</p>
      </div>
    </div>
    
    <div v-if="stats?.mostActiveChannels?.length" class="mt-6 pt-6 border-t">
      <h4 class="text-sm font-medium text-gray-700 mb-3">Most Active Channels</h4>
      <div class="space-y-2">
        <div 
          v-for="channel in stats.mostActiveChannels.slice(0, 3)" 
          :key="channel.channel"
          class="flex justify-between text-sm"
        >
          <span class="text-gray-600">{{ formatChannelName(channel.channel) }}</span>
          <span class="font-medium">{{ channel.count }} clears</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { AlertCircle } from 'lucide-vue-next'
import { clearingService, type UserStatistics } from '@/services/clearing'

const loading = ref(true)
const error = ref(false)
const stats = ref<UserStatistics | null>(null)

onMounted(async () => {
  try {
    stats.value = await clearingService.getUserStatistics()
  } catch (err) {
    error.value = true
    console.error('Failed to load user statistics:', err)
  } finally {
    loading.value = false
  }
})

const formatSuccessRate = (rate?: number): string => {
  if (!rate) return '0'
  return (rate * 100).toFixed(1)
}

const formatChannelName = (channel: string): string => {
  // Format: "osmosis-1/channel-0->cosmoshub-4/channel-141"
  const parts = channel.split('->')
  if (parts.length !== 2) return channel
  
  const src = parts[0].split('/')[1] || parts[0]
  const dst = parts[1].split('/')[1] || parts[1]
  
  return `${src} → ${dst}`
}
</script>