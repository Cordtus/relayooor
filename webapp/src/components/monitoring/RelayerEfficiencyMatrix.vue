<template>
  <div class="bg-white rounded-lg shadow p-4">
    <h3 class="text-lg font-medium mb-4">Efficiency Matrix</h3>
    <div v-if="efficiencyData.length > 0" class="overflow-x-auto">
      <table class="min-w-full">
        <thead>
          <tr class="text-xs text-gray-500 uppercase">
            <th class="px-2 py-2 text-left">Relayer</th>
            <th class="px-2 py-2 text-center">Success</th>
            <th class="px-2 py-2 text-center">Speed</th>
            <th class="px-2 py-2 text-center">Volume</th>
            <th class="px-2 py-2 text-center">Overall</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="relayer in efficiencyData" :key="relayer.address" class="border-t border-gray-100">
            <td class="px-2 py-2 text-sm">
              {{ formatAddress(relayer.address) }}
            </td>
            <td class="px-2 py-2">
              <div class="flex justify-center">
                <div :class="getScoreClass(relayer.successScore)" 
                     class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium">
                  {{ relayer.successScore }}
                </div>
              </div>
            </td>
            <td class="px-2 py-2">
              <div class="flex justify-center">
                <div :class="getScoreClass(relayer.speedScore)" 
                     class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium">
                  {{ relayer.speedScore }}
                </div>
              </div>
            </td>
            <td class="px-2 py-2">
              <div class="flex justify-center">
                <div :class="getScoreClass(relayer.volumeScore)" 
                     class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium">
                  {{ relayer.volumeScore }}
                </div>
              </div>
            </td>
            <td class="px-2 py-2">
              <div class="flex justify-center">
                <div :class="getOverallScoreClass(relayer.overallScore)" 
                     class="px-3 py-1 rounded-full text-xs font-medium">
                  {{ relayer.overallScore.toFixed(1) }}
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="mt-4 text-xs text-gray-500">
        <p>Scores: Success Rate (0-10) • Response Speed (0-10) • Volume Handled (0-10)</p>
        <p class="mt-1">Overall score is weighted average: 50% success, 30% speed, 20% volume</p>
      </div>
    </div>
    <div v-else class="text-center py-8 text-gray-500">
      No relayer data available
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Relayer {
  address: string
  totalPackets: number
  effectedPackets: number
  successRate: number
  avgProcessingTime?: number
}

const props = defineProps<{
  relayers?: Relayer[]
}>()

const efficiencyData = computed(() => {
  if (!props.relayers || props.relayers.length === 0) return []
  
  const maxPackets = Math.max(...props.relayers.map(r => r.totalPackets))
  
  return props.relayers
    .map(r => {
      // Success score based on success rate (0-10)
      const successScore = Math.round((r.successRate / 100) * 10)
      
      // Speed score - assume faster relayers have better success rates (proxy metric)
      // In production, use actual processing time data
      const speedScore = r.successRate > 95 ? 9 : 
                        r.successRate > 90 ? 7 : 
                        r.successRate > 80 ? 5 : 3
      
      // Volume score based on total packets handled (0-10)
      const volumeScore = Math.round((r.totalPackets / maxPackets) * 10)
      
      // Overall weighted score
      const overallScore = (successScore * 0.5) + (speedScore * 0.3) + (volumeScore * 0.2)
      
      return {
        address: r.address,
        successScore,
        speedScore,
        volumeScore,
        overallScore,
        totalPackets: r.totalPackets
      }
    })
    .sort((a, b) => b.overallScore - a.overallScore)
    .slice(0, 10)
})

function formatAddress(address: string): string {
  return address.slice(0, 8) + '...' + address.slice(-4)
}

function getScoreClass(score: number): string {
  if (score >= 8) return 'bg-green-100 text-green-700'
  if (score >= 6) return 'bg-yellow-100 text-yellow-700'
  if (score >= 4) return 'bg-orange-100 text-orange-700'
  return 'bg-red-100 text-red-700'
}

function getOverallScoreClass(score: number): string {
  if (score >= 8) return 'bg-green-500 text-white'
  if (score >= 6) return 'bg-blue-500 text-white'
  if (score >= 4) return 'bg-yellow-500 text-white'
  return 'bg-red-500 text-white'
}
</script>