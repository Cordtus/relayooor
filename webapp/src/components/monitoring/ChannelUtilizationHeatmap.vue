<template>
  <div class="bg-white p-4 rounded-lg">
    <h3 class="text-lg font-medium mb-4">Channel Utilization Heatmap</h3>
    <p class="text-xs text-gray-500 mb-3">Shows direct IBC channels between chains. Some packets may use multi-hop routes not shown here.</p>
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
        <span>High (>10k packets)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-yellow-500 rounded mr-1"></div>
        <span>Medium (1k-10k)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-orange-500 rounded mr-1"></div>
        <span>Low (100-1k)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-red-500 rounded mr-1"></div>
        <span>Minimal (<100)</span>
      </div>
      <div class="flex items-center">
        <div class="w-4 h-4 bg-gray-300 rounded mr-1"></div>
        <span>No Direct Channel</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import type { ChannelMetrics } from '@/types/monitoring'
import { resolveChannels, type ChannelInfo } from '@/services/channel-resolver'

interface Props {
  channels?: ChannelMetrics[]
}

const props = defineProps<Props>()

// Store resolved channel information
const resolvedChannels = ref<Map<string, ChannelInfo>>(new Map())

const uniqueChains = computed(() => {
  if (!props.channels) return []
  const chains = new Set<string>()
  
  // First add all source chains
  props.channels.forEach(channel => {
    chains.add(channel.srcChain)
    
    // Add destination chains, including inferred ones
    if (channel.dstChain && channel.dstChain !== 'unknown') {
      chains.add(channel.dstChain)
    } else {
      // Try to infer destination for unknown channels
      const inferred = inferDestinationChain(channel.dstChannel)
      if (inferred) {
        chains.add(inferred)
      }
    }
  })
  
  // Always include major chains to show the full picture
  const majorChains = ['cosmoshub-4', 'osmosis-1', 'noble-1', 'neutron-1', 'axelar-dojo-1', 'stride-1']
  majorChains.forEach(chain => chains.add(chain))
  
  return Array.from(chains).filter(c => c !== 'unknown').sort()
})

const channelMap = computed(() => {
  const map = new Map<string, { totalPackets: number, channels: ChannelMetrics[] }>()
  
  props.channels?.forEach(channel => {
    let dstChain = channel.dstChain
    
    // If destination is unknown, check resolved channels first
    if (!dstChain || dstChain === 'unknown') {
      const resolvedKey = `${channel.srcChain}:${channel.srcChannel}`
      const resolved = resolvedChannels.value.get(resolvedKey)
      
      if (resolved) {
        dstChain = resolved.counterpartyChainId
      } else {
        // Fall back to inference from channel ID patterns
        const inferredDst = inferDestinationChain(channel.dstChannel)
        if (inferredDst) {
          dstChain = inferredDst
        } else {
          return // Skip this channel
        }
      }
    }
    
    const key = `${channel.srcChain}-${dstChain}`
    const existing = map.get(key)
    
    if (existing) {
      existing.totalPackets += channel.totalPackets
      existing.channels.push(channel)
    } else {
      map.set(key, {
        totalPackets: channel.totalPackets,
        channels: [channel]
      })
    }
  })
  
  return map
})

function inferDestinationChain(dstChannel: string): string | null {
  // Common channel mappings based on typical IBC channel patterns
  const channelMappings: Record<string, string> = {
    // Cosmos Hub channels
    'channel-0': 'cosmoshub-4',
    'channel-141': 'cosmoshub-4',
    'channel-301': 'cosmoshub-4',
    
    // Noble channels
    'channel-1': 'noble-1',
    'channel-750': 'noble-1',
    'channel-169': 'noble-1',
    
    // Osmosis channels
    'channel-42': 'osmosis-1',
    'channel-251': 'osmosis-1',
    'channel-362': 'osmosis-1',
    
    // Neutron channels
    'channel-874': 'neutron-1',
    'channel-653': 'neutron-1',
    
    // Axelar channels
    'channel-208': 'axelar-dojo-1',
    'channel-4': 'axelar-dojo-1',
    
    // Stride channels
    'channel-326': 'stride-1',
    'channel-391': 'stride-1'
  }
  return channelMappings[dstChannel] || null
}

function getChainShortName(chainId: string): string {
  const names: Record<string, string> = {
    'cosmoshub-4': 'Cosmos',
    'osmosis-1': 'Osmosis',
    'neutron-1': 'Neutron'
  }
  return names[chainId] || chainId
}

function getCellValue(src: string, dst: string): string {
  const data = channelMap.value.get(`${src}-${dst}`)
  if (!data || data.totalPackets === 0) return ''
  return data.totalPackets > 1000 ? `${(data.totalPackets / 1000).toFixed(1)}k` :
         data.totalPackets.toString()
}

function getCellClass(src: string, dst: string): string {
  const data = channelMap.value.get(`${src}-${dst}`)
  if (!data || data.totalPackets === 0) return 'bg-gray-300 text-gray-600'
  
  if (data.totalPackets > 10000) return 'bg-green-500 text-white'
  if (data.totalPackets > 1000) return 'bg-yellow-500 text-white'
  if (data.totalPackets > 100) return 'bg-orange-500 text-white'
  if (data.totalPackets > 0) return 'bg-red-500 text-white'
  return 'bg-gray-300 text-gray-600'
}

function getCellTooltip(src: string, dst: string): string {
  const data = channelMap.value.get(`${src}-${dst}`)
  if (!data || data.totalPackets === 0) return 'No channel activity'
  
  const channelList = data.channels
    .map(ch => `${ch.srcChannel} → ${ch.dstChannel}`)
    .join('\n')
  
  return `Total Packets: ${data.totalPackets.toLocaleString()}\n` +
         `Channels (${data.channels.length}):\n${channelList}`
}

// Resolve unknown channels dynamically
async function resolveUnknownChannels() {
  if (!props.channels) return
  
  // Find channels with unknown destinations
  const unknownChannels = props.channels
    .filter(ch => !ch.dstChain || ch.dstChain === 'unknown')
    .map(ch => ({
      sourceChainId: ch.srcChain,
      channelId: ch.srcChannel,
      portId: 'transfer'
    }))
  
  if (unknownChannels.length === 0) return
  
  try {
    // Resolve channels in batches
    const resolved = await resolveChannels(unknownChannels)
    resolvedChannels.value = resolved
  } catch (error) {
    console.warn('Failed to resolve some channels:', error)
  }
}

// Watch for channel changes and resolve unknown ones
watch(() => props.channels, async () => {
  await resolveUnknownChannels()
}, { immediate: true })

onMounted(async () => {
  await resolveUnknownChannels()
})
</script>