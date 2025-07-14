<template>
  <div class="bg-white rounded-lg shadow overflow-hidden">
    <div class="px-6 py-4 border-b border-gray-200">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-medium text-gray-900">Relayer Leaderboard</h3>
        <select v-model="timeframe" class="text-sm border-gray-300 rounded-md">
          <option value="1h">Last Hour</option>
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
        </select>
      </div>
    </div>
    
    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Rank
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Relayer
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Software
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Packets
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Success Rate
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Efficiency
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Fees Spent
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="(relayer, index) in topRelayers" 
              :key="relayer.address"
              class="hover:bg-gray-50 transition-colors">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <span v-if="index < 3" class="text-lg mr-2">
                  {{ index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉' }}
                </span>
                <span class="text-sm font-medium text-gray-900">{{ index + 1 }}</span>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div>
                  <div class="text-sm font-medium text-gray-900">
                    {{ formatAddress(relayer.address) }}
                  </div>
                  <div v-if="relayer.memo" class="text-xs text-gray-500">
                    {{ relayer.memo }}
                  </div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                    :class="getSoftwareClass(relayer.software)">
                {{ relayer.software }} {{ relayer.version }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
              <div>
                <div class="font-medium">{{ formatNumber(relayer.totalPackets) }}</div>
                <div class="text-xs text-gray-500">
                  ✓ {{ formatNumber(relayer.effectedPackets) }} | ✗ {{ formatNumber(relayer.uneffectedPackets) }}
                </div>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <span class="text-sm font-medium mr-2"
                      :class="getSuccessRateClass(relayer.successRate)">
                  {{ relayer.successRate.toFixed(1) }}%
                </span>
                <div class="w-16 bg-gray-200 rounded-full h-2">
                  <div class="h-2 rounded-full"
                       :class="getSuccessRateBarClass(relayer.successRate)"
                       :style="{ width: `${relayer.successRate}%` }">
                  </div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm text-gray-900">
                {{ calculateEfficiency(relayer).toFixed(2) }}
              </div>
              <div class="text-xs text-gray-500">
                {{ relayer.frontrunCount }} frontrun{{ relayer.frontrunCount !== 1 ? 's' : '' }}
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
              {{ estimateFeesSpent(relayer) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <div v-if="topRelayers.length === 0" class="text-center py-8">
      <p class="text-gray-500">No relayer data available</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Relayer {
  address: string
  totalPackets: number
  effectedPackets: number
  uneffectedPackets: number
  frontrunCount: number
  successRate: number
  memo?: string
  software: string
  version: string
}

const props = defineProps<{
  relayers: Relayer[]
}>()

const timeframe = ref('24h')

const topRelayers = computed(() => {
  if (!props.relayers) return []
  
  // Sort by total packets and take top 10
  return [...props.relayers]
    .sort((a, b) => b.totalPackets - a.totalPackets)
    .slice(0, 10)
})

function formatAddress(address: string): string {
  if (!address || address.length < 15) return address
  return address.slice(0, 10) + '...' + address.slice(-4)
}

function formatNumber(num: number): string {
  return new Intl.NumberFormat().format(num)
}

function getSoftwareClass(software: string): string {
  switch (software?.toLowerCase()) {
    case 'hermes':
      return 'bg-blue-100 text-blue-800'
    case 'go relayer':
    case 'rly':
      return 'bg-green-100 text-green-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

function getSuccessRateClass(rate: number): string {
  if (rate >= 95) return 'text-green-600'
  if (rate >= 85) return 'text-yellow-600'
  return 'text-red-600'
}

function getSuccessRateBarClass(rate: number): string {
  if (rate >= 95) return 'bg-green-500'
  if (rate >= 85) return 'bg-yellow-500'
  return 'bg-red-500'
}

function calculateEfficiency(relayer: Relayer): number {
  // Efficiency score based on success rate and frontrun ratio
  const frontrunPenalty = Math.min(relayer.frontrunCount / relayer.totalPackets * 100, 20)
  return Math.max(relayer.successRate - frontrunPenalty, 0)
}

function estimateFeesSpent(relayer: Relayer): string {
  // Estimate fees spent by relayer - in production, use real gas data
  const avgGasFeePerPacket = 0.05 // USD estimate for gas fees
  const totalFees = relayer.totalPackets * avgGasFeePerPacket
  
  if (totalFees >= 1000) {
    return `$${(totalFees / 1000).toFixed(1)}k`
  }
  return `$${totalFees.toFixed(0)}`
}
</script>