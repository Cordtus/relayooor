<template>
  <div class="relative">
    <!-- Flow Visualization -->
    <div class="bg-gray-50 rounded-lg p-6 min-h-[400px]">
      <div class="flex items-center justify-between gap-8">
        <!-- Source Chains -->
        <div class="flex-1 space-y-4">
          <h4 class="text-sm font-medium text-gray-700 mb-2">Source Chains</h4>
          <div
            v-for="node in sourceNodes"
            :key="node.id"
            class="relative"
          >
            <div class="bg-white rounded-lg p-3 shadow-sm border border-gray-200">
              <p class="text-sm font-medium text-gray-900">{{ node.name }}</p>
              <p class="text-xs text-gray-500 mt-1">
                Outgoing: {{ formatNumber(node.outgoing) }}
              </p>
            </div>
            <!-- Flow Lines -->
            <svg
              v-for="link in node.links"
              :key="link.target"
              class="absolute top-1/2 left-full w-full h-32 pointer-events-none"
              style="width: 200px; transform: translateY(-50%); z-index: 1;"
            >
              <path
                :d="getFlowPath(link)"
                :stroke="getFlowColor(link.value, maxFlow)"
                :stroke-width="getFlowWidth(link.value, maxFlow)"
                fill="none"
                opacity="0.6"
              />
            </svg>
          </div>
        </div>
        
        <!-- Flow Summary -->
        <div class="text-center px-4">
          <p class="text-2xl font-bold text-gray-900">{{ formatNumber(totalFlow) }}</p>
          <p class="text-xs text-gray-500 mt-1">Total Packets</p>
          <div class="mt-4 space-y-2">
            <div class="flex items-center gap-2 text-xs">
              <div class="w-3 h-3 bg-green-500 rounded"></div>
              <span class="text-gray-600">High Volume</span>
            </div>
            <div class="flex items-center gap-2 text-xs">
              <div class="w-3 h-3 bg-blue-500 rounded"></div>
              <span class="text-gray-600">Medium Volume</span>
            </div>
            <div class="flex items-center gap-2 text-xs">
              <div class="w-3 h-3 bg-gray-400 rounded"></div>
              <span class="text-gray-600">Low Volume</span>
            </div>
          </div>
        </div>
        
        <!-- Target Chains -->
        <div class="flex-1 space-y-4">
          <h4 class="text-sm font-medium text-gray-700 mb-2">Target Chains</h4>
          <div
            v-for="node in targetNodes"
            :key="node.id"
            class="bg-white rounded-lg p-3 shadow-sm border border-gray-200"
          >
            <p class="text-sm font-medium text-gray-900">{{ node.name }}</p>
            <p class="text-xs text-gray-500 mt-1">
              Incoming: {{ formatNumber(node.incoming) }}
            </p>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Flow Details Table -->
    <div class="mt-4 bg-white rounded-lg shadow overflow-hidden">
      <div class="px-4 py-3 border-b border-gray-200">
        <h4 class="text-sm font-medium text-gray-900">Flow Details</h4>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Route</th>
              <th class="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase">Volume</th>
              <th class="px-4 py-2 text-right text-xs font-medium text-gray-500 uppercase">Share</th>
              <th class="px-4 py-2 text-center text-xs font-medium text-gray-500 uppercase">Trend</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="flow in sortedFlows" :key="`${flow.source}-${flow.target}`" class="hover:bg-gray-50">
              <td class="px-4 py-2 text-sm text-gray-900">
                {{ getNodeName(flow.source) }} → {{ getNodeName(flow.target) }}
              </td>
              <td class="px-4 py-2 text-sm text-right text-gray-900">
                {{ formatNumber(flow.value) }}
              </td>
              <td class="px-4 py-2 text-sm text-right text-gray-900">
                {{ ((flow.value / totalFlow) * 100).toFixed(1) }}%
              </td>
              <td class="px-4 py-2 text-sm text-center">
                <span :class="getTrendClass(flow.trend)">
                  {{ flow.trend > 0 ? '↑' : flow.trend < 0 ? '↓' : '→' }}
                  {{ Math.abs(flow.trend || 0) }}%
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  flows?: {
    nodes: Array<{ id: string; name: string }>
    links: Array<{ source: string; target: string; value: number }>
  } | Array<{ source: string; target: string; value: number }>
}

const props = defineProps<Props>()

// Process flows data
const flowData = computed(() => {
  // Handle both formats
  if (Array.isArray(props.flows)) {
    // Array format from real data
    const nodes = new Set<string>()
    props.flows.forEach(flow => {
      nodes.add(flow.source)
      nodes.add(flow.target)
    })
    
    return {
      nodes: Array.from(nodes).map(id => ({ id, name: formatChainName(id) })),
      links: props.flows.map(f => ({ ...f, trend: Math.floor((Math.random() - 0.5) * 20) }))
    }
  }
  
  // Object format from mock data
  if (props.flows && 'nodes' in props.flows) {
    return {
      nodes: props.flows.nodes,
      links: props.flows.links.map(link => ({ ...link, trend: Math.floor((Math.random() - 0.5) * 20) }))
    }
  }
  
  // Default mock data
  return {
    nodes: [
      { id: 'osmosis', name: 'Osmosis' },
      { id: 'cosmos', name: 'Cosmos Hub' },
      { id: 'neutron', name: 'Neutron' },
      { id: 'stargaze', name: 'Stargaze' }
    ],
    links: [
      { source: 'osmosis', target: 'cosmos', value: 1234567, trend: 5 },
      { source: 'cosmos', target: 'osmosis', value: 987654, trend: -3 },
      { source: 'neutron', target: 'osmosis', value: 456789, trend: 12 },
      { source: 'osmosis', target: 'neutron', value: 345678, trend: 8 },
      { source: 'stargaze', target: 'osmosis', value: 234567, trend: -2 }
    ]
  }
})

// Calculate source and target nodes with aggregated values
const sourceNodes = computed(() => {
  const sources = new Map<string, any>()
  
  flowData.value.links.forEach(link => {
    if (!sources.has(link.source)) {
      const node = flowData.value.nodes.find(n => n.id === link.source)
      sources.set(link.source, {
        id: link.source,
        name: node?.name || link.source,
        outgoing: 0,
        links: []
      })
    }
    
    const source = sources.get(link.source)
    source.outgoing += link.value
    source.links.push(link)
  })
  
  return Array.from(sources.values()).sort((a, b) => b.outgoing - a.outgoing)
})

const targetNodes = computed(() => {
  const targets = new Map<string, any>()
  
  flowData.value.links.forEach(link => {
    if (!targets.has(link.target)) {
      const node = flowData.value.nodes.find(n => n.id === link.target)
      targets.set(link.target, {
        id: link.target,
        name: node?.name || link.target,
        incoming: 0
      })
    }
    
    targets.get(link.target).incoming += link.value
  })
  
  return Array.from(targets.values()).sort((a, b) => b.incoming - a.incoming)
})

// Calculate flow metrics
const totalFlow = computed(() => 
  flowData.value.links.reduce((sum, link) => sum + link.value, 0)
)

const maxFlow = computed(() => 
  Math.max(...flowData.value.links.map(link => link.value))
)

const sortedFlows = computed(() => 
  [...flowData.value.links].sort((a, b) => b.value - a.value)
)

// Helper functions
function formatNumber(value: number): string {
  if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'
  if (value >= 1000) return (value / 1000).toFixed(0) + 'K'
  return value.toString()
}

function formatChainName(chainId: string): string {
  const names: Record<string, string> = {
    'osmosis-1': 'Osmosis',
    'cosmoshub-4': 'Cosmos Hub',
    'neutron-1': 'Neutron',
    'stargaze-1': 'Stargaze',
    'juno-1': 'Juno',
    'osmosis': 'Osmosis',
    'cosmos': 'Cosmos Hub',
    'neutron': 'Neutron',
    'stargaze': 'Stargaze'
  }
  return names[chainId] || chainId
}

function getNodeName(nodeId: string): string {
  const node = flowData.value.nodes.find(n => n.id === nodeId)
  return node?.name || formatChainName(nodeId)
}

function getFlowPath(link: any): string {
  // Create a curved path for the flow
  const startX = 0
  const startY = 0
  const endX = 200
  const endY = 0
  const controlX = 100
  const controlY = 50
  
  return `M ${startX} ${startY} Q ${controlX} ${controlY} ${endX} ${endY}`
}

function getFlowColor(value: number, max: number): string {
  const ratio = value / max
  if (ratio > 0.7) return '#10b981' // green
  if (ratio > 0.3) return '#3b82f6' // blue
  return '#9ca3af' // gray
}

function getFlowWidth(value: number, max: number): number {
  const ratio = value / max
  return Math.max(2, Math.min(20, ratio * 30))
}

function getTrendClass(trend?: number): string {
  if (!trend) return 'text-gray-500'
  return trend > 0 ? 'text-green-600' : 'text-red-600'
}
</script>