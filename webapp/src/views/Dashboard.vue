<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900">IBC Relay Dashboard</h1>
    
    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Total Chains</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.chains }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Active Relayers</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.relayers }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">24h Packets</h3>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.packets }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <h3 class="text-sm font-medium text-gray-500">Success Rate</h3>
        <p class="mt-2 text-3xl font-semibold text-green-600">{{ stats.successRate }}%</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Activity -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Recent Activity</h2>
        <div class="space-y-3">
          <div v-for="i in 5" :key="i" class="flex items-center justify-between py-2 border-b">
            <div>
              <p class="text-sm font-medium">Osmosis → Cosmos Hub</p>
              <p class="text-xs text-gray-500">channel-0 → channel-141</p>
            </div>
            <span class="text-xs text-green-600">Success</span>
          </div>
        </div>
      </div>

      <!-- Top Relayers -->
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-medium mb-4">Top Relayers</h2>
        <div class="space-y-3">
          <div v-for="i in 5" :key="i" class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium">osmo1abc...xyz</p>
              <p class="text-xs text-gray-500">Hermes 1.13.1</p>
            </div>
            <div class="text-right">
              <p class="text-sm font-semibold">1,234</p>
              <p class="text-xs text-gray-500">packets</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="bg-blue-50 rounded-lg p-6">
      <h2 class="text-lg font-medium mb-4">Quick Actions</h2>
      <div class="flex flex-wrap gap-4">
        <router-link
          to="/monitoring"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
        >
          View Full Monitoring
        </router-link>
        <router-link
          to="/packet-clearing"
          class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
        >
          Clear Stuck Packets
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const stats = ref({
  chains: 3,
  relayers: 12,
  packets: '15.2K',
  successRate: 87.3
})
</script>