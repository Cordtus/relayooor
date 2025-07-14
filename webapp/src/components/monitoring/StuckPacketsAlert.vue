<template>
  <div v-if="packets && packets.length > 0" class="bg-red-50 border border-red-200 rounded-lg p-4">
    <div class="flex items-start">
      <AlertTriangle class="h-5 w-5 text-red-600 mt-0.5" />
      <div class="ml-3 flex-1">
        <h3 class="text-sm font-medium text-red-800">
          {{ packets.length }} Stuck {{ packets.length === 1 ? 'Packet' : 'Packets' }} Detected
        </h3>
        <div class="mt-2 text-sm text-red-700">
          <p>The following packets have been stuck for an extended period:</p>
        </div>
        
        <!-- Stuck Packets List -->
        <div class="mt-3 space-y-2">
          <div 
            v-for="packet in displayedPackets" 
            :key="`${packet.chain}-${packet.channel}-${packet.sequence}`"
            class="bg-white rounded-md p-3 border border-red-100"
          >
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-900">
                    {{ formatChainName(packet.src_chain || packet.chain) }}
                  </span>
                  <ArrowRight class="h-4 w-4 text-gray-400" />
                  <span class="text-sm font-medium text-gray-900">
                    {{ formatChainName(packet.dst_chain || inferDestChain(packet)) }}
                  </span>
                </div>
                <div class="mt-1 flex items-center gap-4 text-xs text-gray-500">
                  <span>Channel: {{ packet.src_channel || packet.channel }}</span>
                  <span>Sequence: #{{ packet.sequence }}</span>
                  <span v-if="packet.stuck_duration">
                    Stuck for: {{ formatDuration(packet.stuck_duration) }}
                  </span>
                </div>
              </div>
              <button
                @click="$emit('clear', packet)"
                class="ml-4 inline-flex items-center px-3 py-1.5 border border-red-300 text-xs font-medium rounded text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              >
                Clear Packet
              </button>
            </div>
          </div>
        </div>
        
        <!-- Show More/Less -->
        <div v-if="packets.length > 3" class="mt-3">
          <button
            @click="showAll = !showAll"
            class="text-sm text-red-600 hover:text-red-700 font-medium"
          >
            {{ showAll ? 'Show Less' : `Show ${packets.length - 3} More` }}
          </button>
        </div>
        
        <!-- Bulk Clear Option -->
        <div v-if="packets.length > 1" class="mt-4 pt-4 border-t border-red-200">
          <button
            @click="clearAllPackets"
            class="inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
          >
            <Zap class="h-4 w-4 mr-2" />
            Clear All Stuck Packets
          </button>
        </div>
      </div>
    </div>
  </div>
  
  <!-- No Stuck Packets -->
  <div v-else class="bg-green-50 border border-green-200 rounded-lg p-4">
    <div class="flex items-start">
      <CheckCircle class="h-5 w-5 text-green-600 mt-0.5" />
      <div class="ml-3">
        <h3 class="text-sm font-medium text-green-800">
          All Packets Flowing Smoothly
        </h3>
        <div class="mt-1 text-sm text-green-700">
          No stuck packets detected. All IBC transfers are processing normally.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { AlertTriangle, CheckCircle, ArrowRight, Zap } from 'lucide-vue-next'

interface StuckPacket {
  chain?: string
  src_chain?: string
  dst_chain?: string
  channel?: string
  src_channel?: string
  dst_channel?: string
  sequence: number
  stuck_duration?: number
  timestamp?: string
}

const props = defineProps<{
  packets?: StuckPacket[]
}>()

const emit = defineEmits<{
  clear: [packet: StuckPacket]
}>()

const showAll = ref(false)

const displayedPackets = computed(() => {
  if (!props.packets) return []
  return showAll.value ? props.packets : props.packets.slice(0, 3)
})

function formatChainName(chainId: string): string {
  const names: Record<string, string> = {
    'cosmoshub-4': 'Cosmos Hub',
    'osmosis-1': 'Osmosis',
    'neutron-1': 'Neutron',
    'stargaze-1': 'Stargaze',
    'juno-1': 'Juno'
  }
  return names[chainId] || chainId
}

function inferDestChain(packet: StuckPacket): string {
  // Infer destination based on channel mappings
  const srcChain = packet.src_chain || packet.chain || ''
  const channel = packet.src_channel || packet.channel || ''
  
  if (srcChain === 'osmosis-1' && channel === 'channel-0') return 'cosmoshub-4'
  if (srcChain === 'cosmoshub-4' && channel === 'channel-141') return 'osmosis-1'
  if (srcChain === 'neutron-1' && channel === 'channel-10') return 'osmosis-1'
  
  return 'Unknown'
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

function clearAllPackets() {
  // Emit clear event for each packet
  props.packets?.forEach(packet => {
    emit('clear', packet)
  })
}
</script>