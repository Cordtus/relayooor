<template>
  <div class="dark-mode min-h-screen bg-surface">
    <!-- Navigation -->
    <nav class="bg-surface-secondary shadow-lg border-b border-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <h1 class="text-xl font-bold text-primary">Relayooor</h1>
            </div>
            <div class="hidden sm:ml-8 sm:flex sm:space-x-2">
              <router-link
                v-for="item in navigation"
                :key="item.name"
                :to="item.href"
                :class="[
                  $route.path === item.href
                    ? 'bg-surface-tertiary text-primary border-primary'
                    : 'text-content-secondary hover:bg-surface-tertiary hover:text-content',
                  'inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200'
                ]"
              >
                <component :is="item.icon" class="w-4 h-4 mr-2" />
                {{ item.name }}
              </router-link>
            </div>
          </div>
          <div class="flex items-center space-x-4">
            <ConnectionStatus />
            <WalletConnect />
          </div>
        </div>
      </div>
    </nav>

    <!-- Main content -->
    <main class="max-w-7xl mx-auto">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { Home, Activity, PackageX, BarChart3, Settings } from 'lucide-vue-next'
import ConnectionStatus from '@/components/ConnectionStatus.vue'
import WalletConnect from '@/components/WalletConnect.vue'

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'IBC Monitoring', href: '/monitoring', icon: Activity },
  { name: 'Packet Clearing', href: '/packet-clearing', icon: PackageX },
  { name: 'Analytics', href: '/analytics', icon: BarChart3 },
  { name: 'Settings', href: '/settings', icon: Settings }
]
</script>