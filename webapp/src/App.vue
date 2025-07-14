<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Navigation -->
    <nav class="bg-white shadow-sm border-b">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <h1 class="text-xl font-bold text-gray-900">Relayooor</h1>
            </div>
            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <router-link
                v-for="item in navigation"
                :key="item.name"
                :to="item.href"
                :class="[
                  $route.path === item.href
                    ? 'border-primary-500 text-gray-900'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                  'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium'
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
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { Home, Activity, Link, Users, PackageX, BarChart3, Settings } from 'lucide-vue-next'
import ConnectionStatus from '@/components/ConnectionStatus.vue'
import WalletConnect from '@/components/WalletConnect.vue'

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'Monitoring', href: '/monitoring', icon: Activity },
  { name: 'Channels', href: '/channels', icon: Link },
  { name: 'Relayers', href: '/relayers', icon: Users },
  { name: 'Packet Clearing', href: '/packet-clearing', icon: PackageX },
  { name: 'Analytics', href: '/analytics', icon: BarChart3 },
  { name: 'Settings', href: '/settings', icon: Settings }
]
</script>