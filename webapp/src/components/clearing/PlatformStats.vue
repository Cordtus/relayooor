<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
    <!-- Total Packets Cleared -->
    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <div class="h-12 w-12 rounded-full bg-green-100 flex items-center justify-center">
            <CheckCircle class="h-6 w-6 text-green-600" />
          </div>
        </div>
        <div class="ml-4">
          <p class="text-sm font-medium text-gray-600">Total Cleared</p>
          <p class="text-2xl font-semibold text-gray-900">
            {{ formatNumber(stats?.global.totalPacketsCleared || 0) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Active Users -->
    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <div class="h-12 w-12 rounded-full bg-blue-100 flex items-center justify-center">
            <Users class="h-6 w-6 text-blue-600" />
          </div>
        </div>
        <div class="ml-4">
          <p class="text-sm font-medium text-gray-600">Active Users</p>
          <p class="text-2xl font-semibold text-gray-900">
            {{ formatNumber(stats?.global.totalUsers || 0) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Success Rate -->
    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <div class="h-12 w-12 rounded-full bg-purple-100 flex items-center justify-center">
            <TrendingUp class="h-6 w-6 text-purple-600" />
          </div>
        </div>
        <div class="ml-4">
          <p class="text-sm font-medium text-gray-600">Success Rate</p>
          <p class="text-2xl font-semibold text-gray-900">
            {{ (stats?.global.successRate || 0) * 100 }}%
          </p>
        </div>
      </div>
    </div>

    <!-- Daily Activity -->
    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <div class="h-12 w-12 rounded-full bg-orange-100 flex items-center justify-center">
            <Activity class="h-6 w-6 text-orange-600" />
          </div>
        </div>
        <div class="ml-4">
          <p class="text-sm font-medium text-gray-600">Today's Clears</p>
          <p class="text-2xl font-semibold text-gray-900">
            {{ formatNumber(stats?.daily.packetsCleared || 0) }}
          </p>
        </div>
      </div>
    </div>
  </div>

  <!-- Top Channels -->
  <div v-if="stats?.topChannels?.length" class="mt-6 bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-gray-200">
      <h3 class="text-lg font-medium text-gray-900">Most Active Channels</h3>
    </div>
    <div class="p-6">
      <div class="space-y-4">
        <div 
          v-for="(channel, index) in stats.topChannels.slice(0, 5)" 
          :key="channel.channel"
          class="flex items-center justify-between"
        >
          <div class="flex items-center">
            <span class="text-sm font-medium text-gray-500 w-6">{{ index + 1 }}.</span>
            <span class="text-sm text-gray-900 ml-2">{{ formatChannelName(channel.channel) }}</span>
          </div>
          <div class="text-right">
            <p class="text-sm font-medium text-gray-900">{{ formatNumber(channel.packetsCleared) }} packets</p>
            <p class="text-xs text-gray-500">~{{ (channel.avgClearTime / 1000).toFixed(1) }}s avg</p>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Peak Hours Chart -->
  <div v-if="stats?.peakHours?.length" class="mt-6 bg-white rounded-lg shadow p-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Activity by Hour (UTC)</h3>
    <div class="flex items-end justify-between h-32">
      <div 
        v-for="hour in hourlyActivity" 
        :key="hour.hour"
        class="flex-1 mx-0.5"
      >
        <div 
          class="bg-blue-500 rounded-t hover:bg-blue-600 transition-colors"
          :style="{ height: `${hour.percentage}%` }"
          :title="`${hour.hour}:00 - ${hour.activity} operations`"
        ></div>
      </div>
    </div>
    <div class="flex justify-between mt-2 text-xs text-gray-500">
      <span>00:00</span>
      <span>06:00</span>
      <span>12:00</span>
      <span>18:00</span>
      <span>23:00</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { CheckCircle, Users, TrendingUp, Activity } from 'lucide-vue-next'
import { clearingService, type PlatformStatistics } from '@/services/clearing'

const stats = ref<PlatformStatistics | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    stats.value = await clearingService.getPlatformStatistics()
  } catch (error) {
    console.error('Failed to load platform statistics:', error)
  } finally {
    loading.value = false
  }
})

const hourlyActivity = computed(() => {
  if (!stats.value?.peakHours) return []
  
  // Create 24-hour array
  const hours = Array.from({ length: 24 }, (_, i) => ({
    hour: i,
    activity: 0,
    percentage: 0
  }))
  
  // Fill with actual data
  const maxActivity = Math.max(...stats.value.peakHours.map(h => h.activity))
  
  stats.value.peakHours.forEach(h => {
    hours[h.hour] = {
      hour: h.hour,
      activity: h.activity,
      percentage: (h.activity / maxActivity) * 100
    }
  })
  
  return hours
})

const formatNumber = (num: number): string => {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
  return num.toString()
}

const formatChannelName = (channel: string): string => {
  // Simplify channel display
  const match = channel.match(/(\w+-\d+)\/channel-(\d+)->(\w+-\d+)\/channel-(\d+)/)
  if (match) {
    return `${match[1]} ch-${match[2]} → ${match[3]} ch-${match[4]}`
  }
  return channel
}
</script>