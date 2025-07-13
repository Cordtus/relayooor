<template>
  <div class="bg-white p-4 rounded-lg">
    <h3 class="text-lg font-medium mb-4">Channel Utilization Heatmap</h3>
    <div class="overflow-x-auto">
      <table class="min-w-full">
        <thead>
          <tr>
            <th class="text-xs text-gray-500 font-normal text-left p-2">From/To</th>
            <th v-for="chain in uniqueChains" :key="chain" 
              class="text-xs text-gray-500 font-normal text-center p-2 min-w-[100px]">
              {{ getChainShortName(chain) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="srcChain in uniqueChains" :key="srcChain">
            <td class="text-xs font-medium text-gray-700 p-2">
              {{ getChainShortName(srcChain) }}
            </td>
            <td v-for="dstChain in uniqueChains" :key="dstChain"
              class="p-2 text-center">
              <div v-if="srcChain !== dstChain"
                class="h-12 w-full rounded flex items-center justify-center text-xs font-medium cursor-pointer hover:opacity-80 transition-opacity"
                :class="getCellClass(srcChain, dstChain)"
                :title="getCellTooltip(srcChain, dstChain)">
                {{ getCellValue(srcChain, dstChain) }}
              </div>
              <div v-else class="h-12 w-full bg-gray-100 rounded"></div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- Legend -->
    <div class="mt-4 flex items-center justify-center space-x-4 text-xs">
      <div class="flex items-center">
        <div class="w-4 h-4 bg-green-500 rounded mr-1"></div>
        <span>High (>1000)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-yellow-500 rounded mr-1"></div>
        <span>Medium (100-1000)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-red-500 rounded mr-1"></div>
        <span>Low (<100)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-gray-300 rounded mr-1"></div>
        <span>No Activity</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ChannelMetrics } from '@/types/monitoring'

interface Props {
  channels?: ChannelMetrics[]
}

const props = defineProps<Props>()

const uniqueChains = computed(() => {
  if (!props.channels) return []
  const chains = new Set<string>()
  props.channels.forEach(channel => {
    chains.add(channel.srcChain)
    chains.add(channel.dstChain)
  })
  return Array.from(chains).sort()
})

const channelMap = computed(() => {
  const map = new Map<string, ChannelMetrics>()
  props.channels?.forEach(channel => {
    const key = `${channel.srcChain}-${channel.dstChain}`
    map.set(key, channel)
  })
  return map
})

function getChainShortName(chainId: string): string {
  const names: Record<string, string> = {
    'cosmoshub-4': 'Cosmos',
    'osmosis-1': 'Osmosis',
    'neutron-1': 'Neutron'
  }
  return names[chainId] || chainId
}

function getCellValue(src: string, dst: string): string {
  const channel = channelMap.value.get(`${src}-${dst}`)
  if (!channel) return ''
  return channel.totalPackets > 1000 ? `${(channel.totalPackets / 1000).toFixed(1)}k` :
         channel.totalPackets.toString()
}

function getCellClass(src: string, dst: string): string {
  const channel = channelMap.value.get(`${src}-${dst}`)
  if (!channel) return 'bg-gray-300 text-gray-600'
  
  if (channel.totalPackets > 1000) return 'bg-green-500 text-white'
  if (channel.totalPackets > 100) return 'bg-yellow-500 text-white'
  if (channel.totalPackets > 0) return 'bg-red-500 text-white'
  return 'bg-gray-300 text-gray-600'
}

function getCellTooltip(src: string, dst: string): string {
  const channel = channelMap.value.get(`${src}-${dst}`)
  if (!channel) return 'No channel activity'
  return `${channel.srcChannel} → ${channel.dstChannel}\n` +
         `Packets: ${channel.totalPackets.toLocaleString()}\n` +
         `Success Rate: ${channel.successRate.toFixed(1)}%`
}
</script>