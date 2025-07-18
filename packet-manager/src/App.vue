<template>
  <div class="min-h-screen bg-gray-100">
    <div class="container mx-auto px-4 py-8">
      <h1 class="text-3xl font-bold text-gray-800 mb-8">IBC Packet Manager</h1>
      
      <!-- Chain and Channel Selection -->
      <div class="bg-white rounded-lg shadow p-6 mb-8">
        <h2 class="text-xl font-semibold mb-4">Select Chain and Channel</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Chain</label>
            <select 
              v-model="selectedChain" 
              @change="onChainChange"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Select a chain</option>
              <option v-for="chain in chains" :key="chain.id" :value="chain.id">
                {{ chain.name || chain.id }}
              </option>
            </select>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Channel</label>
            <select 
              v-model="selectedChannel"
              :disabled="!selectedChain"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
            >
              <option value="">Select a channel</option>
              <option v-for="channel in channels" :key="channel.id" :value="channel.id">
                {{ channel.id }} ({{ channel.state }})
              </option>
            </select>
          </div>
        </div>
        
        <button 
          @click="queryPackets"
          :disabled="!selectedChain || !selectedChannel || loading"
          class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {{ loading ? 'Loading...' : 'Query Packets' }}
        </button>
      </div>

      <!-- Data Sources Comparison -->
      <div v-if="showResults" class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
        <!-- Chainpulse Data -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-semibold mb-4 text-orange-600">Chainpulse Data</h3>
          <div v-if="chainpulsePackets.length > 0">
            <div class="mb-4 text-sm text-gray-600">
              Found {{ chainpulsePackets.length }} stuck packets
            </div>
            <div class="space-y-2 max-h-96 overflow-y-auto">
              <div 
                v-for="packet in chainpulsePackets" 
                :key="packet.sequence"
                class="border border-gray-200 rounded p-3"
              >
                <div class="font-mono text-sm">Sequence: {{ packet.sequence }}</div>
                <div class="text-xs text-gray-600">
                  <div>Source: {{ packet.src_port }}/{{ packet.src_channel }}</div>
                  <div>Dest: {{ packet.dst_port }}/{{ packet.dst_channel }}</div>
                  <div v-if="packet.timeout_height">Timeout Height: {{ packet.timeout_height }}</div>
                  <div v-if="packet.timeout_timestamp">Timeout: {{ formatTimestamp(packet.timeout_timestamp) }}</div>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="text-gray-500">No stuck packets found</div>
        </div>

        <!-- Hermes Metrics -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-semibold mb-4 text-blue-600">Hermes Metrics</h3>
          <div v-if="hermesMetrics.length > 0">
            <div class="mb-4 text-sm text-gray-600">
              Pending packets by channel
            </div>
            <div class="space-y-2">
              <div 
                v-for="metric in hermesMetrics" 
                :key="`${metric.chain_id}-${metric.channel_id}`"
                class="border border-gray-200 rounded p-3"
              >
                <div class="font-semibold">{{ metric.channel_id }}</div>
                <div class="text-sm text-gray-600">
                  Chain: {{ metric.chain_id }}, Port: {{ metric.port_id }}
                </div>
                <div class="text-lg font-bold text-orange-600">
                  {{ metric.pending_count }} pending packets
                </div>
              </div>
            </div>
          </div>
          <div v-else class="text-gray-500">No metrics available</div>
        </div>
      </div>

      <!-- Packet List and Actions -->
      <div v-if="showResults && chainpulsePackets.length > 0" class="bg-white rounded-lg shadow p-6">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-semibold">Packet Actions</h3>
          <button 
            @click="clearAllPackets"
            :disabled="clearing || chainpulsePackets.length === 0"
            class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:bg-gray-400"
          >
            Clear All Packets ({{ chainpulsePackets.length }})
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Sequence
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Source
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Destination
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Action
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="packet in chainpulsePackets" :key="packet.sequence">
                <td class="px-6 py-4 whitespace-nowrap font-mono text-sm">
                  {{ packet.sequence }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  {{ packet.src_port }}/{{ packet.src_channel }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  {{ packet.dst_port }}/{{ packet.dst_channel }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
                    Stuck
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  <button 
                    @click="clearPacket(packet)"
                    :disabled="clearing"
                    class="text-blue-600 hover:text-blue-900 disabled:text-gray-400"
                  >
                    Clear
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Status Messages -->
      <div v-if="statusMessage" class="mt-4">
        <div 
          :class="[
            'p-4 rounded-md',
            statusType === 'success' ? 'bg-green-100 text-green-800' : 
            statusType === 'error' ? 'bg-red-100 text-red-800' : 
            'bg-blue-100 text-blue-800'
          ]"
        >
          {{ statusMessage }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { packetService } from './services/api'

// State
const chains = ref([])
const channels = ref([])
const selectedChain = ref('')
const selectedChannel = ref('')
const chainpulsePackets = ref([])
const hermesMetrics = ref([])
const loading = ref(false)
const clearing = ref(false)
const showResults = ref(false)
const statusMessage = ref('')
const statusType = ref('info')

// Methods
const loadChains = async () => {
  try {
    chains.value = await packetService.getChains()
  } catch (error) {
    console.error('Failed to load chains:', error)
    showStatus('Failed to load chains', 'error')
  }
}

const onChainChange = async () => {
  selectedChannel.value = ''
  channels.value = []
  
  if (selectedChain.value) {
    try {
      channels.value = await packetService.getChannels(selectedChain.value)
    } catch (error) {
      console.error('Failed to load channels:', error)
      showStatus('Failed to load channels', 'error')
    }
  }
}

const queryPackets = async () => {
  loading.value = true
  showResults.value = false
  chainpulsePackets.value = []
  hermesMetrics.value = []
  
  try {
    // Get stuck packets from Chainpulse
    const stuckPackets = await packetService.getStuckPackets(selectedChain.value, selectedChannel.value)
    chainpulsePackets.value = stuckPackets.packets || []
    
    // Get Hermes metrics
    const metricsText = await packetService.getHermesMetrics()
    if (metricsText) {
      hermesMetrics.value = packetService.parseHermesMetrics(metricsText, selectedChain.value, selectedChannel.value)
    }
    
    showResults.value = true
    showStatus(`Found ${chainpulsePackets.value.length} stuck packets`, 'success')
  } catch (error) {
    console.error('Failed to query packets:', error)
    showStatus('Failed to query packets', 'error')
  } finally {
    loading.value = false
  }
}

const clearPacket = async (packet) => {
  clearing.value = true
  
  try {
    await packetService.clearPacket(
      selectedChain.value,
      selectedChannel.value,
      packet.src_port || 'transfer',
      packet.sequence
    )
    
    showStatus(`Successfully cleared packet ${packet.sequence}`, 'success')
    
    // Refresh the packet list
    setTimeout(() => queryPackets(), 2000)
  } catch (error) {
    console.error('Failed to clear packet:', error)
    showStatus(`Failed to clear packet ${packet.sequence}`, 'error')
  } finally {
    clearing.value = false
  }
}

const clearAllPackets = async () => {
  if (!confirm(`Are you sure you want to clear all ${chainpulsePackets.value.length} packets?`)) {
    return
  }
  
  clearing.value = true
  
  try {
    const sequences = chainpulsePackets.value.map(p => p.sequence)
    await packetService.clearPackets(
      selectedChain.value,
      selectedChannel.value,
      'transfer',
      sequences
    )
    
    showStatus(`Successfully cleared ${sequences.length} packets`, 'success')
    
    // Refresh the packet list
    setTimeout(() => queryPackets(), 2000)
  } catch (error) {
    console.error('Failed to clear packets:', error)
    showStatus('Failed to clear packets', 'error')
  } finally {
    clearing.value = false
  }
}

const showStatus = (message, type = 'info') => {
  statusMessage.value = message
  statusType.value = type
  
  setTimeout(() => {
    statusMessage.value = ''
  }, 5000)
}

const formatTimestamp = (timestamp) => {
  if (!timestamp) return 'N/A'
  const date = new Date(parseInt(timestamp) / 1000000) // Convert from nanoseconds
  return date.toLocaleString()
}

// Lifecycle
onMounted(() => {
  loadChains()
})
</script>