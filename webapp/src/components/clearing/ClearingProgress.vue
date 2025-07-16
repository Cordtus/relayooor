<template>
  <div class="clearing-progress">
    <div class="flex flex-col items-center">
      <!-- Animation -->
      <div class="relative w-32 h-32 mb-6">
        <div class="absolute inset-0 border-4 border-gray-200 rounded-full"></div>
        <div 
          class="absolute inset-0 border-4 border-blue-500 rounded-full animate-spin"
          style="border-top-color: transparent; animation-duration: 2s;"
        ></div>
        <div class="absolute inset-0 flex items-center justify-center">
          <PackageIcon class="h-12 w-12 text-blue-500" />
        </div>
      </div>

      <!-- Status text -->
      <h3 class="text-lg font-medium text-gray-900 mb-2">
        {{ statusText }}
      </h3>
      <p class="text-sm text-gray-600 text-center max-w-md">
        {{ statusDescription }}
      </p>

      <!-- Progress details -->
      <div v-if="status?.execution" class="mt-6 w-full max-w-sm">
        <div class="bg-gray-50 rounded-lg p-4 space-y-3">
          <!-- Packets progress -->
          <div v-if="totalPackets > 0">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-600">Packets Cleared</span>
              <span class="font-medium">
                {{ status.execution.packetsCleared || 0 }} / {{ totalPackets }}
              </span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div 
                class="bg-blue-500 h-2 rounded-full transition-all duration-500"
                :style="{ width: progressPercentage + '%' }"
              ></div>
            </div>
          </div>

          <!-- Failed packets -->
          <div v-if="status.execution.packetsFailed" class="flex justify-between text-sm">
            <span class="text-gray-600">Failed Packets</span>
            <span class="text-red-600 font-medium">{{ status.execution.packetsFailed }}</span>
          </div>

          <!-- Duration -->
          <div v-if="status.execution.startedAt" class="flex justify-between text-sm">
            <span class="text-gray-600">Duration</span>
            <span class="font-medium">{{ duration }}</span>
          </div>
        </div>
      </div>

      <!-- Error state -->
      <div v-if="status?.execution?.error" class="mt-6 w-full max-w-sm">
        <div class="bg-red-50 border border-red-200 rounded-lg p-4">
          <div class="flex">
            <AlertCircleIcon class="h-5 w-5 text-red-400 flex-shrink-0" />
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800">
                Clearing Failed
              </h3>
              <p class="text-sm text-red-700 mt-1">
                {{ status.execution.error }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Transaction hashes -->
      <div v-if="status?.execution?.txHashes?.length" class="mt-6 w-full max-w-sm">
        <p class="text-sm text-gray-600 mb-2">Transaction{{ status.execution.txHashes.length > 1 ? 's' : '' }}:</p>
        <div class="space-y-1">
          <a 
            v-for="(hash, index) in status.execution.txHashes"
            :key="hash"
            :href="getExplorerUrl(hash)"
            target="_blank"
            rel="noopener noreferrer"
            class="block text-sm text-blue-600 hover:underline font-mono truncate"
          >
            {{ hash }}
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Package as PackageIcon, AlertCircle as AlertCircleIcon } from 'lucide-vue-next'
import { clearingService, type ClearingStatus } from '@/services/clearing'

const props = defineProps<{
  token: string | undefined
  status: ClearingStatus | null
}>()

const emit = defineEmits<{
  complete: []
}>()

const duration = ref('0s')
const totalPackets = ref(10) // Would get from parent or API

let durationInterval: ReturnType<typeof setInterval> | null = null

const statusText = computed(() => {
  if (!props.status) return 'Initializing...'
  
  switch (props.status.status) {
    case 'paid':
      return 'Payment Verified'
    case 'executing':
      return 'Clearing Packets...'
    case 'completed':
      return 'Clearing Complete!'
    case 'failed':
      return 'Clearing Failed'
    default:
      return 'Processing...'
  }
})

const statusDescription = computed(() => {
  if (!props.status) return 'Setting up clearing operation'
  
  switch (props.status.status) {
    case 'paid':
      return 'Starting packet clearing process'
    case 'executing':
      return 'Our relayer is clearing your stuck packets'
    case 'completed':
      return 'All packets have been successfully cleared'
    case 'failed':
      return 'Some packets could not be cleared'
    default:
      return 'Please wait while we process your request'
  }
})

const progressPercentage = computed(() => {
  if (!props.status?.execution || totalPackets.value === 0) return 0
  
  const cleared = props.status.execution.packetsCleared || 0
  return Math.round((cleared / totalPackets.value) * 100)
})

const getExplorerUrl = (txHash: string): string => {
  // Default to Osmosis explorer
  return `https://www.mintscan.io/osmosis/tx/${txHash}`
}

const updateDuration = () => {
  if (!props.status?.execution?.startedAt) return
  
  const start = new Date(props.status.execution.startedAt).getTime()
  const now = Date.now()
  const seconds = Math.floor((now - start) / 1000)
  
  if (seconds < 60) {
    duration.value = `${seconds}s`
  } else {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    duration.value = `${minutes}m ${remainingSeconds}s`
  }
}

// Watch for completion
watch(() => props.status?.status, (newStatus) => {
  if (newStatus === 'completed' || newStatus === 'failed') {
    emit('complete')
    if (durationInterval) {
      clearInterval(durationInterval)
    }
  }
})

// Poll for updates if WebSocket fails
onMounted(async () => {
  if (props.token) {
    // Start duration timer
    durationInterval = setInterval(updateDuration, 1000)
    
    // Poll for status as fallback
    try {
      const finalStatus = await clearingService.pollForCompletion(props.token)
      // Status will be updated via props from parent
    } catch (error) {
      console.error('Polling failed:', error)
    }
  }
})

onUnmounted(() => {
  if (durationInterval) {
    clearInterval(durationInterval)
  }
})
</script>