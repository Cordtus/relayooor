<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-3xl font-bold mb-8 text-gray-800">Settings</h1>
    
    <div class="grid grid-cols-1 gap-6">
      <!-- Monitoring Configuration -->
      <Card>
        <template #header>
          <h3 class="text-lg font-semibold">Monitoring Configuration</h3>
        </template>
        <p class="text-gray-600 mb-4">Configure monitoring preferences and notification settings.</p>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Refresh Interval (seconds)
            </label>
            <input
              v-model.number="settings.refreshInterval"
              type="number"
              min="5"
              max="300"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              @change="saveSettings"
            >
            <p class="text-sm text-gray-500 mt-1">How often to refresh monitoring data (5-300 seconds)</p>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Stuck Packet Threshold (minutes)
            </label>
            <input
              v-model.number="settings.stuckThreshold"
              type="number"
              min="15"
              max="1440"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              @change="saveSettings"
            >
            <p class="text-sm text-gray-500 mt-1">Minutes before a packet is considered stuck (15-1440)</p>
          </div>
          
          <div>
            <label class="flex items-center space-x-2">
              <input
                v-model="settings.notifications"
                type="checkbox"
                class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                @change="saveSettings"
              >
              <span class="text-sm font-medium text-gray-700">Enable browser notifications</span>
            </label>
          </div>
        </div>
      </Card>

      <!-- Connected Services -->
      <Card>
        <template #header>
          <h3 class="text-lg font-semibold">Connected Services</h3>
        </template>
        <p class="text-gray-600 mb-4">Status of connected monitoring and relayer services.</p>
        
        <div class="space-y-3">
          <div class="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
            <div>
              <span class="text-sm font-medium text-gray-700">Chainpulse</span>
              <p class="text-xs text-gray-500">{{ chainpulseUrl }}</p>
            </div>
            <span 
              class="text-sm font-medium"
              :class="services.chainpulse ? 'text-green-600' : 'text-red-600'"
            >
              {{ services.chainpulse ? 'Connected' : 'Disconnected' }}
            </span>
          </div>
          
          <div class="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
            <div>
              <span class="text-sm font-medium text-gray-700">API Backend</span>
              <p class="text-xs text-gray-500">{{ apiUrl }}</p>
            </div>
            <span 
              class="text-sm font-medium"
              :class="services.api ? 'text-green-600' : 'text-red-600'"
            >
              {{ services.api ? 'Connected' : 'Disconnected' }}
            </span>
          </div>
          
          <div class="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
            <div>
              <span class="text-sm font-medium text-gray-700">WebSocket</span>
              <p class="text-xs text-gray-500">Real-time updates</p>
            </div>
            <span 
              class="text-sm font-medium"
              :class="services.websocket ? 'text-green-600' : 'text-red-600'"
            >
              {{ services.websocket ? 'Connected' : 'Disconnected' }}
            </span>
          </div>
        </div>
        
        <button
          @click="testConnections"
          class="mt-4 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
          :disabled="testing"
        >
          {{ testing ? 'Testing...' : 'Test Connections' }}
        </button>
      </Card>

      <!-- Chain Configuration -->
      <Card>
        <template #header>
          <h3 class="text-lg font-semibold">Chain Configuration</h3>
        </template>
        <p class="text-gray-600 mb-4">Configure chain-specific settings and RPC endpoints.</p>
        
        <div class="space-y-4">
          <div v-for="chain in chains" :key="chain.id" class="p-3 bg-gray-50 rounded-lg">
            <div class="flex justify-between items-start">
              <div>
                <h4 class="font-medium text-gray-900">{{ chain.name }}</h4>
                <p class="text-sm text-gray-500">{{ chain.id }}</p>
              </div>
              <button
                @click="editChain(chain)"
                class="text-sm text-primary-600 hover:text-primary-700"
              >
                Edit
              </button>
            </div>
            <div class="mt-2 text-xs text-gray-600">
              <p>Fee: {{ chain.clearingFee }} {{ chain.denom }}</p>
            </div>
          </div>
        </div>
      </Card>

      <!-- Advanced Settings -->
      <Card>
        <template #header>
          <h3 class="text-lg font-semibold">Advanced Settings</h3>
        </template>
        <p class="text-gray-600 mb-4">Advanced configuration options.</p>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Cache TTL (seconds)
            </label>
            <input
              v-model.number="settings.cacheTTL"
              type="number"
              min="60"
              max="3600"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              @change="saveSettings"
            >
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Max Packets Per Clear
            </label>
            <input
              v-model.number="settings.maxPacketsPerClear"
              type="number"
              min="1"
              max="100"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
              @change="saveSettings"
            >
          </div>
          
          <div>
            <label class="flex items-center space-x-2">
              <input
                v-model="settings.devMode"
                type="checkbox"
                class="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                @change="saveSettings"
              >
              <span class="text-sm font-medium text-gray-700">Developer Mode</span>
            </label>
            <p class="text-sm text-gray-500 mt-1">Shows additional debugging information</p>
          </div>
        </div>
      </Card>

      <!-- Actions -->
      <div class="flex justify-between items-center">
        <button
          @click="resetSettings"
          class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500"
        >
          Reset to Defaults
        </button>
        
        <div class="space-x-3">
          <button
            @click="exportSettings"
            class="px-4 py-2 border border-primary-600 text-primary-600 rounded-md hover:bg-primary-50 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            Export Settings
          </button>
          <button
            @click="importSettings"
            class="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            Import Settings
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useToast } from 'vue-toastification'
import Card from '@/components/Card.vue'
import { useSettingsStore } from '@/stores/settings'
import { useConnectionStore } from '@/stores/connection'
import { apiClient } from '@/services/api'
import { configService } from '@/services/config'
import { REFRESH_INTERVALS, UI_THRESHOLDS, CACHE_DURATIONS, PAGINATION } from '@/config/constants'

const toast = useToast()
const settingsStore = useSettingsStore()
const connectionStore = useConnectionStore()

const settings = ref({
  refreshInterval: 30,
  stuckThreshold: 60,
  notifications: false,
  cacheTTL: 300,
  maxPacketsPerClear: 20,
  devMode: false
})

const services = ref({
  chainpulse: false,
  api: false,
  websocket: false
})

// Generate chains from config service
const chains = ref<Array<{
  id: string
  name: string
  clearingFee: string
  denom: string
}>>([])

// Load chains from config service
onMounted(async () => {
  const allChains = await configService.getAllChains()
  
  // Get clearing fees and denoms from API if available
  try {
    const registry = await configService.getRegistry()
    const clearingFees = registry.clearingFees || {}
    const denomMap = registry.denoms || {
      'osmosis-1': 'uosmo',
      'cosmoshub-4': 'uatom',
      'neutron-1': 'untrn',
      'noble-1': 'uusdc'
    }
    
    chains.value = allChains.map(chain => ({
      id: chain.chain_id,
      name: chain.chain_name,
      clearingFee: clearingFees[chain.chain_id] || '0.1',
      denom: denomMap[chain.chain_id] || 'uatom'
    }))
  } catch (error) {
    // Fallback to basic chain info
    chains.value = allChains.map(chain => ({
      id: chain.chain_id,
      name: chain.chain_name,
      clearingFee: '0.1',
      denom: getDenomForChain(chain.chain_id)
    }))
  }
})

// Helper function to get denom based on chain ID
function getDenomForChain(chainId: string): string {
  const denomMap: Record<string, string> = {
    'osmosis-1': 'uosmo',
    'cosmoshub-4': 'uatom',
    'neutron-1': 'untrn',
    'noble-1': 'uusdc',
    'akash-1': 'uakt',
    'stargaze-1': 'ustars',
    'juno-1': 'ujuno',
    'stride-1': 'ustrd',
    'axelar-1': 'uaxl'
  }
  return denomMap[chainId] || 'uatom'
}

const testing = ref(false)

const apiUrl = computed(() => import.meta.env.VITE_API_URL || 'http://localhost:3000')
const chainpulseUrl = computed(() => import.meta.env.VITE_CHAINPULSE_URL || 'http://localhost:3001')

onMounted(() => {
  loadSettings()
  checkServiceStatus()
})

function loadSettings() {
  const stored = localStorage.getItem('relayooor_settings')
  if (stored) {
    try {
      settings.value = { ...settings.value, ...JSON.parse(stored) }
    } catch (e) {
      console.error('Failed to load settings:', e)
    }
  }
}

function saveSettings() {
  localStorage.setItem('relayooor_settings', JSON.stringify(settings.value))
  settingsStore.updateSettings(settings.value)
  toast.success('Settings saved')
}

function resetSettings() {
  settings.value = {
    refreshInterval: 30,
    stuckThreshold: 60,
    notifications: false,
    cacheTTL: 300,
    maxPacketsPerClear: 20,
    devMode: false
  }
  saveSettings()
}

function exportSettings() {
  const data = {
    settings: settings.value,
    supportedChains: chains.value.map(c => c.id), // Only export chain IDs
    exported: new Date().toISOString()
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `relayooor-settings-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  toast.success('Settings exported')
}

function importSettings() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return
    
    try {
      const text = await file.text()
      const data = JSON.parse(text)
      
      if (data.settings) {
        settings.value = { ...settings.value, ...data.settings }
      }
      if (data.supportedChains) {
        // Map chain IDs to our chain objects
        const allChains = await configService.getAllChains()
        chains.value = data.supportedChains
          .map((chainId: string) => {
            const chain = allChains.find(c => c.chain_id === chainId)
            return chain ? {
              id: chain.chain_id,
              name: chain.chain_name,
              clearingFee: '0.1',
              denom: getDenomForChain(chain.chain_id)
            } : null
          })
          .filter(Boolean)
      }
      
      saveSettings()
      toast.success('Settings imported')
    } catch (e) {
      toast.error('Failed to import settings')
    }
  }
  input.click()
}

async function checkServiceStatus() {
  // Check API
  try {
    await apiClient.get('/health')
    services.value.api = true
  } catch {
    services.value.api = false
  }
  
  // Check WebSocket status from connection store
  services.value.websocket = connectionStore.isConnected
  
  // Check Chainpulse
  try {
    const response = await fetch(`${chainpulseUrl.value}/api/health`)
    services.value.chainpulse = response.ok
  } catch {
    services.value.chainpulse = false
  }
}

async function testConnections() {
  testing.value = true
  await checkServiceStatus()
  testing.value = false
  
  const allConnected = services.value.api && services.value.websocket && services.value.chainpulse
  if (allConnected) {
    toast.success('All services connected')
  } else {
    toast.warning('Some services are not connected')
  }
}

function editChain(chain: any) {
  toast.info(`Chain editing not yet implemented for ${chain.name}`)
}
</script>