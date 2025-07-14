import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useConnectionStore = defineStore('connection', () => {
  const isConnected = ref(false)
  const connectionStatus = ref<'disconnected' | 'connecting' | 'connected'>('disconnected')
  const lastError = ref<string | null>(null)
  const reconnectAttempts = ref(0)

  function setConnected(status: boolean) {
    isConnected.value = status
    connectionStatus.value = status ? 'connected' : 'disconnected'
    if (status) {
      lastError.value = null
      reconnectAttempts.value = 0
    }
  }

  function setConnecting() {
    connectionStatus.value = 'connecting'
  }

  function setError(error: string) {
    lastError.value = error
    isConnected.value = false
    connectionStatus.value = 'disconnected'
  }

  function incrementReconnectAttempts() {
    reconnectAttempts.value++
  }

  function resetReconnectAttempts() {
    reconnectAttempts.value = 0
  }

  return {
    isConnected,
    connectionStatus,
    lastError,
    reconnectAttempts,
    setConnected,
    setConnecting,
    setError,
    incrementReconnectAttempts,
    resetReconnectAttempts
  }
})