<template>
  <div class="bg-white rounded-lg shadow p-4">
    <h4 class="text-sm font-medium text-gray-900 mb-3">Relayer Churn Analysis</h4>
    
    <!-- Summary Stats -->
    <div class="grid grid-cols-3 gap-2 mb-4">
      <div class="text-center">
        <p class="text-xs text-gray-500">New</p>
        <p class="text-lg font-semibold text-green-600">+{{ churnStats.new }}</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-gray-500">Active</p>
        <p class="text-lg font-semibold text-blue-600">{{ churnStats.active }}</p>
      </div>
      <div class="text-center">
        <p class="text-xs text-gray-500">Inactive</p>
        <p class="text-lg font-semibold text-red-600">-{{ churnStats.inactive }}</p>
      </div>
    </div>
    
    <!-- Churn Details -->
    <div class="space-y-2">
      <!-- New Relayers -->
      <div v-if="newRelayers.length > 0" class="border-t pt-2">
        <p class="text-xs font-medium text-gray-700 mb-1">New Relayers (Last 7d)</p>
        <div class="space-y-1">
          <div v-for="relayer in newRelayers.slice(0, 3)" :key="relayer.address" class="flex items-center justify-between text-xs">
            <span class="text-gray-600 truncate max-w-[150px]">{{ formatAddress(relayer.address) }}</span>
            <span class="text-green-600 font-medium">{{ relayer.totalPackets }} packets</span>
          </div>
        </div>
      </div>
      
      <!-- Churned Relayers -->
      <div v-if="churnedRelayers.length > 0" class="border-t pt-2">
        <p class="text-xs font-medium text-gray-700 mb-1">Recently Inactive</p>
        <div class="space-y-1">
          <div v-for="relayer in churnedRelayers.slice(0, 3)" :key="relayer.address" class="flex items-center justify-between text-xs">
            <span class="text-gray-600 truncate max-w-[150px]">{{ formatAddress(relayer.address) }}</span>
            <span class="text-gray-500">{{ formatTimeAgo(relayer.lastSeen) }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Trend Indicator -->
    <div class="mt-4 pt-3 border-t">
      <div class="flex items-center justify-between">
        <span class="text-xs text-gray-500">Churn Rate</span>
        <div class="flex items-center gap-1">
          <span class="text-sm font-medium" :class="getChurnRateClass()">{{ churnRate }}%</span>
          <component :is="getChurnTrendIcon()" class="h-3 w-3" :class="getChurnRateClass()" />
        </div>
      </div>
      <p class="text-xs text-gray-500 mt-1">{{ churnTrend }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { TrendingUp, TrendingDown, Minus } from 'lucide-vue-next'

interface Relayer {
  address: string
  totalPackets: number
  lastSeen?: Date
  firstSeen?: Date
}

const props = defineProps<{
  relayers?: Relayer[]
  historical?: Relayer[]
}>()

// Calculate churn statistics
const churnStats = computed(() => {
  const current = props.relayers || []
  const historical = props.historical || []
  
  const currentAddresses = new Set(current.map(r => r.address))
  const historicalAddresses = new Set(historical.map(r => r.address))
  
  // New relayers: in current but not in historical
  const newCount = current.filter(r => !historicalAddresses.has(r.address)).length
  
  // Inactive relayers: in historical but not in current
  const inactiveCount = historical.filter(r => !currentAddresses.has(r.address)).length
  
  // Active relayers: currently active
  const activeCount = current.filter(r => r.totalPackets > 0).length
  
  return {
    new: newCount,
    active: activeCount,
    inactive: inactiveCount
  }
})

// Get new relayers
const newRelayers = computed(() => {
  const historical = props.historical || []
  const current = props.relayers || []
  const historicalAddresses = new Set(historical.map(r => r.address))
  
  return current
    .filter(r => !historicalAddresses.has(r.address))
    .sort((a, b) => b.totalPackets - a.totalPackets)
})

// Get churned relayers
const churnedRelayers = computed(() => {
  const current = props.relayers || []
  const historical = props.historical || []
  const currentAddresses = new Set(current.map(r => r.address))
  
  return historical
    .filter(r => !currentAddresses.has(r.address))
    .filter(r => r.lastSeen) // Only include relayers with actual lastSeen data
    .sort((a, b) => (b.lastSeen?.getTime() || 0) - (a.lastSeen?.getTime() || 0))
})

// Calculate churn rate
const churnRate = computed(() => {
  const total = (props.historical || []).length
  if (total === 0) return 0
  
  const churned = churnedRelayers.value.length
  return Math.round((churned / total) * 100)
})

// Determine churn trend
const churnTrend = computed(() => {
  const rate = churnRate.value
  if (rate < 5) return 'Very low churn - stable relayer pool'
  if (rate < 10) return 'Normal churn rate'
  if (rate < 20) return 'Elevated churn - monitor for issues'
  return 'High churn - investigate causes'
})

function formatAddress(address: string): string {
  return `${address.slice(0, 8)}...${address.slice(-6)}`
}

function formatTimeAgo(date: Date): string {
  const hours = Math.floor((Date.now() - date.getTime()) / (1000 * 60 * 60))
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

function getChurnRateClass(): string {
  const rate = churnRate.value
  if (rate < 5) return 'text-green-600'
  if (rate < 10) return 'text-gray-600'
  if (rate < 20) return 'text-yellow-600'
  return 'text-red-600'
}

function getChurnTrendIcon() {
  const rate = churnRate.value
  if (rate < 5) return TrendingDown
  if (rate > 15) return TrendingUp
  return Minus
}
</script>