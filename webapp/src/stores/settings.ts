import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Settings {
  refreshInterval: number
  stuckThreshold: number
  notifications: boolean
  cacheTTL: number
  maxPacketsPerClear: number
  devMode: boolean
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Settings>({
    refreshInterval: 30,
    stuckThreshold: 60,
    notifications: false,
    cacheTTL: 300,
    maxPacketsPerClear: 20,
    devMode: false
  })

  // Load settings from localStorage on initialization
  const stored = localStorage.getItem('relayooor_settings')
  if (stored) {
    try {
      settings.value = { ...settings.value, ...JSON.parse(stored) }
    } catch (e) {
      console.error('Failed to load settings:', e)
    }
  }

  function updateSettings(newSettings: Partial<Settings>) {
    settings.value = { ...settings.value, ...newSettings }
    localStorage.setItem('relayooor_settings', JSON.stringify(settings.value))
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
    localStorage.setItem('relayooor_settings', JSON.stringify(settings.value))
  }

  return {
    settings,
    updateSettings,
    resetSettings
  }
})