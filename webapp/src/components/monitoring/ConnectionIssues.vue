<template>
  <div class="bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-gray-200">
      <h3 class="text-lg font-medium text-gray-900">Connection Issues</h3>
    </div>
    
    <div class="p-6">
      <!-- Active Issues -->
      <div v-if="activeIssues.length > 0" class="space-y-4">
        <div
          v-for="issue in activeIssues"
          :key="issue.chainId"
          class="border rounded-lg p-4"
          :class="getIssueSeverityClass(issue.severity)"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <component :is="getIssueIcon(issue.type)" class="h-5 w-5" :class="getIssueIconClass(issue.severity)" />
                <h4 class="text-sm font-medium text-gray-900">
                  {{ issue.chainName || issue.chainId }}
                </h4>
                <span :class="[
                  'px-2 py-0.5 text-xs font-medium rounded-full',
                  issue.severity === 'critical' ? 'bg-red-100 text-red-800' :
                  issue.severity === 'warning' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-blue-100 text-blue-800'
                ]">
                  {{ issue.severity }}
                </span>
              </div>
              
              <p class="mt-1 text-sm text-gray-600">{{ issue.description }}</p>
              
              <div class="mt-2 flex items-center gap-4 text-xs text-gray-500">
                <span>Started: {{ formatTime(issue.startTime) }}</span>
                <span>Duration: {{ formatDuration(issue.duration) }}</span>
                <span v-if="issue.reconnectAttempts > 0">
                  Reconnect attempts: {{ issue.reconnectAttempts }}
                </span>
              </div>
              
              <!-- Affected Channels -->
              <div v-if="issue.affectedChannels && issue.affectedChannels.length > 0" class="mt-2">
                <p class="text-xs text-gray-500">Affected channels:</p>
                <div class="mt-1 flex flex-wrap gap-1">
                  <span
                    v-for="channel in issue.affectedChannels"
                    :key="channel"
                    class="px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded"
                  >
                    {{ channel }}
                  </span>
                </div>
              </div>
            </div>
            
            <!-- Actions -->
            <div class="ml-4 flex flex-col gap-2">
              <button
                @click="attemptReconnect(issue)"
                class="px-3 py-1 text-xs font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-md"
              >
                Retry Connection
              </button>
              <button
                @click="viewDetails(issue)"
                class="px-3 py-1 text-xs font-medium text-gray-700 bg-gray-50 hover:bg-gray-100 rounded-md"
              >
                View Details
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- No Issues -->
      <div v-else class="text-center py-8">
        <CheckCircle class="mx-auto h-12 w-12 text-green-500" />
        <h3 class="mt-2 text-sm font-medium text-gray-900">All Connections Healthy</h3>
        <p class="mt-1 text-sm text-gray-500">
          All chains are connected and operating normally.
        </p>
      </div>
      
      <!-- Connection History -->
      <div class="mt-6 border-t pt-6">
        <h4 class="text-sm font-medium text-gray-900 mb-3">Recent Connection Events</h4>
        <div class="space-y-2">
          <div
            v-for="event in recentEvents"
            :key="event.id"
            class="flex items-center justify-between text-sm"
          >
            <div class="flex items-center gap-2">
              <div
                class="w-2 h-2 rounded-full"
                :class="[
                  event.type === 'connected' ? 'bg-green-500' :
                  event.type === 'disconnected' ? 'bg-red-500' :
                  event.type === 'reconnecting' ? 'bg-yellow-500 animate-pulse' :
                  'bg-gray-400'
                ]"
              />
              <span class="text-gray-600">{{ event.chainName }}</span>
              <span class="text-gray-500">{{ event.message }}</span>
            </div>
            <span class="text-xs text-gray-400">{{ formatTimeAgo(event.timestamp) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { 
  CheckCircle, 
  WifiOff, 
  AlertTriangle, 
  RefreshCw,
  Clock,
  Server
} from 'lucide-vue-next'

interface Chain {
  chainId: string
  chainName?: string
  status: string
  reconnects?: number
  timeouts?: number
  errors?: number
  lastUpdate?: Date
}

interface ConnectionIssue {
  chainId: string
  chainName?: string
  type: 'disconnected' | 'timeout' | 'high_latency' | 'sync_lag'
  severity: 'critical' | 'warning' | 'info'
  description: string
  startTime: Date
  duration: number // seconds
  reconnectAttempts: number
  affectedChannels?: string[]
}

interface ConnectionEvent {
  id: string
  chainName: string
  type: 'connected' | 'disconnected' | 'reconnecting' | 'error'
  message: string
  timestamp: Date
}

const props = defineProps<{
  chains?: Chain[]
}>()

// Compute active connection issues from chain data
const activeIssues = computed((): ConnectionIssue[] => {
  if (!props.chains) return []
  
  const issues: ConnectionIssue[] = []
  const now = new Date()
  
  props.chains.forEach(chain => {
    // Check for disconnected chains
    if (chain.status === 'disconnected') {
      issues.push({
        chainId: chain.chainId,
        chainName: chain.chainName,
        type: 'disconnected',
        severity: 'critical',
        description: 'Chain is disconnected from the network',
        startTime: chain.lastUpdate || new Date(now.getTime() - 300000), // 5 min ago
        duration: chain.lastUpdate ? Math.floor((now.getTime() - chain.lastUpdate.getTime()) / 1000) : 300,
        reconnectAttempts: chain.reconnects || 0,
        affectedChannels: getAffectedChannels(chain.chainId)
      })
    }
    
    // Check for high timeout rate
    else if (chain.timeouts && chain.timeouts > 5) {
      issues.push({
        chainId: chain.chainId,
        chainName: chain.chainName,
        type: 'timeout',
        severity: 'warning',
        description: `High timeout rate detected (${chain.timeouts} timeouts)`,
        startTime: new Date(now.getTime() - 600000), // 10 min ago
        duration: 600,
        reconnectAttempts: 0,
        affectedChannels: getAffectedChannels(chain.chainId)
      })
    }
    
    // Check for sync lag
    else if (chain.status === 'syncing') {
      issues.push({
        chainId: chain.chainId,
        chainName: chain.chainName,
        type: 'sync_lag',
        severity: 'info',
        description: 'Chain is syncing with the network',
        startTime: chain.lastUpdate || new Date(now.getTime() - 120000), // 2 min ago
        duration: 120,
        reconnectAttempts: 0
      })
    }
  })
  
  return issues
})

// Recent connection events from chain status changes
const recentEvents = computed((): ConnectionEvent[] => {
  // In production, this would come from a store that tracks connection state changes
  // For now, return empty array when no real data is available
  return []
})

function getAffectedChannels(chainId: string): string[] {
  // In production, this should query actual channel data from the API
  // Return empty array when no real data is available
  return []
}

function getIssueSeverityClass(severity: string): string {
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

function getIssueIcon(type: string) {
  switch (type) {
    case 'disconnected':
      return WifiOff
    case 'timeout':
      return Clock
    case 'high_latency':
      return AlertTriangle
    case 'sync_lag':
      return RefreshCw
    default:
      return Server
  }
}

function getIssueIconClass(severity: string): string {
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
  return date.toLocaleTimeString()
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}

function formatTimeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

function attemptReconnect(issue: ConnectionIssue) {
  console.log('Attempting to reconnect:', issue.chainId)
  // In production, this would trigger a reconnection attempt
}

function viewDetails(issue: ConnectionIssue) {
  console.log('Viewing details for:', issue.chainId)
  // In production, this could open a modal or navigate to a details page
}
</script>