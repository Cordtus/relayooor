import axios from 'axios'
import type { MetricsSnapshot } from '@/types/monitoring'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
})

export const metricsService = {
  // Fetch raw Prometheus metrics from Chainpulse
  async getRawMetrics(): Promise<string> {
    const response = await api.get('/api/metrics/chainpulse', {
      responseType: 'text'
    })
    return response.data
  },

  // Fetch structured monitoring data
  async getMonitoringData(): Promise<any> {
    const response = await api.get('/api/monitoring/data')
    return response.data
  },

  // Parse Prometheus metrics into structured format
  parsePrometheusMetrics(rawMetrics: string): MetricsSnapshot {
    const lines = rawMetrics.split('\n')
    const metrics: any = {
      system: {
        totalChains: 0,
        totalTransactions: 0,
        totalPackets: 0,
        totalErrors: 0,
        uptime: 0,
        lastSync: new Date()
      },
      chains: [],
      relayers: new Map(),
      channels: new Map(),
      recentPackets: [],
      stuckPackets: [],
      frontrunEvents: [],
      timestamp: new Date()
    }

    // Parse each line
    lines.forEach(line => {
      if (line.startsWith('#') || !line.trim()) return

      const match = line.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)(?:{([^}]+)})?\s+([+-]?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/)
      if (!match) return

      const [, metricName, labelsStr, value] = match
      const labels = this.parseLabels(labelsStr || '')
      const numValue = parseFloat(value)

      // Process different metric types
      switch (metricName) {
        case 'chainpulse_chains':
          metrics.system.totalChains = numValue
          break
        
        case 'chainpulse_txs':
          metrics.system.totalTransactions += numValue
          this.updateChainMetrics(metrics.chains, labels.chain_id, 'totalTxs', numValue)
          break
        
        case 'chainpulse_packets':
          metrics.system.totalPackets += numValue
          this.updateChainMetrics(metrics.chains, labels.chain_id, 'totalPackets', numValue)
          break
        
        case 'chainpulse_errors':
          metrics.system.totalErrors += numValue
          this.updateChainMetrics(metrics.chains, labels.chain_id, 'errors', numValue)
          break
        
        case 'chainpulse_reconnects':
          this.updateChainMetrics(metrics.chains, labels.chain_id, 'reconnects', numValue)
          break
        
        case 'ibc_effected_packets':
          this.processPacketMetric(metrics, labels, numValue, true)
          break
        
        case 'ibc_uneffected_packets':
          this.processPacketMetric(metrics, labels, numValue, false)
          break
        
        case 'ibc_frontrun_counter':
          this.processFrontrunMetric(metrics, labels, numValue)
          break
        
        case 'ibc_stuck_packets':
          if (numValue > 0) {
            metrics.stuckPackets.push({
              srcChain: labels.src_chain,
              dstChain: labels.dst_chain,
              srcChannel: labels.src_channel,
              sequence: numValue,
              stuckDuration: 0 // Would need additional data
            })
          }
          break
      }
    })

    // Convert maps to arrays
    metrics.relayers = Array.from(metrics.relayers.values())
      .sort((a, b) => b.effectedPackets - a.effectedPackets)
    metrics.channels = Array.from(metrics.channels.values())
      .sort((a, b) => b.totalPackets - a.totalPackets)

    return metrics
  },

  parseLabels(labelsStr: string): Record<string, string> {
    const labels: Record<string, string> = {}
    if (!labelsStr) return labels

    const regex = /(\w+)="([^"]+)"/g
    let match
    while ((match = regex.exec(labelsStr)) !== null) {
      labels[match[1]] = match[2]
    }
    return labels
  },

  updateChainMetrics(chains: any[], chainId: string, field: string, value: number) {
    let chain = chains.find(c => c.chainId === chainId)
    if (!chain) {
      chain = {
        chainId,
        chainName: this.getChainName(chainId),
        totalTxs: 0,
        totalPackets: 0,
        reconnects: 0,
        timeouts: 0,
        errors: 0,
        status: 'connected',
        lastUpdate: new Date()
      }
      chains.push(chain)
    }
    chain[field] = value
  },

  processPacketMetric(metrics: any, labels: any, value: number, effected: boolean) {
    // Update relayer metrics
    const relayerKey = labels.signer
    if (!metrics.relayers.has(relayerKey)) {
      metrics.relayers.set(relayerKey, {
        address: labels.signer,
        totalPackets: 0,
        effectedPackets: 0,
        uneffectedPackets: 0,
        frontrunCount: 0,
        successRate: 0,
        memo: labels.memo || '',
        software: this.extractSoftware(labels.memo),
        version: this.extractVersion(labels.memo)
      })
    }
    
    const relayer = metrics.relayers.get(relayerKey)
    relayer.totalPackets += value
    if (effected) {
      relayer.effectedPackets += value
    } else {
      relayer.uneffectedPackets += value
    }
    relayer.successRate = (relayer.effectedPackets / relayer.totalPackets) * 100

    // Update channel metrics
    const channelKey = `${labels.chain_id}-${labels.src_channel}-${labels.dst_channel}`
    if (!metrics.channels.has(channelKey)) {
      metrics.channels.set(channelKey, {
        srcChain: labels.chain_id,
        dstChain: this.inferDestChain(labels.chain_id, labels.dst_channel),
        srcChannel: labels.src_channel,
        dstChannel: labels.dst_channel,
        srcPort: labels.src_port,
        dstPort: labels.dst_port,
        totalPackets: 0,
        effectedPackets: 0,
        uneffectedPackets: 0,
        successRate: 0,
        status: 'active'
      })
    }
    
    const channel = metrics.channels.get(channelKey)
    channel.totalPackets += value
    if (effected) {
      channel.effectedPackets += value
    } else {
      channel.uneffectedPackets += value
    }
    channel.successRate = (channel.effectedPackets / channel.totalPackets) * 100
  },

  processFrontrunMetric(metrics: any, labels: any, value: number) {
    metrics.frontrunEvents.push({
      chain_id: labels.chain_id,
      channel: labels.src_channel,
      signer: labels.signer,
      frontrunned_by: labels.frontrunned_by,
      count: value,
      timestamp: new Date()
    })
  },

  getChainName(chainId: string): string {
    const names: Record<string, string> = {
      'cosmoshub-4': 'Cosmos Hub',
      'osmosis-1': 'Osmosis',
      'neutron-1': 'Neutron'
    }
    return names[chainId] || chainId
  },

  inferDestChain(srcChain: string, dstChannel: string): string {
    // Known channel mappings
    if (srcChain === 'osmosis-1' && dstChannel === 'channel-0') return 'cosmoshub-4'
    if (srcChain === 'cosmoshub-4' && dstChannel === 'channel-141') return 'osmosis-1'
    if (srcChain === 'neutron-1' && dstChannel === 'channel-10') return 'osmosis-1'
    return 'unknown'
  },

  extractSoftware(memo: string): string {
    if (memo.includes('hermes')) return 'Hermes'
    if (memo.includes('rly')) return 'Go Relayer'
    return 'Unknown'
  },

  extractVersion(memo: string): string {
    const match = memo.match(/(\d+\.\d+\.\d+)/)?.[1]
    return match || 'Unknown'
  }
}