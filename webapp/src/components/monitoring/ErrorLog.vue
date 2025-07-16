<template>
  <div class="bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
      <h3 class="text-lg font-medium text-gray-900">Error Log</h3>
      <div class="flex items-center gap-2">
        <button
          @click="filterType = 'all'"
          :class="[
            'px-3 py-1 text-xs font-medium rounded-md',
            filterType === 'all' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          ]"
        >
          All ({{ totalErrors }})
        </button>
        <button
          @click="filterType = 'critical'"
          :class="[
            'px-3 py-1 text-xs font-medium rounded-md',
            filterType === 'critical' ? 'bg-red-600 text-white' : 'bg-red-100 text-red-700 hover:bg-red-200'
          ]"
        >
          Critical ({{ criticalCount }})
        </button>
        <button
          @click="filterType = 'warning'"
          :class="[
            'px-3 py-1 text-xs font-medium rounded-md',
            filterType === 'warning' ? 'bg-yellow-600 text-white' : 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200'
          ]"
        >
          Warning ({{ warningCount }})
        </button>
      </div>
    </div>
    
    <div class="p-6">
      <!-- Error List -->
      <div v-if="filteredErrors.length > 0" class="space-y-3">
        <div
          v-for="error in paginatedErrors"
          :key="error.id"
          class="border rounded-lg p-4"
          :class="getErrorClass(error.severity)"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <component :is="getErrorIcon(error.type)" class="h-4 w-4" :class="getErrorIconClass(error.severity)" />
                <span class="text-sm font-medium text-gray-900">{{ error.message }}</span>
                <span :class="[
                  'px-2 py-0.5 text-xs font-medium rounded-full',
                  error.severity === 'critical' ? 'bg-red-100 text-red-800' :
                  error.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-blue-100 text-blue-800'
                ]">
                  {{ error.severity }}
                </span>
              </div>
              
              <div class="mt-2 grid grid-cols-2 gap-4 text-xs text-gray-500">
                <div>
                  <span class="font-medium">Source:</span> {{ error.source }}
                </div>
                <div>
                  <span class="font-medium">Time:</span> {{ formatTime(error.timestamp) }}
                </div>
                <div v-if="error.context.channel">
                  <span class="font-medium">Channel:</span> {{ error.context.channel }}
                </div>
                <div v-if="error.context.relayer">
                  <span class="font-medium">Relayer:</span> {{ formatAddress(error.context.relayer) }}
                </div>
              </div>
              
              <!-- Error Details -->
              <div v-if="error.details" class="mt-2">
                <button
                  @click="toggleDetails(error.id)"
                  class="text-xs text-blue-600 hover:text-blue-800 font-medium"
                >
                  {{ expandedErrors.has(error.id) ? 'Hide' : 'Show' }} Details
                </button>
                <div v-if="expandedErrors.has(error.id)" class="mt-2 p-2 bg-gray-50 rounded text-xs font-mono text-gray-700">
                  {{ error.details }}
                </div>
              </div>
              
              <!-- Error Count -->
              <div v-if="error.count > 1" class="mt-2">
                <span class="text-xs text-gray-500">
                  This error occurred {{ error.count }} times in the last hour
                </span>
              </div>
            </div>
            
            <!-- Actions -->
            <div class="ml-4 flex flex-col gap-1">
              <button
                v-if="error.context.txHash"
                @click="viewTransaction(error.context.txHash)"
                class="px-2 py-1 text-xs font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded"
              >
                View Tx
              </button>
              <button
                @click="dismissError(error.id)"
                class="px-2 py-1 text-xs font-medium text-gray-700 bg-gray-50 hover:bg-gray-100 rounded"
              >
                Dismiss
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- No Errors -->
      <div v-else class="text-center py-8">
        <CheckCircle class="mx-auto h-12 w-12 text-green-500" />
        <h3 class="mt-2 text-sm font-medium text-gray-900">No Errors Found</h3>
        <p class="mt-1 text-sm text-gray-500">
          {{ filterType === 'all' ? 'No errors in the system' : `No ${filterType} errors` }}
        </p>
      </div>
      
      <!-- Pagination -->
      <div v-if="totalPages > 1" class="mt-4 flex items-center justify-between">
        <p class="text-sm text-gray-700">
          Showing {{ (currentPage - 1) * pageSize + 1 }} to {{ Math.min(currentPage * pageSize, filteredErrors.length) }} of {{ filteredErrors.length }} errors
        </p>
        <div class="flex gap-1">
          <button
            @click="currentPage--"
            :disabled="currentPage === 1"
            class="px-3 py-1 text-sm font-medium rounded border"
            :class="currentPage === 1 ? 'bg-gray-50 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'"
          >
            Previous
          </button>
          <button
            @click="currentPage++"
            :disabled="currentPage === totalPages"
            class="px-3 py-1 text-sm font-medium rounded border"
            :class="currentPage === totalPages ? 'bg-gray-50 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'"
          >
            Next
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { CheckCircle, AlertTriangle, AlertCircle, XCircle, WifiOff, Clock, Database } from 'lucide-vue-next'

interface ErrorContext {
  channel?: string
  relayer?: string
  txHash?: string
  chain?: string
}

interface Error {
  id: string
  type: 'connection' | 'timeout' | 'validation' | 'transaction' | 'unknown'
  severity: 'critical' | 'warning' | 'info'
  message: string
  source: string
  timestamp: Date
  context: ErrorContext
  details?: string
  count: number
}

const props = defineProps<{
  errors?: Error[]
}>()

// State
const filterType = ref<'all' | 'critical' | 'warning'>('all')
const expandedErrors = ref(new Set<string>())
const dismissedErrors = ref(new Set<string>())
const currentPage = ref(1)
const pageSize = 10

// Use real errors if provided, otherwise show empty state
const allErrors = computed(() => {
  const errors = props.errors || []
  return errors.filter(e => !dismissedErrors.value.has(e.id))
})

// Computed
const totalErrors = computed(() => allErrors.value.length)
const criticalCount = computed(() => allErrors.value.filter(e => e.severity === 'critical').length)
const warningCount = computed(() => allErrors.value.filter(e => e.severity === 'warning').length)

const filteredErrors = computed(() => {
  if (filterType.value === 'all') return allErrors.value
  return allErrors.value.filter(e => e.severity === filterType.value)
})

const totalPages = computed(() => Math.ceil(filteredErrors.value.length / pageSize))

const paginatedErrors = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return filteredErrors.value.slice(start, end)
})

// Methods
function getErrorClass(severity: string): string {
  switch (severity) {
    case 'critical':
      return 'border-red-300 bg-red-50'
    case 'warning':
      return 'border-yellow-300 bg-yellow-50'
    case 'info':
      return 'border-blue-300 bg-blue-50'
    default:
      return 'border-gray-300'
  }
}

function getErrorIcon(type: string) {
  switch (type) {
    case 'connection':
      return WifiOff
    case 'timeout':
      return Clock
    case 'transaction':
      return XCircle
    case 'validation':
      return AlertCircle
    default:
      return AlertTriangle
  }
}

function getErrorIconClass(severity: string): string {
  switch (severity) {
    case 'critical':
      return 'text-red-600'
    case 'warning':
      return 'text-yellow-600'
    case 'info':
      return 'text-blue-600'
    default:
      return 'text-gray-600'
  }
}

function formatTime(date: Date): string {
  const now = Date.now()
  const diff = now - date.getTime()
  const minutes = Math.floor(diff / 60000)
  
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (minutes < 1440) return `${Math.floor(minutes / 60)}h ago`
  return date.toLocaleDateString()
}

function formatAddress(address: string): string {
  return `${address.slice(0, 8)}...${address.slice(-6)}`
}

function toggleDetails(id: string) {
  if (expandedErrors.value.has(id)) {
    expandedErrors.value.delete(id)
  } else {
    expandedErrors.value.add(id)
  }
}

function viewTransaction(txHash: string) {
  // Open in block explorer
  console.log('View transaction:', txHash)
}

function dismissError(id: string) {
  dismissedErrors.value.add(id)
}
</script>