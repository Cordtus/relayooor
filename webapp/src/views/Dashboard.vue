<template>
  <div class="space-y-6 p-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold">IBC Relay Dashboard</h1>
      <RefreshRateSelector :lastUpdate="lastUpdateTime" />
    </div>
    
    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card v-for="stat in statsCards" :key="stat.label">
        <template #content>
          <div class="text-center">
            <p class="text-sm text-neutral-400 mb-2">{{ stat.label }}</p>
            <p class="text-3xl font-bold" :class="stat.colorClass">{{ stat.value }}</p>
          </div>
        </template>
      </Card>
    </div>

    <!-- Packet Search -->
    <Card>
      <template #content>
        <PacketSearch @view-packet="viewPacketDetails" />
      </template>
    </Card>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Activity -->
      <Card>
        <template #title>Recent Activity</template>
        <template #content>
          <DataTable 
            v-if="recentActivity.length > 0" 
            :value="recentActivity" 
            :rows="5"
            class="p-datatable-sm"
            stripedRows
          >
            <Column field="route" header="Route">
              <template #body="{ data }">
                <div>
                  <p class="font-medium">{{ getChainNameSync(data.from_chain) }} → {{ getChainNameSync(data.to_chain) }}</p>
                  <p class="text-xs text-neutral-400">{{ data.channel }}</p>
                </div>
              </template>
            </Column>
            <Column field="status" header="Status" class="w-24">
              <template #body="{ data }">
                <Tag 
                  :value="data.status" 
                  :severity="data.status === 'success' ? 'success' : 'danger'"
                  class="text-xs"
                />
              </template>
            </Column>
          </DataTable>
          <div v-else class="text-center py-8 text-neutral-400">
            <i class="pi pi-inbox text-4xl mb-2"></i>
            <p>No recent activity</p>
          </div>
        </template>
      </Card>

      <!-- Top Relayers -->
      <Card>
        <template #title>Top Relayers</template>
        <template #content>
          <DataTable 
            v-if="topRelayers.length > 0" 
            :value="topRelayers" 
            :rows="5"
            class="p-datatable-sm"
            stripedRows
          >
            <Column field="address" header="Address">
              <template #body="{ data }">
                <div>
                  <p class="font-medium font-mono text-sm">{{ formatAddress(data.address) }}</p>
                  <p class="text-xs text-neutral-400">Success Rate: {{ (data.successRate || 0).toFixed(1) }}%</p>
                </div>
              </template>
            </Column>
            <Column field="totalPackets" header="Packets" class="w-24 text-right">
              <template #body="{ data }">
                <div class="text-right">
                  <p class="font-semibold">{{ formatNumber(data.totalPackets || 0) }}</p>
                  <p class="text-xs text-neutral-400">packets</p>
                </div>
              </template>
            </Column>
          </DataTable>
          <div v-else class="text-center py-8 text-neutral-400">
            <i class="pi pi-users text-4xl mb-2"></i>
            <p>No relayer data available</p>
          </div>
        </template>
      </Card>
    </div>

    <!-- Three Column Layout for Additional Stats -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Top 5 Token Routes -->
      <Card>
        <template #title>Top Token Routes</template>
        <template #content>
          <div v-if="topTokenRoutes.length > 0" class="space-y-3">
            <div v-for="(route, index) in topTokenRoutes" :key="index" 
                 class="flex items-center justify-between p-3 bg-neutral-800 rounded-lg">
              <div class="flex-1">
                <p class="font-medium">{{ route.token }}</p>
                <p class="text-xs text-neutral-400">{{ route.srcChain }} → {{ route.dstChain }}</p>
              </div>
              <div class="text-right">
                <p class="font-semibold">{{ formatNumber(route.packetCount) }}</p>
                <p class="text-xs text-neutral-400">packets</p>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-8 text-neutral-400">
            <i class="pi pi-chart-line text-4xl mb-2"></i>
            <p>No token route data</p>
          </div>
        </template>
      </Card>

      <!-- Top 5 Relayers by Memo -->
      <Card>
        <template #title>Top Relayers by Label</template>
        <template #content>
          <div v-if="topRelayersByMemo.length > 0" class="space-y-3">
            <div v-for="relayer in topRelayersByMemo" :key="relayer.memo || relayer.address" 
                 class="flex items-center justify-between p-3 bg-neutral-800 rounded-lg">
              <div class="flex-1">
                <p class="font-medium">{{ relayer.memo || 'Anonymous' }}</p>
                <p class="text-xs text-neutral-400">{{ relayer.addressCount }} address{{ relayer.addressCount > 1 ? 'es' : '' }}</p>
              </div>
              <div class="text-right">
                <p class="font-semibold">{{ formatNumber(relayer.totalPackets) }}</p>
                <p class="text-xs text-neutral-400">{{ relayer.successRate.toFixed(1) }}% success</p>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-8 text-neutral-400">
            <i class="pi pi-tag text-4xl mb-2"></i>
            <p>No relayer data</p>
          </div>
        </template>
      </Card>

      <!-- Top 5 Chains by Timeouts -->
      <Card>
        <template #title>Chains by Packet Timeouts</template>
        <template #content>
          <div v-if="topChainsByTimeouts.length > 0" class="space-y-3">
            <div v-for="chain in topChainsByTimeouts" :key="chain.chainId" 
                 class="flex items-center justify-between p-3 bg-neutral-800 rounded-lg">
              <div class="flex-1">
                <p class="font-medium">{{ chain.chainName || chain.chainId }}</p>
                <p class="text-xs text-neutral-400">{{ formatNumber(chain.totalPackets) }} total packets</p>
              </div>
              <div class="text-right">
                <p class="font-semibold text-orange-400">{{ chain.timeouts }}</p>
                <p class="text-xs text-neutral-400">timeouts</p>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-8 text-neutral-400">
            <i class="pi pi-clock text-4xl mb-2"></i>
            <p>No timeout data</p>
          </div>
        </template>
      </Card>
    </div>

    <!-- Quick Actions -->
    <Card class="bg-primary/5">
      <template #title>Quick Actions</template>
      <template #content>
        <div class="flex flex-wrap gap-4">
          <Button 
            label="View Full Monitoring" 
            icon="pi pi-chart-bar"
            @click="$router.push('/monitoring')"
          />
          <Button 
            label="Clear Stuck Packets" 
            icon="pi pi-sync"
            severity="secondary"
            @click="$router.push('/packet-clearing')"
          />
        </div>
      </template>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watchEffect } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { api, metricsService } from '@/services/api'
import { configService } from '@/services/config'
import { useSettingsStore } from '@/stores/settings'
import { formatNumber, formatAddress } from '@/utils/formatting'
import RefreshRateSelector from '@/components/RefreshRateSelector.vue'
import PacketSearch from '@/components/search/PacketSearch.vue'
import { resolveChannels, type ChannelInfo } from '@/services/channel-resolver'
import { useRouter } from 'vue-router'

// PrimeVue Components
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'

const settingsStore = useSettingsStore()
const lastUpdateTime = ref(new Date())
const router = useRouter()

// Fetch monitoring data
const { data: monitoringData } = useQuery({
  queryKey: ['monitoring-data'],
  queryFn: async () => {
    lastUpdateTime.value = new Date()
    return metricsService.getMonitoringData()
  },
  refetchInterval: () => settingsStore.settings.refreshInterval
})

// Fetch comprehensive metrics
const { data: comprehensiveMetrics } = useQuery({
  queryKey: ['comprehensive-metrics'],
  queryFn: async () => {
    lastUpdateTime.value = new Date()
    return metricsService.getMonitoringMetrics()
  },
  refetchInterval: () => settingsStore.settings.refreshInterval
})

// Fetch stuck packets data
const { data: stuckPacketsData } = useQuery({
  queryKey: ['stuck-packets'],
  queryFn: async () => {
    try {
      const response = await api.get('/packets/stuck')
      return response.data
    } catch (error) {
      return { packets: [] }
    }
  },
  refetchInterval: () => settingsStore.settings.refreshInterval
})

const stats = ref({
  chains: 0,
  relayers: 0,
  packets: 0,
  successRate: 0
})

const statsCards = computed(() => [
  { label: 'Total Chains', value: stats.value.chains, colorClass: 'text-neutral-100' },
  { label: 'Active Relayers', value: stats.value.relayers, colorClass: 'text-neutral-100' },
  { label: '24h Packets', value: formatNumber(stats.value.packets), colorClass: 'text-neutral-100' },
  { label: 'Success Rate', value: `${stats.value.successRate}%`, colorClass: 'text-green-400' }
])

const topRelayers = ref<any[]>([])
const recentActivity = ref<any[]>([])
const topTokenRoutes = ref<any[]>([])
const topRelayersByMemo = ref<any[]>([])
const topChainsByTimeouts = ref<any[]>([])
const resolvedChannels = ref<Map<string, ChannelInfo>>(new Map())

// Update stats when data is loaded
const updateStats = async () => {
  if (comprehensiveMetrics.value) {
    stats.value.chains = comprehensiveMetrics.value.system?.totalChains || 0
    stats.value.relayers = comprehensiveMetrics.value.relayers?.length || 0
    stats.value.packets = comprehensiveMetrics.value.system?.totalPackets || 0
    
    const channels = comprehensiveMetrics.value.channels || []
    if (channels.length > 0) {
      const avgSuccessRate = channels.reduce((acc: number, ch: any) => acc + (ch.successRate || 0), 0) / channels.length
      stats.value.successRate = Math.round(avgSuccessRate * 10) / 10
    }
    
    topRelayers.value = comprehensiveMetrics.value.relayers?.slice(0, 5) || []
    
    if (stuckPacketsData.value?.packets) {
      const tokenRouteMap = new Map()
      stuckPacketsData.value.packets.forEach((packet: any) => {
        const token = extractTokenFromDenom(packet.denom)
        const key = `${token}-${packet.chain_id}-${packet.dst_channel}`
        if (!tokenRouteMap.has(key)) {
          tokenRouteMap.set(key, {
            token,
            srcChain: packet.chain_id,
            dstChain: getDestChainFromChannel(packet.dst_channel),
            packetCount: 0,
            totalValue: 0
          })
        }
        const route = tokenRouteMap.get(key)
        route.packetCount++
        route.totalValue += parseInt(packet.amount) || 0
      })
      topTokenRoutes.value = Array.from(tokenRouteMap.values())
        .sort((a, b) => b.packetCount - a.packetCount)
        .slice(0, 5)
    }
    
    if (comprehensiveMetrics.value.relayers) {
      const memoMap = new Map()
      comprehensiveMetrics.value.relayers.forEach((relayer: any) => {
        const memo = relayer.memo || 'Anonymous'
        if (!memoMap.has(memo)) {
          memoMap.set(memo, {
            memo,
            addresses: new Set(),
            totalPackets: 0,
            effectedPackets: 0,
            addressCount: 0
          })
        }
        const group = memoMap.get(memo)
        group.addresses.add(relayer.address)
        group.totalPackets += relayer.totalPackets
        group.effectedPackets += relayer.effectedPackets
      })
      
      topRelayersByMemo.value = Array.from(memoMap.values())
        .map(group => ({
          ...group,
          addressCount: group.addresses.size,
          successRate: group.totalPackets > 0 ? (group.effectedPackets / group.totalPackets) * 100 : 0
        }))
        .sort((a, b) => b.totalPackets - a.totalPackets)
        .slice(0, 5)
    }
    
    if (comprehensiveMetrics.value.chains) {
      topChainsByTimeouts.value = comprehensiveMetrics.value.chains
        .map((chain: any) => ({
          ...chain,
          timeouts: chain.timeouts || 0
        }))
        .filter(chain => chain.timeouts > 0)
        .sort((a, b) => b.timeouts - a.timeouts)
        .slice(0, 5)
    }
    
    if (comprehensiveMetrics.value.recentPackets && comprehensiveMetrics.value.recentPackets.length > 0) {
      const channelsToResolve = comprehensiveMetrics.value.recentPackets
        .filter((p: any) => (!p.dst_chain || p.dst_chain === 'unknown') && p.src_channel)
        .map((p: any) => ({
          sourceChainId: p.chain_id,
          channelId: p.src_channel,
          portId: 'transfer'
        }))
      
      if (channelsToResolve.length > 0) {
        try {
          const resolved = await resolveChannels(channelsToResolve)
          resolvedChannels.value = resolved
        } catch (error) {
          console.warn('Failed to resolve some channels:', error)
        }
      }
      
      recentActivity.value = comprehensiveMetrics.value.recentPackets.slice(0, 5).map((p: any) => {
        let destChain = p.dst_chain
        
        if (!destChain || destChain === 'unknown') {
          const resolvedKey = `${p.chain_id}:${p.src_channel}`
          const resolved = resolvedChannels.value.get(resolvedKey)
          if (resolved) {
            destChain = resolved.counterpartyChainId
          }
        }
        
        return {
          from_chain: p.chain_id,
          to_chain: destChain || 'Unknown',
          channel: p.src_channel || p.dst_channel || 'unknown',
          status: p.effected ? 'success' : 'pending',
          timestamp: p.timestamp
        }
      })
    }
  }
  if (monitoringData.value) {
    if (!topRelayers.value.length && monitoringData.value.top_relayers) {
      topRelayers.value = monitoringData.value.top_relayers
    }
    if (!recentActivity.value.length && monitoringData.value.recent_activity) {
      recentActivity.value = monitoringData.value.recent_activity
    }
  }
}

watchEffect(async () => {
  await updateStats()
})

function viewPacketDetails(packet: any) {
  router.push({
    name: 'packet-clearing',
    query: {
      packet: JSON.stringify({
        id: `${packet.chain_id}-${packet.sequence}`,
        channelId: packet.src_channel,
        sequence: packet.sequence,
        sourceChain: packet.chain_id,
        amount: packet.amount,
        denom: packet.denom,
        sender: packet.sender,
        receiver: packet.receiver
      })
    }
  })
}

const chainNames = ref<Record<string, string>>({})

onMounted(async () => {
  const chains = await configService.getAllChains()
  chains.forEach(chain => {
    chainNames.value[chain.chain_id] = chain.chain_name
  })
})

function getChainNameSync(chainId: string): string {
  if (chainNames.value[chainId]) {
    return chainNames.value[chainId]
  }
  
  const commonNames: Record<string, string> = {
    'cosmoshub-4': 'Cosmos Hub',
    'osmosis-1': 'Osmosis',
    'neutron-1': 'Neutron',
    'noble-1': 'Noble',
    'axelar-dojo-1': 'Axelar',
    'stride-1': 'Stride',
    'dydx-mainnet-1': 'dYdX',
    'celestia': 'Celestia',
    'injective-1': 'Injective',
    'kava_2222-10': 'Kava',
    'secret-4': 'Secret',
    'stargaze-1': 'Stargaze'
  }
  
  return commonNames[chainId] || chainId
}

function extractTokenFromDenom(denom: string): string {
  if (denom.includes('/')) {
    const parts = denom.split('/')
    const token = parts[parts.length - 1]
    if (token === 'uusdc') return 'USDC'
    if (token === 'uatom') return 'ATOM'
    if (token === 'uosmo') return 'OSMO'
    if (token === 'ustrd') return 'STRD'
    if (token === 'utia') return 'TIA'
    if (token === 'inj') return 'INJ'
    if (token.startsWith('u')) return token.substring(1).toUpperCase()
    return token.toUpperCase()
  }
  if (denom === 'uusdc') return 'USDC'
  if (denom === 'uatom') return 'ATOM'
  if (denom === 'uosmo') return 'OSMO'
  if (denom.startsWith('u')) return denom.substring(1).toUpperCase()
  return denom.toUpperCase()
}

function getDestChainFromChannel(dstChannel: string): string {
  if (dstChannel === 'channel-0') return 'cosmoshub-4'
  if (dstChannel === 'channel-1') return 'noble-1'
  if (dstChannel === 'channel-141') return 'cosmoshub-4'
  if (dstChannel === 'channel-208') return 'axelar-dojo-1'
  return 'Unknown'
}
</script>