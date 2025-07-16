<template>
  <div class="packet-selector">
    <!-- Summary stats -->
    <div class="grid grid-cols-3 gap-4 mb-6">
      <div class="bg-blue-50 rounded-lg p-4">
        <p class="text-sm text-blue-600 font-medium">Stuck Packets</p>
        <p class="text-2xl font-bold text-blue-900">{{ stuckPackets.length }}</p>
      </div>
      <div class="bg-orange-50 rounded-lg p-4">
        <p class="text-sm text-orange-600 font-medium">Selected</p>
        <p class="text-2xl font-bold text-orange-900">{{ selected.length }}</p>
      </div>
      <div class="bg-green-50 rounded-lg p-4">
        <p class="text-sm text-green-600 font-medium">Total Value</p>
        <p class="text-lg font-bold text-green-900">{{ formatTotalValue() }}</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-4">
        <button
          @click="selectAll"
          class="text-sm text-blue-600 hover:text-blue-700 font-medium"
        >
          Select All
        </button>
        <button
          @click="selectNone"
          class="text-sm text-gray-600 hover:text-gray-700 font-medium"
        >
          Clear Selection
        </button>
      </div>
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <label class="text-sm text-gray-600">Chain:</label>
          <select 
            v-model="chainFilter"
            class="text-sm border border-gray-300 rounded-md px-2 py-1"
          >
            <option value="">All Chains</option>
            <option v-for="chain in availableChains" :key="chain" :value="chain">
              {{ getChainName(chain) }}
            </option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-sm text-gray-600">Sort by:</label>
          <select 
            v-model="sortBy"
            class="text-sm border border-gray-300 rounded-md px-2 py-1"
          >
            <option value="age">Age</option>
            <option value="value">Value</option>
            <option value="attempts">Attempts</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Packet list -->
    <div class="space-y-2 max-h-96 overflow-y-auto">
      <div
        v-for="packet in filteredAndSortedPackets"
        :key="packet.id"
        @click="togglePacket(packet)"
        :class="[
          'border rounded-lg p-4 cursor-pointer transition-colors',
          isSelected(packet) 
            ? 'border-blue-500 bg-blue-50' 
            : 'border-gray-200 hover:border-gray-300'
        ]"
      >
        <div class="flex items-start justify-between">
          <div class="flex items-start gap-3">
            <input
              type="checkbox"
              :checked="isSelected(packet)"
              @click.stop
              @change="togglePacket(packet)"
              class="mt-1 h-4 w-4 text-blue-600 rounded border-gray-300"
            />
            <div>
              <div class="flex items-center gap-2 mb-1">
                <span class="font-mono text-sm text-gray-900">
                  Sequence #{{ packet.sequence }}
                </span>
                <span class="text-xs text-gray-500">
                  {{ packet.channel }} • {{ getChainName(packet.chain) }}
                </span>
              </div>
              <div class="text-sm text-gray-600">
                <p>
                  From: <span class="font-mono">{{ formatAddress(packet.sender) }}</span>
                </p>
                <p>
                  To: <span class="font-mono">{{ formatAddress(packet.receiver) }}</span>
                </p>
              </div>
            </div>
          </div>
          <div class="text-right">
            <p class="font-medium text-gray-900">
              {{ formatAmount(packet.amount, packet.denom) }}
            </p>
            <p class="text-xs text-gray-500 mt-1">
              Stuck for {{ formatAge(packet.age) }}
            </p>
            <p class="text-xs text-orange-600" v-if="packet.attempts > 2">
              {{ packet.attempts }} failed attempts
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="stuckPackets.length === 0" class="text-center py-12">
      <PackageIcon class="w-12 h-12 text-gray-400 mx-auto mb-3" />
      <p class="text-gray-600">No stuck packets found</p>
      <p class="text-sm text-gray-500 mt-1">
        All your transfers are flowing smoothly!
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Package as PackageIcon } from 'lucide-vue-next'
import { clearingService } from '@/services/clearing'

interface StuckPacket {
  id: string
  chain: string
  channel: string
  sequence: number
  sender: string
  receiver: string
  amount: string
  denom: string
  age: number // seconds
  attempts: number
}

const props = defineProps<{
  stuckPackets: StuckPacket[]
  selected: StuckPacket[]
}>()

const emit = defineEmits<{
  'update:selected': [packets: StuckPacket[]]
}>()

const sortBy = ref<'age' | 'value' | 'attempts'>('age')
const chainFilter = ref<string>('')

const availableChains = computed(() => {
  const chains = new Set(props.stuckPackets.map(p => p.chain))
  return Array.from(chains).sort()
})

const filteredAndSortedPackets = computed(() => {
  let packets = [...props.stuckPackets]
  
  // Apply chain filter
  if (chainFilter.value) {
    packets = packets.filter(p => p.chain === chainFilter.value)
  }
  
  // Apply sorting
  switch (sortBy.value) {
    case 'age':
      return packets.sort((a, b) => b.age - a.age)
    case 'value':
      return packets.sort((a, b) => {
        const aValue = BigInt(a.amount)
        const bValue = BigInt(b.amount)
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0
      })
    case 'attempts':
      return packets.sort((a, b) => b.attempts - a.attempts)
    default:
      return packets
  }
})

// Helper to get chain name
function getChainName(chainId: string): string {
  const names: Record<string, string> = {
    'osmosis-1': 'Osmosis',
    'cosmoshub-4': 'Cosmos Hub',
    'neutron-1': 'Neutron'
  }
  return names[chainId] || chainId
}

const isSelected = (packet: StuckPacket): boolean => {
  return props.selected.some(p => p.id === packet.id)
}

const togglePacket = (packet: StuckPacket) => {
  const newSelection = isSelected(packet)
    ? props.selected.filter(p => p.id !== packet.id)
    : [...props.selected, packet]
  
  emit('update:selected', newSelection)
}

const selectAll = () => {
  emit('update:selected', [...props.stuckPackets])
}

const selectNone = () => {
  emit('update:selected', [])
}

const formatAddress = (address: string): string => {
  if (address.length <= 20) return address
  return `${address.slice(0, 10)}...${address.slice(-8)}`
}

const formatAmount = (amount: string, denom: string): string => {
  return clearingService.formatAmount(amount, denom)
}

const formatAge = (seconds: number): string => {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

const formatTotalValue = (): string => {
  if (props.selected.length === 0) return '0'
  
  // Group by denom
  const byDenom = props.selected.reduce((acc, packet) => {
    if (!acc[packet.denom]) {
      acc[packet.denom] = BigInt(0)
    }
    acc[packet.denom] += BigInt(packet.amount)
    return acc
  }, {} as Record<string, bigint>)
  
  // Format each denom
  const formatted = Object.entries(byDenom).map(([denom, amount]) => {
    return clearingService.formatAmount(amount.toString(), denom)
  })
  
  // Return first one for simplicity (could show all)
  return formatted[0] || '0'
}
</script>